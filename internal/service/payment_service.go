package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Emerald211/healthconnect/internal/config"
	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/Emerald211/healthconnect/internal/repository"
)

type PaymentService struct {
	paymentRepo     *repository.PaymentRepository
	appointmentRepo *repository.AppointmentRepository
	patientRepo     *repository.PatientRepository
	cfg             *config.Config
	emailService    *EmailService
}

func NewPaymentService(paymentRepo *repository.PaymentRepository, appointmentRepo *repository.AppointmentRepository, patientRepo *repository.PatientRepository, cfg *config.Config, emailService *EmailService) *PaymentService {
	return &PaymentService{
		paymentRepo:     paymentRepo,
		appointmentRepo: appointmentRepo,
		patientRepo:     patientRepo,
		cfg:             cfg,
		emailService:    emailService,
	}
}

func (s *PaymentService) InitializePayment(ctx context.Context, patientId string, req domain.InitializePaymentRequest) (domain.InitializePaymentResponse, error) {

	appointment, err := s.appointmentRepo.GetByID(ctx, req.AppointmentID)
	if err != nil {
		return domain.InitializePaymentResponse{}, err
	}

	if appointment.PatientID != patientId {
		slog.Warn("unauthorized payment attempt",
			"patient_id", patientId,
			"appointment_id", req.AppointmentID,
		)
		return domain.InitializePaymentResponse{}, domain.ErrNotYourAppointment
	}

	existingPayment, err := s.paymentRepo.GetByAppointmentID(ctx, appointment.ID)
	if err != nil {
		if existingPayment.Status == "successful" {
			return domain.InitializePaymentResponse{}, domain.ErrAlreadyPaid
		}

		if existingPayment.Status == "pending" && existingPayment.ExpiresAt != nil {
			if time.Now().Before(*existingPayment.ExpiresAt) && existingPayment.PaystackAccessCode != nil {
				slog.Info("returning existing payment link",
					"appointment_id", appointment.ID,
					"expires_at", existingPayment.ExpiresAt,
				)
				return domain.InitializePaymentResponse{
					AuthorizationURL: "https://checkout.paystack.com/" + *existingPayment.PaystackAccessCode,
					Reference:        *existingPayment.PaystackReference,
					Payment:          existingPayment,
				}, nil
			}
		}
	}

	patient, _, err := s.patientRepo.FindByID(ctx, patientId)
	if err != nil {
		return domain.InitializePaymentResponse{}, fmt.Errorf("failed to find patient: %w", err)
	}

	payment, err := s.paymentRepo.GetByAppointmentID(ctx, appointment.ID)
	if err == nil && payment.Status == "successful" {
		return domain.InitializePaymentResponse{}, fmt.Errorf("payment already completed for this appointment")
	}

	if err != nil {
		payment, err = s.paymentRepo.CreatePayment(ctx, appointment.ID, patientId, appointment.Amount)
		if err != nil {
			return domain.InitializePaymentResponse{}, fmt.Errorf("failed to create payment: %w", err)
		}

		slog.Info("payment record created",
			"payment_id", payment.ID,
			"appointment_id", appointment.ID,
			"amount", appointment.Amount,
		)
	}

	amountInKobo := int(appointment.Amount * 100)

	paystackBody := map[string]interface{}{
		"email":  patient.Email,
		"amount": amountInKobo,
		"metadata": map[string]interface{}{
			"appointment_id": appointment.ID,
			"payment_id":     payment.ID,
			"patient_name":   patient.Name,
		},
	}

	bodyBytes, err := json.Marshal(paystackBody)
	if err != nil {
		return domain.InitializePaymentResponse{}, fmt.Errorf("preparing paystack request: %w", err)
	}

	// http post request to paystack

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.cfg.PaystackBaseURL+"/transaction/initialize", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return domain.InitializePaymentResponse{}, fmt.Errorf("error preparing paystack request: %w", err)
	}

	// set headers
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.PaystackSecretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// send request to paystack

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		slog.Error("paystack API call failed",
			"appointment_id", appointment.ID,
			"error", err,
		)
		return domain.InitializePaymentResponse{}, fmt.Errorf("error sending paystack request: %w", err)
	}
	defer resp.Body.Close()

	// read paystack response
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.InitializePaymentResponse{}, fmt.Errorf("error reading paystack response: %w", err)
	}

	// parse paystack response
	var paystackResp struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
			AccessCode       string `json:"access_code"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBytes, &paystackResp); err != nil {
		return domain.InitializePaymentResponse{}, domain.ErrPaymentFailed
	}

	// check if paystack returned an error
	if !paystackResp.Status {
		slog.Error("paystack returned error",
			"message", paystackResp.Message,
			"appointment_id", appointment.ID,
		)
		return domain.InitializePaymentResponse{}, fmt.Errorf("paystack response status: %s", paystackResp.Message)
	}

	// save reference to record so the webhook can verify the payment
	if err := s.paymentRepo.UpdatePaymentDetails(ctx, payment.ID, paystackResp.Data.Reference, paystackResp.Data.AccessCode, "pending"); err != nil {
		return domain.InitializePaymentResponse{}, fmt.Errorf("saving paystack details: %w", err)
	}

	slog.Info("payment initialized",
		"reference", paystackResp.Data.Reference,
		"appointment_id", appointment.ID,
		"patient_id", patientId,
		"amount_naira", appointment.Amount,
	)

	// get updated payment reecord
	payment, err = s.paymentRepo.GetByAppointmentID(ctx, appointment.ID)
	if err != nil {
		return domain.InitializePaymentResponse{}, fmt.Errorf("getting updated payment: %w", err)
	}

	return domain.InitializePaymentResponse{
		AuthorizationURL: paystackResp.Data.AuthorizationURL,
		Reference:        paystackResp.Data.Reference,
		Payment:          payment,
	}, nil

}

func (s *PaymentService) HandleWebhook(ctx context.Context, body []byte, signature string) error {
	// verify webhook
	if !s.verifySignature(body, signature) {
		slog.Warn("invalid webhook signature received")
		return domain.ErrInvalidWebhook
	}

	// parse event body
	var event domain.PaystackWebhookEvent

	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("parsing event body: %w", err)
	}

	slog.Info("webhook received", "event", event.Event)

	if event.Event != "charge.success" {
		slog.Info("ignoring webhook event", "event", event.Event)
		return nil
	}

	reference, ok := event.Data["reference"].(string)
	if !ok || reference == "" {
		return fmt.Errorf("missing or invalid reference in webhook")
	}

	// Verify with Paystack directly
	if err := s.verifyWithPaystack(ctx, reference); err != nil {
		slog.Error("paystack verification failed",
			"reference", reference,
			"error", err,
		)
		return err
	}

	// confirm the payment from db with the refeerence which was originally saved during initialize payment and so if status is pending is when it would be updated to success
	if err := s.paymentRepo.ConfirmPayment(ctx, reference); err != nil {
		slog.Error("failed to confirm payment",
			"reference", reference,
			"error", err,
		)
		return fmt.Errorf("confirming payment: %w", err)
	}

	slog.Info("payment confirmed", "reference", reference)
	return nil
}

func (s *PaymentService) verifySignature(body []byte, signature string) bool {
	mac := hmac.New(sha512.New, []byte(s.cfg.PaystackSecretKey))

	mac.Write(body)

	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

func (s *PaymentService) GetPaymentStatus(ctx context.Context, patientID, appointmentID string) (domain.Payment, error) {
	appointment, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return domain.Payment{}, err
	}

	if appointment.PatientID != patientID {
		return domain.Payment{}, fmt.Errorf("appointment does not belong to patient")
	}

	payment, err := s.paymentRepo.GetByAppointmentID(ctx, appointmentID)
	if err != nil {
		return domain.Payment{}, err
	}

	return payment, nil
}

func (s *PaymentService) verifyWithPaystack(ctx context.Context, reference string) error {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"GET",
		s.cfg.PaystackBaseURL+"/transaction/verify/"+reference,
		nil,
	)
	if err != nil {
		return fmt.Errorf("building verify request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.PaystackSecretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("calling paystack verify: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading verify response: %w", err)
	}

	// Log the raw response so we can see exactly what Paystack returns
	slog.Info("paystack verify raw response", "body", string(respBytes))

	var paystackResp struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status    string `json:"status"` // "success", "failed", "abandoned"
			Reference string `json:"reference"`
			Amount    int    `json:"amount"`
			Currency  string `json:"currency"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBytes, &paystackResp); err != nil {
		return fmt.Errorf("parsing verify response: %w", err)
	}

	slog.Info("paystack verify parsed",
		"status", paystackResp.Status,
		"data_status", paystackResp.Data.Status,
		"reference", reference,
	)

	if !paystackResp.Status {
		return fmt.Errorf("paystack verify returned false: %s", paystackResp.Message)
	}

	if paystackResp.Data.Status != "success" {
		slog.Warn("paystack verification failed",
			"reference", reference,
			"status", paystackResp.Data.Status,
		)
		return domain.ErrPaymentFailed
	}

	slog.Info("paystack verification successful", "reference", reference)
	return nil
}

