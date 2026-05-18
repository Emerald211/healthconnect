package domain

import "fmt"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

var (
	ErrInvalidToken     = NewAppError("invalid_token", "invalid or expired token", 401)
	ErrTooManyAttempts  = NewAppError("too_many_attempts", "too many failed attempts, try again in 15 minutes", 429)
	ErrInvalidOTP       = NewAppError("invalid_otp", "invalid or expired OTP", 400)
	ErrEmailNotVerified = NewAppError("email_not_verified", "please verify your email first", 403)
	ErrAlreadyVerified  = NewAppError("already_verified", "email is already verified", 400)
)
