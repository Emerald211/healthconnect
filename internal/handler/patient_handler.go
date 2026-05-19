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

// Register godoc
// @Summary      Register a new patient
// @Description  Creates a new patient account and sends OTP to email
// @Tags         patient-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.RegisterPatientDTO  true  "Patient registration data"
// @Success      201      {object}  domain.AuthResponse
// @Failure      400      {object}  map[string]interface{}
// @Failure      409      {object}  map[string]interface{}
// @Router       /api/v1/auth/register [post]
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

// Login godoc
// @Summary      Login as a patient
// @Description  Authenticates a patient and returns access + refresh tokens
// @Tags         patient-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.LoginDto  true  "Login credentials"
// @Success      200      {object}  domain.AuthResponse
// @Failure      401      {object}  map[string]interface{}
// @Failure      429      {object}  map[string]interface{}
// @Router       /api/v1/auth/login [post]
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

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Issues a new access token using a valid refresh token
// @Tags         patient-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.RefreshTokenRequest  true  "Refresh token data"
// @Success      200      {object}  domain.AuthResponse
// @Failure      401      {object}  map[string]interface{}
// @Router       /api/v1/auth/refresh [post]
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

// VerifyEmail godoc
// @Summary      Verify email with OTP
// @Description  Verifies a patient's email using the OTP sent during registration
// @Tags         patient-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.VerifyEmailRequest  true  "Email and OTP"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Router       /api/v1/auth/verify-email [post]
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
// ResendOTP godoc
// @Summary      Resend verification OTP
// @Description  Resends the email verification OTP to the patient
// @Tags         patient-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.ResendOTPRequest  true  "Patient email"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Router       /api/v1/auth/resend-otp [post]
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
// ForgotPassword godoc
// @Summary      Request password reset
// @Description  Sends a password reset OTP to the patient's email
// @Tags         patient-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.ForgotPasswordRequest  true  "Patient email"
// @Success      200      {object}  map[string]interface{}
// @Router       /api/v1/auth/forgot-password [post]
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
// ResetPassword godoc
// @Summary      Reset password with OTP
// @Description  Resets the patient's password using the OTP sent to their email
// @Tags         patient-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.ResetPasswordRequest  true  "Reset password data"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Router       /api/v1/auth/reset-password [post]
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
// ChangePassword godoc
// @Summary      Change password
// @Description  Changes the authenticated patient's password
// @Tags         patients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      domain.ChangePasswordRequest  true  "Change password data"
// @Success      200      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Router       /api/v1/patients/change-password [post]
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
// Logout godoc
// @Summary      Logout patient
// @Description  Invalidates the patient's refresh token
// @Tags         patients
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/patients/logout [post]
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
// GetMe godoc
// @Summary      Get current patient profile
// @Description  Returns the authenticated patient's profile
// @Tags         patients
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  domain.Patient
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/patients/me [get]
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
