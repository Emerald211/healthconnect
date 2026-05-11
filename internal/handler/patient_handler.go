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