// RefundPayment refunds a payment via Paystack
func (s *PaymentService) RefundPayment(ctx context.Context, appointmentID, doctorID string) error {
	// Get appointment
	appointment, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return err
	}

	if appointment.DoctorID != doctorID {
		return domain.ErrNotYourAppointment
	}

	// Get payment
	payment, err := s.paymentRepo.GetByAppointmentID(ctx, appointmentID)
	if err != nil {
		return fmt.Errorf("getting payment: %w", err)
	}

	if payment.Status != "successful" {
		return domain.ErrPaymentNotFound
	}

	// Call Paystack refund API
	refundBody := map[string]interface{}{
		"transaction": *payment.PaystackReference,
	}

	bodyBytes, err := json.Marshal(refundBody)
	if err != nil {
		return fmt.Errorf("preparing refund request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		s.cfg.PaystackBaseURL+"/refund",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return fmt.Errorf("building refund request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.PaystackSecretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("calling paystack refund: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading refund response: %w", err)
	}

	var paystackResp struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(respBytes, &paystackResp); err != nil {
		return fmt.Errorf("parsing refund response: %w", err)
	}

	if !paystackResp.Status {
		slog.Error("paystack refund failed",
			"appointment_id", appointmentID,
			"message", paystackResp.Message,
		)
		return domain.ErrPaymentFailed
	}

	slog.Info("payment refunded",
		"appointment_id", appointmentID,
		"reference", *payment.PaystackReference,
	)

	return nil
}
