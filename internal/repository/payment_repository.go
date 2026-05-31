package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) CreatePayment(ctx context.Context, appointmentID, patientID string, amount float64) (domain.Payment, error) {
	var p domain.Payment

	query := `
			INSERT INTO payments (appointment_id, patient_id, amount)
			VALUES ($1, $2, $3)
			RETURNING id, appointment_id, patient_id, amount, currency, status,
			          paystack_reference, paystack_access_code, paid_at, expires_at, created_at, updated_at
		`

	err := r.db.QueryRow(ctx, query, appointmentID, patientID, amount).Scan(
		&p.ID,
		&p.AppointmentID,
		&p.PatientID,
		&p.Amount,
		&p.Currency,
		&p.Status,
		&p.PaystackReference,
		&p.PaystackAccessCode,
		&p.PaidAt,
		&p.ExpiresAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		return domain.Payment{}, fmt.Errorf("creating payment: %w", err)
	}

	return p, nil
}

func (r *PaymentRepository) UpdatePaymentDetails(ctx context.Context, paymentID, reference, accessCode, status string) error {
	query := `
		UPDATE payments
		SET paystack_reference = $1, paystack_access_code = $2, status = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, reference, accessCode, status, paymentID)
	if err != nil {
		return fmt.Errorf("updating payment details: %w", err)
	}
	return nil
}

func (r *PaymentRepository) ConfirmPayment(ctx context.Context, reference string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("confirming payment: %w", err)
	}
	defer tx.Rollback(ctx)

	var appointmentID string

	query := `
			UPDATE payments
			SET status = 'successful', paid_at = NOW(), updated_at = NOW()
			WHERE paystack_reference = $1 AND status = 'pending'
			RETURNING appointment_id
		`

	err = tx.QueryRow(ctx, query, reference).Scan(&appointmentID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("confirming payment: %w", err)
	}

	_, err = tx.Exec(ctx, `
			UPDATE appointments SET status = 'confirmed', updated_at = NOW()
			WHERE id = $1
		`, appointmentID)

	if err != nil {
		return fmt.Errorf("confirming appointment: %w", err)
	}

	return tx.Commit(ctx)

}

func (r *PaymentRepository) GetByAppointmentID(ctx context.Context, appointmentID string) (domain.Payment, error) {
	query := `
			SELECT id, appointment_id, patient_id, amount, currency, status,
			       paystack_reference, paystack_access_code, paid_at, expires_at, created_at, updated_at
			FROM payments WHERE appointment_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`

	var payment domain.Payment

	err := r.db.QueryRow(ctx, query, appointmentID).Scan(
		&payment.ID,
		&payment.AppointmentID,
		&payment.PatientID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaystackReference,
		&payment.PaystackAccessCode,
		&payment.PaidAt,
		&payment.ExpiresAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		return domain.Payment{}, fmt.Errorf("getting payment by appointment id: %w", err)
	}

	return payment, nil
}
