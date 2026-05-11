package domain

import "fmt"

type AppError struct {
	Code   string  `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}


func (e * AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}



