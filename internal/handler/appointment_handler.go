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
// @Description  Allows an authenticated doctor to set or update their weekly availability schedule.
// @Tags         Appointments
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
// @Description  Returns a specific doctor's weekly availability schedule. Accessible by any authenticated user.
// @Tags         Appointments
// @Produce      json
// @Security     BearerAuth
// @Param        doctor_id  path  string  true  "The ID of the doctor to retrieve availability for."
// @Success      200        {array}   domain.DoctorAvailability
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
// @Description  Allows an authenticated doctor to generate bookable appointment slots for a specific date based on their availability.
// @Tags         Appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        doctor_id  path  string  true  "The ID of the doctor to generate slots for."
// @Param        date       query  string  true  "The date (YYYY-MM-DD) to generate slots for."
// @Success      200        {array}   domain.AppointmentSlot
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
// @Summary      Get available appointment slots
// @Description  Returns all available (unbooked) appointment slots for a specific doctor on a given date. Accessible by any authenticated user.
// @Tags         Appointments
// @Produce      json
// @Security     BearerAuth
// @Param        doctor_id  path   string  true  "The ID of the doctor to retrieve slots for."
// @Param        date       query  string  true  "The date (YYYY-MM-DD) to check for available slots."
// @Success      200        {array}   domain.AppointmentSlot
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
// @Description  Allows an authenticated patient to book an appointment slot with a doctor.
// @Tags         Appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      domain.BookAppointmentRequest  true  "Booking data for the appointment."
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

// GetPatientAppointments godoc
// @Summary      Get patient appointments
// @Description  Returns all appointments for the currently authenticated patient.
// @Tags         Appointments
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.Appointment
// @Router       /api/v1/appointments/my [get]
func (h *AppointmentHandler) GetPatientAppointments(c *gin.Context) {
	userID, _ := c.Get("user_id")

	appointments, err := h.service.GetPatientAppointments(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, appointments)
}

// GetDoctorAppointments godoc
// @Summary      Get doctor's appointments
// @Description  Returns all appointments for the currently authenticated doctor. This endpoint is similar to `/api/v1/appointments/my` but for doctors.
// @Tags         Appointments
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.Appointment
// @Router       /api/v1/appointments/doctor/my [get]
func (h *AppointmentHandler) GetDoctorAppointments(c *gin.Context) {
	userID, _ := c.Get("user_id")

	appointments, err := h.service.GetDoctorAppointments(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, appointments)
}

// UpdateAppointmentStatus godoc
// @Summary      Update appointment status
// @Description  Allows an authenticated doctor to update the status of an appointment and add notes.
// @Tags         Appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path   string                               true  "The ID of the appointment to update."
// @Param        request  body   domain.UpdateAppointmentStatusRequest true  "Status update and optional doctor notes."
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
