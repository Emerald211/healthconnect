package domain

import "time"

type DoctorAvailability struct {
	ID                  string    `json:"id"`
	DoctorID            string    `json:"doctor_id"`
	DayOfWeek           int       `json:"day_of_week"` // 0=Sunday, 1=Monday...6=Saturday
	StartTime           string    `json:"start_time"`  // "09:00"
	EndTime             string    `json:"end_time"`    // "17:00"
	SlotDurationMinutes int       `json:"slot_duration_minutes"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// AppointmentSlot is a specific bookable time slot
type AppointmentSlot struct {
	ID        string    `json:"id"`
	DoctorID  string    `json:"doctor_id"`
	Date      time.Time `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	IsBooked  bool      `json:"is_booked"`
	CreatedAt time.Time `json:"created_at"`
}

// Appointment is a confirmed booking
type Appointment struct {
	ID          string    `json:"id"`
	PatientID   string    `json:"patient_id"`
	DoctorID    string    `json:"doctor_id"`
	SlotID      string    `json:"slot_id"`
	Status      string    `json:"status"`       // pending, confirmed, completed, cancelled
	Type        string    `json:"type"`         // consultation, follow_up, emergency
	Notes       string    `json:"notes"`        // patient's reason for visit
	DoctorNotes *string    `json:"doctor_notes"` // filled after consultation
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Joined fields — populated when we fetch with joins
	Patient *Patient         `json:"patient,omitempty"`
	Doctor  *Doctor          `json:"doctor,omitempty"`
	Slot    *AppointmentSlot `json:"slot,omitempty"`
}

// Payment tracks payment for an appointment
type Payment struct {
	ID                 string     `json:"id"`
	AppointmentID      string     `json:"appointment_id"`
	PatientID          string     `json:"patient_id"`
	Amount             float64    `json:"amount"`
	Currency           string     `json:"currency"`
	Status             string     `json:"status"` // pending, successful, failed, refunded
	PaystackReference  string     `json:"paystack_reference"`
	PaystackAccessCode string     `json:"paystack_access_code"`
	PaidAt             *time.Time `json:"paid_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// --- Request DTOs ---

type SetAvailabilityRequest struct {
	DayOfWeek           int    `json:"day_of_week"            binding:"required,min=0,max=6"`
	StartTime           string `json:"start_time"             binding:"required"` // "09:00"
	EndTime             string `json:"end_time"               binding:"required"` // "17:00"
	SlotDurationMinutes int    `json:"slot_duration_minutes"  binding:"required,min=15,max=120"`
}

type GenerateSlotsRequest struct {
	Date string `json:"date" binding:"required"` // "2026-06-02"
}

type BookAppointmentRequest struct {
	SlotID string `json:"slot_id" binding:"required"`
	Type   string `json:"type"    binding:"required,oneof=consultation follow_up emergency"`
	Notes  string `json:"notes"`
}

type UpdateAppointmentStatusRequest struct {
	Status      string `json:"status"       binding:"required,oneof=confirmed completed cancelled"`
	DoctorNotes *string `json:"doctor_notes"`
}

// --- Errors ---
var (
	ErrSlotNotFound        = NewAppError("slot_not_found", "appointment slot not found", 404)
	ErrSlotAlreadyBooked   = NewAppError("slot_already_booked", "this slot has already been booked", 409)
	ErrAppointmentNotFound = NewAppError("appointment_not_found", "appointment not found", 404)
	ErrAvailabilityExists  = NewAppError("availability_exists", "availability for this day already set", 409)
	ErrNotYourAppointment  = NewAppError("not_your_appointment", "this appointment does not belong to you", 403)
)
