package domain

import "time"

type Doctor struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Specialty       string    `json:"specialty"`
	LicenseNumber   string    `json:"license_number"`
	YearsExperience int       `json:"years_experience"`
	ConsultationFee float64   `json:"consultation_fee"`
	Bio             string    `json:"bio"`
	IsActive        bool      `json:"is_active"`
	IsVerified      bool      `json:"is_verified"`
	IsEmailVerified bool `json:"is_email_verified"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RegisterDoctorDTO struct {
	Name            string  `json:"name"             binding:"required,min=2,max=100"`
	Email           string  `json:"email"            binding:"required,email"`
	Phone           string  `json:"phone"            binding:"required,min=10,max=20"`
	Password        string  `json:"password"         binding:"required,min=8"`
	Specialty       string  `json:"specialty"        binding:"required"`
	LicenseNumber   string  `json:"license_number"   binding:"required"`
	YearsExperience int     `json:"years_experience" binding:"required,min=0"`
	ConsultationFee float64 `json:"consultation_fee" binding:"required,min=0"`
	Bio             string  `json:"bio"`
}

type DoctorAuthResponse struct {
	Doctor      Doctor `json:"doctor"`
	AccessToken string `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
}

var (
	ErrDoctorNotFound    = NewAppError("doctor_not_found", "doctor not found", 404)
	ErrLicenseTaken      = NewAppError("license_taken", "license number already registered", 409)
	ErrDoctorEmailTaken  = NewAppError("email_taken", "email already registered", 409)
	ErrDoctorInvalidPass = NewAppError("invalid_credentials", "invalid email or password", 401)
)
