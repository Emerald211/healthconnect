package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DoctorRepository struct {
	db *pgxpool.Pool
}

func NewDoctorRepository(db *pgxpool.Pool) *DoctorRepository {
	return &DoctorRepository{db: db}
}

func (r *DoctorRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := "SELECT EXISTS(SELECT 1 FROM doctors WHERE email = $1)"

	err := r.db.QueryRow(ctx, query, email).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("checking doctor email exists: %w", err)
	}

	return exists, nil
}

func (r *DoctorRepository) LicenseExists(ctx context.Context, license string) (bool, error) {
	var exists bool

	query := "SELECT EXISTS(SELECT 1 FROM doctors WHERE license_number = $1)"

	err := r.db.QueryRow(ctx, query, license).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("checking license exists: %w", err)
	}

	return exists, nil
}

func (r *DoctorRepository) Create(ctx context.Context, d domain.Doctor, hashedPassword string) (domain.Doctor, error) {
	var doctor domain.Doctor
	query := `
			INSERT INTO doctors (name, email, phone, password, specialty, license_number, years_experience, consultation_fee, bio)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, name, email, phone, specialty, license_number, years_experience, consultation_fee, bio, is_active, is_verified, created_at, updated_at
		`

	err := r.db.QueryRow(ctx, query,
		d.Name,
		d.Email,
		d.Phone,
		hashedPassword,
		d.Specialty,
		d.LicenseNumber,
		d.YearsExperience,
		d.ConsultationFee,
		d.Bio,
	).Scan(
		&doctor.ID,
		&doctor.Name,
		&doctor.Email,
		&doctor.Phone,
		&doctor.Specialty,
		&doctor.LicenseNumber,
		&doctor.YearsExperience,
		&doctor.ConsultationFee,
		&doctor.Bio,
		&doctor.IsActive,
		&doctor.IsVerified,
		&doctor.CreatedAt,
		&doctor.UpdatedAt,
	)
	if err != nil {
		return domain.Doctor{}, fmt.Errorf("creating doctor: %w", err)
	}
	return doctor, nil

}

func (r *DoctorRepository) FindByEmail(ctx context.Context, email string) (domain.Doctor, string, error) {
	query := `
		SELECT id, name, email, phone, password, specialty, license_number,
		       years_experience, consultation_fee, bio, is_active, is_verified, is_email_verified, created_at, updated_at
		FROM doctors WHERE email = $1
	`
	var doctor domain.Doctor
	var hashedPassword string
	err := r.db.QueryRow(ctx, query, email).Scan(
		&doctor.ID,
		&doctor.Name,
		&doctor.Email,
		&doctor.Phone,
		&hashedPassword,
		&doctor.Specialty,
		&doctor.LicenseNumber,
		&doctor.YearsExperience,
		&doctor.ConsultationFee,
		&doctor.Bio,
		&doctor.IsActive,
		&doctor.IsVerified,
		&doctor.IsEmailVerified,
		&doctor.CreatedAt,
		&doctor.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Doctor{}, "", domain.ErrDoctorNotFound
		}
		return domain.Doctor{}, "", fmt.Errorf("finding doctor by email: %w", err)
	}
	return doctor, hashedPassword, nil
}

func (r *DoctorRepository) FindByID(ctx context.Context, id string) (domain.Doctor, string, error) {
	query := `
		SELECT id, name, email, phone, password, specialty, license_number,
		       years_experience, consultation_fee, bio, is_active, is_verified, is_email_verified,created_at, updated_at
		FROM doctors WHERE id = $1
	`
	var doctor domain.Doctor
	var hashedPassword string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&doctor.ID,
		&doctor.Name,
		&doctor.Email,
		&doctor.Phone,
		&hashedPassword,
		&doctor.Specialty,
		&doctor.LicenseNumber,
		&doctor.YearsExperience,
		&doctor.ConsultationFee,
		&doctor.IsEmailVerified,
		&doctor.Bio,
		&doctor.IsActive,
		&doctor.IsVerified,
		&doctor.CreatedAt,
		&doctor.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Doctor{}, "", domain.ErrDoctorNotFound
		}
		return domain.Doctor{}, "", fmt.Errorf("finding doctor by id: %w", err)
	}
	return doctor, hashedPassword, nil
}

func (r *DoctorRepository) FindAll(ctx context.Context, specialty string) ([]domain.Doctor, error) {
	query := `
		SELECT id, name, email, phone, specialty, license_number,
		       years_experience, consultation_fee, bio, is_active, is_verified, is_email_verified, created_at, updated_at
		FROM doctors
		WHERE is_active = true
		AND ($1 = '' OR specialty ILIKE $1)
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, specialty)
	if err != nil {
		return nil, fmt.Errorf("finding all doctors: %w", err)
	}
	defer rows.Close()

	var doctors []domain.Doctor
	for rows.Next() {
		var doctor domain.Doctor
		err := rows.Scan(
			&doctor.ID,
			&doctor.Name,
			&doctor.Email,
			&doctor.Phone,
			&doctor.Specialty,
			&doctor.LicenseNumber,
			&doctor.YearsExperience,
			&doctor.ConsultationFee,
			&doctor.Bio,
			&doctor.IsActive,
			&doctor.IsVerified,
			&doctor.IsEmailVerified,
			&doctor.CreatedAt,
			&doctor.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning doctor row: %w", err)
		}
		doctors = append(doctors, doctor)
	}
	return doctors, nil
}

func (r *DoctorRepository) MarkEmailVerified(ctx context.Context, email string) error {
	query := `UPDATE doctors SET is_email_verified = true, updated_at = NOW() WHERE email = $1`
	_, err := r.db.Exec(ctx, query, email)
	if err != nil {
		return fmt.Errorf("marking doctor email verified: %w", err)
	}
	return nil
}

func (r *DoctorRepository) UpdatePassword(ctx context.Context, email, hashedPassword string) error {
	query := `UPDATE doctors SET password = $1, updated_at = NOW() WHERE email = $2`
	_, err := r.db.Exec(ctx, query, hashedPassword, email)
	if err != nil {
		return fmt.Errorf("updating doctor password: %w", err)
	}
	return nil
}
