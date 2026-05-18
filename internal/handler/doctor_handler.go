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
