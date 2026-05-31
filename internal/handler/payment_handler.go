package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/Emerald211/healthconnect/internal/service"
	"github.com/Emerald211/healthconnect/pkg/response"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	service *service.PaymentService
}

func NewPaymentHandler(service *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

// InitializePayment godoc
// @Summary      Initialize payment for an appointment
// @Description  Creates a Paystack payment link for the patient to pay
// @Tags         payments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      domain.InitializePaymentRequest  true  "Payment data"
// @Success      200      {object}  domain.InitializePaymentResponse
// @Failure      400      {object}  map[string]interface{}
// @Failure      403      {object}  map[string]interface{}
// @Router       /api/v1/payments/initialize [post]
func (h *PaymentHandler) InitiliazePayment(c *gin.Context) {
	var req domain.InitializePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_eerror",
			"message": err.Error(),
		})

		return
	}

	patientID, _ := c.Get("user_id")

	result, err := h.service.InitializePayment(c.Request.Context(), patientID.(string), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

// HandleWebhook godoc
// @Summary      Paystack webhook
// @Description  Receives and processes payment notifications from Paystack
// @Tags         payments
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/payments/webhook [post]
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {

	body, err := io.ReadAll(c.Request.Body)

	if err != nil {
		slog.Error("failed to read webhook body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "bad_request",
			"message": "could not read request body",
		})

		return
	}

	signature := c.GetHeader("X-Paystack-Signature")

	if signature == "" {
		slog.Warn("webhook received without signature")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "missing_signature",
			"message": "missing paystack signature",
		})
		return
	}

	if err := h.service.HandleWebhook(c.Request.Context(), body, signature); err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "webhook received",
	})

}

// GetPaymentStatus godoc
// @Summary      Get payment status
// @Description  Returns the payment status for an appointment
// @Tags         payments
// @Produce      json
// @Security     BearerAuth
// @Param        appointment_id  path  string  true  "Appointment ID"
// @Success      200  {object}  domain.Payment
// @Failure      403  {object}  map[string]interface{}
// @Router       /api/v1/payments/{appointment_id} [get]
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	appointmentID := c.Param("appointment_id")
	patientID, _ := c.Get("user_id")

	payment, err := h.service.GetPaymentStatus(c.Request.Context(), patientID.(string), appointmentID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, payment)
}
