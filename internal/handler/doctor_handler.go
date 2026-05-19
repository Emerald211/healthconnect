package handler

import (
	"net/http"

	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/Emerald211/healthconnect/internal/service"
	"github.com/Emerald211/healthconnect/pkg/response"
	"github.com/gin-gonic/gin"
)

type DoctorHandler struct {
	service *service.DoctorService
}

func NewDoctorHandler(service *service.DoctorService) *DoctorHandler {
	return &DoctorHandler{service: service}
}

// RegisterDoctor godoc
// @Summary      Register a new doctor
// @Description  Creates a new doctor account and sends OTP to email
// @Tags         doctor-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.RegisterDoctorDTO  true  "Doctor registration data"
// @Success      201      {object}  domain.DoctorAuthResponse
// @Failure      400      {object}  map[string]interface{}
// @Failure      409      {object}  map[string]interface{}
// @Router       /api/v1/doctors/auth/register [post]
func (h *DoctorHandler) RegisterDoctor(c *gin.Context) {
	var req domain.RegisterDoctorDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})

		return
	}

	result, err := h.service.RegisterDoctor(c.Request.Context(), req)

	if err != nil {
		response.Error(c, err)
	}

	response.Success(c, http.StatusCreated, result)
}

// Login godoc
// @Summary      Login as a doctor
// @Description  Authenticates a doctor and returns access + refresh tokens
// @Tags         doctor-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.LoginDto  true  "Login credentials"
// @Success      200      {object}  domain.DoctorAuthResponse
// @Failure      401      {object}  map[string]interface{}
// @Router       /api/v1/doctors/auth/login [post]
func (h *DoctorHandler) Login(c *gin.Context) {
	var req domain.LoginDto

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})

		return
	}

	result, err := h.service.Login(c.Request.Context(), req)

	if err != nil {
		response.Error(c, err)
	}

	response.Success(c, http.StatusOK, result)
}

// RefreshToken godoc
// @Summary      Refresh doctor access token
// @Description  Issues a new access token using a valid refresh token
// @Tags         doctor-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.RefreshTokenRequest  true  "Refresh token data"
// @Success      200      {object}  domain.DoctorAuthResponse
// @Failure      401      {object}  map[string]interface{}
// @Router       /api/v1/doctors/auth/refresh [post]
func (h *DoctorHandler) RefreshToken(c *gin.Context) {
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
// @Summary      Verify doctor email with OTP
// @Description  Verifies a doctor's email using the OTP sent during registration
// @Tags         doctor-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.VerifyEmailRequest  true  "Email and OTP"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Router       /api/v1/doctors/auth/verify-email [post]
func (h *DoctorHandler) VerifyEmail(c *gin.Context) {
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
// @Summary      Resend doctor verification OTP
// @Description  Resends the email verification OTP to the doctor
// @Tags         doctor-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.ResendOTPRequest  true  "Doctor email"
// @Success      200      {object}  map[string]interface{}
// @Router       /api/v1/doctors/auth/resend-otp [post]
func (h *DoctorHandler) ResendOTP(c *gin.Context) {
	var req domain.ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SendVerificationEmail(c.Request.Context(), req.Email); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "OTP sent successfully",
	})
}

// ForgotPassword godoc
// @Summary      Request doctor password reset
// @Description  Sends a password reset OTP to the doctor's email
// @Tags         doctor-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.ForgotPasswordRequest  true  "Doctor email"
// @Success      200      {object}  map[string]interface{}
// @Router       /api/v1/doctors/auth/forgot-password [post]
func (h *DoctorHandler) ForgotPassword(c *gin.Context) {
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
// @Summary      Reset doctor password with OTP
// @Description  Resets the doctor's password using the OTP sent to their email
// @Tags         doctor-auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.ResetPasswordRequest  true  "Reset password data"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Router       /api/v1/doctors/auth/reset-password [post]
func (h *DoctorHandler) ResetPassword(c *gin.Context) {
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
// @Summary      Change doctor password
// @Description  Changes the authenticated doctor's password
// @Tags         doctors
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      domain.ChangePasswordRequest  true  "Change password data"
// @Success      200      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Router       /api/v1/doctors/change-password [post]
func (h *DoctorHandler) ChangePassword(c *gin.Context) {
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
// @Summary      Logout doctor
// @Description  Invalidates the doctor's refresh token
// @Tags         doctors
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/doctors/logout [post]
func (h *DoctorHandler) Logout(c *gin.Context) {
	doctorID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
			"message": "not authenticated",
		})

		return
	}

	if err := h.service.Logout(c.Request.Context(), doctorID.(string)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// GetMe godoc
// @Summary      Get current doctor profile
// @Description  Returns the authenticated doctor's profile
// @Tags         doctors
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  domain.Doctor
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/doctors/me [get]
func (h *DoctorHandler) GetMe(c *gin.Context) {
	doctorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
			"message": "not authenticated",
		})
		return
	}

	doctor, err := h.service.GetMe(c.Request.Context(), doctorID.(string))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, doctor)
}

// GetAll godoc
// @Summary      List all doctors
// @Description  Returns all active doctors, optionally filtered by specialty
// @Tags         doctors
// @Produce      json
// @Security     BearerAuth
// @Param        specialty  query     string  false  "Filter by specialty"
// @Success      200        {object}  []domain.Doctor
// @Failure      401        {object}  map[string]interface{}
// @Router       /api/v1/doctors [get]
func (h *DoctorHandler) GetAll(c *gin.Context) {
	// Optional query param: /api/v1/doctors?specialty=cardiology
	specialty := c.Query("specialty")

	doctors, err := h.service.GetAllDoctors(c.Request.Context(), specialty)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, doctors)
}
