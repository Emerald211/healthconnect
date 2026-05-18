package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PatientRepository struct {
	db *pgxpool.Pool
}

func NewPatientRepository(db *pgxpool.Pool) *PatientRepository {
	return &PatientRepository{db: db}
}

// For Checking Email exist from Database
func (r *PatientRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM patients WHERE email = $1)`
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking email exists: %w", err)
	}

	return exists, nil
}

func (r *PatientRepository) Create(ctx context.Context, p domain.Patient, hashedPassword string) (domain.Patient, error) {
	var patient domain.Patient
	query := `
		INSERT INTO patients (name, email, phone, password, date_of_birth, gender, address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, email, phone, date_of_birth, gender, address, is_active, is_email_verified, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		p.Name,
		p.Email,
		p.Phone,
		hashedPassword,
		p.DateOfBirth,
		p.Gender,
		p.Address,
	).Scan(
		&patient.ID,
		&patient.Name,
		&patient.Email,
		&patient.Phone,
		&patient.DateOfBirth,
		&patient.Gender,
		&patient.Address,
		&patient.IsActive,
		&patient.IsEmailVerified,
		&patient.CreatedAt,
		&patient.UpdatedAt,
	)
	if err != nil {
		return domain.Patient{}, fmt.Errorf("creating patient: %w", err)
	}
	return patient, nil
}

func (r *PatientRepository) FindByEmail(ctx context.Context, email string) (domain.Patient, string, error) {
	query := `
		SELECT id, name, email, phone, password, date_of_birth, gender, address, is_active, is_email_verified, created_at, updated_at
		FROM patients
		WHERE email = $1
	`

	var patient domain.Patient
	var hashedPassword string

	err := r.db.QueryRow(ctx, query, email).Scan(
		&patient.ID,
		&patient.Name,
		&patient.Email,
		&patient.Phone,
		&hashedPassword,
		&patient.DateOfBirth,
		&patient.Gender,
		&patient.Address,
		&patient.IsActive,
		&patient.IsEmailVerified,
		&patient.CreatedAt,
		&patient.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Patient{}, "", domain.ErrPatientNotFound
		}
		return domain.Patient{}, "", fmt.Errorf("finding patient by email: %w", err)
	}

	return patient, hashedPassword, nil
}

func (r *PatientRepository) FindByID(ctx context.Context, id string) (domain.Patient, string, error) {

	var patient domain.Patient

	query := `
			SELECT id, name, email, phone, password ,date_of_birth, gender, address, is_active, is_email_verified, created_at, updated_at
			FROM patients
			WHERE id = $1
		`

	var hashedPassword string

	err := r.db.QueryRow(ctx, query, id).Scan(&patient.ID,
		&patient.Name,
		&patient.Email,
		&patient.Phone,
		&hashedPassword,
		&patient.DateOfBirth,
		&patient.Gender,
		&patient.Address,
		&patient.IsActive,
		&patient.IsEmailVerified,
		&patient.CreatedAt,
		&patient.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Patient{}, "", domain.ErrPatientNotFound
		}

		return domain.Patient{}, "", fmt.Errorf("finding patient by id: %w", err)
	}

	return patient, hashedPassword, nil
}

func (r *PatientRepository) MarkEmailVerified(ctx context.Context, email string) error {
	query := `UPDATE patients SET is_email_verified = true, updated_at = NOW() WHERE email = $1`
	_, err := r.db.Exec(ctx, query, email)
	if err != nil {
		return fmt.Errorf("marking email verified: %w", err)
	}
	return nil
}

func (r *PatientRepository) UpdatePassword(ctx context.Context, email, hashedPassword string) error {
	query := `UPDATE patients SET password = $1, updated_at = NOW() WHERE email = $2`
	_, err := r.db.Exec(ctx, query, hashedPassword, email)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}
	return nil
}

func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
