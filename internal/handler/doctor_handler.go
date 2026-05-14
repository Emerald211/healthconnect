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

	result, err := h.service.RegisterDoctor(c, req)

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

	result, err := h.service.Login(c, req)

	if err != nil {
		response.Error(c, err)
	}

	response.Success(c, http.StatusOK, result)
}

func (h *DoctorHandler) GetMe(c *gin.Context) {
	doctorID, exists := c.Get("patient_id")
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
