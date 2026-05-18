package handler

import (
	"net/http"

	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/Emerald211/healthconnect/internal/service"
	"github.com/Emerald211/healthconnect/pkg/response"
	"github.com/gin-gonic/gin"
)

type PatientHandler struct {
	service *service.PatientService
}

func NewPatientHandler(service *service.PatientService) *PatientHandler {
	return &PatientHandler{service: service}
}

func (h *PatientHandler) Register(c *gin.Context) {
	var req domain.RegisterPatientDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})

		return
	}

	newPatient, err := h.service.RegisterPatient(c.Request.Context(), req)

	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusCreated, newPatient)
}

func (h *PatientHandler) Login(c *gin.Context) {
	var req domain.LoginDto

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})

		return
	}

	patient, err := h.service.LoginPatient(c.Request.Context(), req)

	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, patient)
}

func (h *PatientHandler) RefreshToken(c *gin.Context) {
	var req domain.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.RefreshToken(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

func (h *PatientHandler) VerifyEmail(c *gin.Context) {
	var req domain.VerifyEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})

		return
	}

	if err := h.service.VerifyEmail(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "email verified successfully",
	})
}

func (h *PatientHandler) ResendOTP(c *gin.Context) {
	var req domain.ResendOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SendVerificationOTP(c.Request.Context(), req.Email); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "otp resent successfully",
	})
}

func (h *PatientHandler) ForgotPassword(c *gin.Context) {
	var req domain.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})

		return
	}

	h.service.ForgotPassword(c.Request.Context(), req)

	response.Success(c, http.StatusOK, gin.H{
		"message": "if your email is registered you will receive a reset code",
	})
}

func (h *PatientHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "password reset successfully, please login",
	})
}

func (h *PatientHandler) ChangePassword(c *gin.Context) {
	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.service.ChangePassword(c.Request.Context(), userID.(string), req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "password changed successfully",
	})
}

func (h *PatientHandler) Logout(c *gin.Context) {

	patientID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
			"message": "not authenticated",
		})
		return
	}

	if err := h.service.Logout(c.Request.Context(), patientID.(string)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

func (h *PatientHandler) GetMe(c *gin.Context) {
	patientId, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
			"message": "not authenticated",
		})

		return
	}

	patient, err := h.service.GetMe(c.Request.Context(), patientId.(string))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, patient)
}
