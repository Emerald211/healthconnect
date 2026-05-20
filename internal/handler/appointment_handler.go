package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/Emerald211/healthconnect/internal/service"
	"github.com/Emerald211/healthconnect/pkg/response"
)

type AppointmentHandler struct {
	service *service.AppointmentService
}

func NewAppointmentHandler(service *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{service: service}
}

// SetAvailability godoc
// @Summary      Set doctor availability
// @Description  Allows a doctor to set their weekly availability schedule
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      domain.SetAvailabilityRequest  true  "Availability data"
// @Success      200      {object}  domain.DoctorAvailability
// @Failure      400      {object}  map[string]interface{}
// @Router       /api/v1/appointments/availability [post]
func (h *AppointmentHandler) SetAvailability(c *gin.Context) {
	var req domain.SetAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	doctorID, _ := c.Get("user_id")

	availability, err := h.service.SetAvailability(c.Request.Context(), doctorID.(string), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, availability)
}

// GetAvailability godoc
// @Summary      Get doctor availability
// @Description  Returns a doctor's weekly availability schedule
// @Tags         appointments
// @Produce      json
// @Security     BearerAuth
// @Param        doctor_id  path  string  true  "Doctor ID"
// @Success      200        {object}  []domain.DoctorAvailability
// @Router       /api/v1/appointments/availability/{doctor_id} [get]
func (h *AppointmentHandler) GetAvailability(c *gin.Context) {
	doctorID := c.Param("doctor_id")

	availability, err := h.service.GetAvailability(c.Request.Context(), doctorID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, availability)
}

// GenerateSlots godoc
// @Summary      Generate appointment slots
// @Description  Generates bookable slots for a doctor on a specific date
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        doctor_id  path  string  true  "Doctor ID"
// @Param        date       query  string  true  "Date (YYYY-MM-DD)"
// @Success      200        {object}  []domain.AppointmentSlot
// @Router       /api/v1/appointments/slots/{doctor_id}/generate [post]
func (h *AppointmentHandler) GenerateSlots(c *gin.Context) {
	doctorID := c.Param("doctor_id")
	date := c.Query("date")

	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": "date query parameter is required (YYYY-MM-DD)",
		})
		return
	}

	slots, err := h.service.GenerateBookableSlots(c.Request.Context(), doctorID, date)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, slots)
}

// GetAvailableSlots godoc
// @Summary      Get available slots
// @Description  Returns all available (unbooked) slots for a doctor on a date
// @Tags         appointments
// @Produce      json
// @Security     BearerAuth
// @Param        doctor_id  path   string  true  "Doctor ID"
// @Param        date       query  string  true  "Date (YYYY-MM-DD)"
// @Success      200        {object}  []domain.AppointmentSlot
// @Router       /api/v1/appointments/slots/{doctor_id} [get]
func (h *AppointmentHandler) GetAvailableSlots(c *gin.Context) {
	doctorID := c.Param("doctor_id")
	date := c.Query("date")

	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": "date query parameter is required (YYYY-MM-DD)",
		})
		return
	}

	slots, err := h.service.GetAvailableSlots(c.Request.Context(), doctorID, date)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, slots)
}

// BookAppointment godoc
// @Summary      Book an appointment
// @Description  Patient books an appointment slot with a doctor
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      domain.BookAppointmentRequest  true  "Booking data"
// @Success      201      {object}  domain.Appointment
// @Failure      409      {object}  map[string]interface{}
// @Router       /api/v1/appointments [post]
func (h *AppointmentHandler) BookAppointment(c *gin.Context) {
	var req domain.BookAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	patientID, _ := c.Get("user_id")

	appt, err := h.service.BookAppointment(c.Request.Context(), patientID.(string), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusCreated, appt)
}

// GetMyAppointments godoc
// @Summary      Get patient appointments
// @Description  Returns all appointments for the logged-in patient
// @Tags         appointments
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  []domain.Appointment
// @Router       /api/v1/appointments/my [get]
func (h *AppointmentHandler) GetMyAppointments(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var appointments []domain.Appointment
	var err error

	if role == "patient" {
		appointments, err = h.service.GetPatientAppointments(c.Request.Context(), userID.(string))
	} else {
		appointments, err = h.service.GetDoctorAppointments(c.Request.Context(), userID.(string))
	}

	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, appointments)
}

// UpdateAppointmentStatus godoc
// @Summary      Update appointment status
// @Description  Allows a doctor to update appointment status and add notes
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path   string                               true  "Appointment ID"
// @Param        request  body   domain.UpdateAppointmentStatusRequest true  "Status update"
// @Success      200      {object}  domain.Appointment
// @Router       /api/v1/appointments/{id}/status [patch]
func (h *AppointmentHandler) UpdateAppointmentStatus(c *gin.Context) {
	appointmentID := c.Param("id")

	var req domain.UpdateAppointmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	doctorID, _ := c.Get("user_id")

	appt, err := h.service.UpdateAppointmentStatus(c.Request.Context(), doctorID.(string), appointmentID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, appt)
}
