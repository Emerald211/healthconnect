package domain

import "time"

type Patient struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Address     string    `json:"address"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RegisterPatientDTO struct {
	Name        string `json:"name"   binding:"required,min=2,max=100"`
	Email       string `json:"email"  binding:"required,email"`
	Phone       string `json:"phone"  binding:"required,min=10,max=20"`
	Password    string `json:"password"      binding:"required,min=8"`
	DateOfBirth string `json:"date_of_birth" binding:"required"`
	Gender      string `json:"gender"    binding:"required,oneof=male female other"`
	Address     string `json:"address"`
}

type LoginDto struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type AuthResponse struct {
	Patient     Patient `json:"patient"`
	AccessToken string  `json:"access_token"`
}

var (
	ErrPatientNotFound   = NewAppError("patient_not_found",   "patient not found",            404)
	ErrEmailAlreadyTaken = NewAppError("email_taken",         "email already registered",      409)
	ErrInvalidPassword   = NewAppError("invalid_password",    "invalid email or password",     401)
)