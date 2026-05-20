package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Emerald211/healthconnect/internal/domain"
)

type AppointmentRepository struct {
	db *pgxpool.Pool
}

func NewAppointmentRepository(db *pgxpool.Pool) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) SetAvailability(ctx context.Context, doctorID string, req domain.SetAvailabilityRequest) (domain.DoctorAvailability, error) {
	query := `
		INSERT INTO doctor_availability (doctor_id, day_of_week, start_time, end_time, slot_duration_minutes)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (doctor_id, day_of_week)
		DO UPDATE SET
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			slot_duration_minutes = EXCLUDED.slot_duration_minutes,
			updated_at = NOW()
		RETURNING id, doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, is_active, created_at, updated_at
	`

	var a domain.DoctorAvailability
	var startTime, endTime time.Time

	err := r.db.QueryRow(ctx, query,
		doctorID,
		req.DayOfWeek,
		req.StartTime,
		req.EndTime,
		req.SlotDurationMinutes,
	).Scan(
		&a.ID,
		&a.DoctorID,
		&a.DayOfWeek,
		&startTime,
		&endTime,
		&a.SlotDurationMinutes,
		&a.IsActive,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		return domain.DoctorAvailability{}, fmt.Errorf("setting availability: %w", err)
	}

	a.StartTime = startTime.Format("15:04")
	a.EndTime = endTime.Format("15:04")

	return a, nil
}

func (r *AppointmentRepository) GetAvailability(ctx context.Context, doctorID string) ([]domain.DoctorAvailability, error) {
	query := `
		SELECT id, doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, is_active, created_at, updated_at
		FROM doctor_availability
		WHERE doctor_id = $1 AND is_active = true
		ORDER BY day_of_week ASC
	`

	rows, err := r.db.Query(ctx, query, doctorID)
	if err != nil {
		return nil, fmt.Errorf("getting availability: %w", err)
	}
	defer rows.Close()

	var availabilities []domain.DoctorAvailability
	for rows.Next() {
		var a domain.DoctorAvailability
		var startTime, endTime time.Time
		err := rows.Scan(
			&a.ID,
			&a.DoctorID,
			&a.DayOfWeek,
			&startTime,
			&endTime,
			&a.SlotDurationMinutes,
			&a.IsActive,
			&a.CreatedAt,
			&a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning availability: %w", err)
		}
		a.StartTime = startTime.Format("15:04")
		a.EndTime = endTime.Format("15:04")
		availabilities = append(availabilities, a)
	}

	return availabilities, nil
}

// GenerateSlots creates bookable slots for a specific date
// based on the doctor's availability for that day of week
func (r *AppointmentRepository) GenerateSlots(ctx context.Context, doctorID, date string) ([]domain.AppointmentSlot, error) {
	// Parse the date
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("parsing date: %w", err)
	}

	dayOfWeek := int(parsedDate.Weekday())

	// Get doctor's availability for this day of week
	var startTimeStr, endTimeStr string
	var slotDuration int

	err = r.db.QueryRow(ctx, `
		SELECT start_time, end_time, slot_duration_minutes
		FROM doctor_availability
		WHERE doctor_id = $1 AND day_of_week = $2 AND is_active = true
	`, doctorID, dayOfWeek).Scan(&startTimeStr, &endTimeStr, &slotDuration)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("doctor is not available on this day")
		}
		return nil, fmt.Errorf("getting doctor availability: %w", err)
	}

	// Parse start and end times
	baseDate := parsedDate.Format("2006-01-02")
	startTime, err := time.Parse("2006-01-02 15:04:05", baseDate+" "+startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("parsing start time: %w", err)
	}
	endTime, err := time.Parse("2006-01-02 15:04:05", baseDate+" "+endTimeStr)
	if err != nil {
		return nil, fmt.Errorf("parsing end time: %w", err)
	}

	// Generate slots
	var slots []domain.AppointmentSlot
	current := startTime
	for current.Before(endTime) {
		slotEnd := current.Add(time.Duration(slotDuration) * time.Minute)
		if slotEnd.After(endTime) {
			break
		}

		// Insert slot — ignore if already exists (ON CONFLICT DO NOTHING)
		query := `
			INSERT INTO appointment_slots (doctor_id, date, start_time, end_time)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (doctor_id, date, start_time) DO NOTHING
			RETURNING id, doctor_id, date, start_time, end_time, is_booked, created_at
		`

		var slot domain.AppointmentSlot
		var slotStartTime, slotEndTime time.Time

		err := r.db.QueryRow(ctx, query,
			doctorID,
			parsedDate,
			current.Format("15:04:05"),
			slotEnd.Format("15:04:05"),
		).Scan(
			&slot.ID,
			&slot.DoctorID,
			&slot.Date,
			&slotStartTime,
			&slotEndTime,
			&slot.IsBooked,
			&slot.CreatedAt,
		)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("inserting slot: %w", err)
		}

		if slot.ID != "" {
			slot.StartTime = slotStartTime.Format("15:04")
			slot.EndTime = slotEndTime.Format("15:04")
			slots = append(slots, slot)
		}

		current = slotEnd
	}

	return slots, nil
}

// GetAvailableSlots returns all unbooked slots for a doctor on a date
func (r *AppointmentRepository) GetAvailableSlots(ctx context.Context, doctorID, date string) ([]domain.AppointmentSlot, error) {
	query := `
		SELECT id, doctor_id, date, start_time, end_time, is_booked, created_at
		FROM appointment_slots
		WHERE doctor_id = $1 AND date = $2 AND is_booked = false
		ORDER BY start_time ASC
	`

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("parsing date: %w", err)
	}

	rows, err := r.db.Query(ctx, query, doctorID, parsedDate)
	if err != nil {
		return nil, fmt.Errorf("getting available slots: %w", err)
	}
	defer rows.Close()

	var slots []domain.AppointmentSlot
	for rows.Next() {
		var slot domain.AppointmentSlot
		var startTime, endTime time.Time
		err := rows.Scan(
			&slot.ID,
			&slot.DoctorID,
			&slot.Date,
			&startTime,
			&endTime,
			&slot.IsBooked,
			&slot.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning slot: %w", err)
		}
		slot.StartTime = startTime.Format("15:04")
		slot.EndTime = endTime.Format("15:04")
		slots = append(slots, slot)
	}

	return slots, nil
}

func (r *AppointmentRepository) GetDoctorIDBySlot(ctx context.Context, slotID string) (string, error) {
	var doctorID string
	err := r.db.QueryRow(ctx, `
		SELECT doctor_id FROM appointment_slots WHERE id = $1
	`, slotID).Scan(&doctorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrSlotNotFound
		}
		return "", fmt.Errorf("getting doctor id by slot: %w", err)
	}
	return doctorID, nil
}

// BookAppointment creates an appointment with conflict detection
// Uses a database transaction to prevent double booking
func (r *AppointmentRepository) BookAppointment(ctx context.Context, patientID string, req domain.BookAppointmentRequest, amount float64) (domain.Appointment, error) {
	// Start a transaction
	// This is the key to preventing double bookings
	// Everything inside either ALL succeeds or ALL fails
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Appointment{}, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(ctx) // rollback if we return early with an error

	// Step 1 — Lock the slot row so no other transaction can touch it
	// "FOR UPDATE" means: lock this row until our transaction commits
	// Any other transaction trying to book the same slot will WAIT here
	var isBooked bool
	var doctorID string
	err = tx.QueryRow(ctx, `
		SELECT is_booked, doctor_id
		FROM appointment_slots
		WHERE id = $1
		FOR UPDATE
	`, req.SlotID).Scan(&isBooked, &doctorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Appointment{}, domain.ErrSlotNotFound
		}
		return domain.Appointment{}, fmt.Errorf("locking slot: %w", err)
	}

	// Step 2 — Check if already booked
	if isBooked {
		return domain.Appointment{}, domain.ErrSlotAlreadyBooked
	}

	// Step 3 — Mark slot as booked
	_, err = tx.Exec(ctx, `
		UPDATE appointment_slots SET is_booked = true WHERE id = $1
	`, req.SlotID)
	if err != nil {
		return domain.Appointment{}, fmt.Errorf("marking slot booked: %w", err)
	}

	// Step 4 — Create the appointment
	var appt domain.Appointment
	err = tx.QueryRow(ctx, `
		INSERT INTO appointments (patient_id, doctor_id, slot_id, type, notes, amount)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, patient_id, doctor_id, slot_id, status, type, notes, doctor_notes, amount, created_at, updated_at
	`,
		patientID,
		doctorID,
		req.SlotID,
		req.Type,
		req.Notes,
		amount,
	).Scan(
		&appt.ID,
		&appt.PatientID,
		&appt.DoctorID,
		&appt.SlotID,
		&appt.Status,
		&appt.Type,
		&appt.Notes,
		&appt.DoctorNotes,
		&appt.Amount,
		&appt.CreatedAt,
		&appt.UpdatedAt,
	)
	if err != nil {
		return domain.Appointment{}, fmt.Errorf("creating appointment: %w", err)
	}

	// Step 5 — Commit the transaction
	// Only now does everything become permanent
	if err := tx.Commit(ctx); err != nil {
		return domain.Appointment{}, fmt.Errorf("committing transaction: %w", err)
	}

	return appt, nil
}

func (r *AppointmentRepository) GetPatientAppointments(ctx context.Context, patientID string) ([]domain.Appointment, error) {
	query := `
		SELECT a.id, a.patient_id, a.doctor_id, a.slot_id, a.status, a.type,
		       a.notes, a.doctor_notes, a.amount, a.created_at, a.updated_at,
		       s.date, s.start_time, s.end_time,
		       d.name, d.specialty
		FROM appointments a
		JOIN appointment_slots s ON s.id = a.slot_id
		JOIN doctors d ON d.id = a.doctor_id
		WHERE a.patient_id = $1
		ORDER BY a.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, patientID)
	if err != nil {
		return nil, fmt.Errorf("getting patient appointments: %w", err)
	}
	defer rows.Close()

	var appointments []domain.Appointment
	for rows.Next() {
		var appt domain.Appointment
		var slotDate time.Time
		var slotStart, slotEnd time.Time
		var doctorName, doctorSpecialty string

		err := rows.Scan(
			&appt.ID,
			&appt.PatientID,
			&appt.DoctorID,
			&appt.SlotID,
			&appt.Status,
			&appt.Type,
			&appt.Notes,
			&appt.DoctorNotes,
			&appt.Amount,
			&appt.CreatedAt,
			&appt.UpdatedAt,
			&slotDate,
			&slotStart,
			&slotEnd,
			&doctorName,
			&doctorSpecialty,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning appointment: %w", err)
		}

		appt.Slot = &domain.AppointmentSlot{
			Date:      slotDate,
			StartTime: slotStart.Format("15:04"),
			EndTime:   slotEnd.Format("15:04"),
		}
		appt.Doctor = &domain.Doctor{
			Name:      doctorName,
			Specialty: doctorSpecialty,
		}

		appointments = append(appointments, appt)
	}

	return appointments, nil
}

func (r *AppointmentRepository) GetDoctorAppointments(ctx context.Context, doctorID string) ([]domain.Appointment, error) {
	query := `
		SELECT a.id, a.patient_id, a.doctor_id, a.slot_id, a.status, a.type,
		       a.notes, a.doctor_notes, a.amount, a.created_at, a.updated_at,
		       s.date, s.start_time, s.end_time,
		       p.name, p.phone
		FROM appointments a
		JOIN appointment_slots s ON s.id = a.slot_id
		JOIN patients p ON p.id = a.patient_id
		WHERE a.doctor_id = $1
		ORDER BY a.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, doctorID)
	if err != nil {
		return nil, fmt.Errorf("getting doctor appointments: %w", err)
	}
	defer rows.Close()

	var appointments []domain.Appointment
	for rows.Next() {
		var appt domain.Appointment
		var slotDate time.Time
		var slotStart, slotEnd time.Time
		var patientName, patientPhone string

		err := rows.Scan(
			&appt.ID,
			&appt.PatientID,
			&appt.DoctorID,
			&appt.SlotID,
			&appt.Status,
			&appt.Type,
			&appt.Notes,
			&appt.DoctorNotes,
			&appt.Amount,
			&appt.CreatedAt,
			&appt.UpdatedAt,
			&slotDate,
			&slotStart,
			&slotEnd,
			&patientName,
			&patientPhone,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning appointment: %w", err)
		}

		appt.Slot = &domain.AppointmentSlot{
			Date:      slotDate,
			StartTime: slotStart.Format("15:04"),
			EndTime:   slotEnd.Format("15:04"),
		}
		appt.Patient = &domain.Patient{
			Name:  patientName,
			Phone: patientPhone,
		}

		appointments = append(appointments, appt)
	}

	return appointments, nil
}

func (r *AppointmentRepository) GetByID(ctx context.Context, appointmentID string) (domain.Appointment, error) {
	query := `
		SELECT id, patient_id, doctor_id, slot_id, status, type,
		       notes, doctor_notes, amount, created_at, updated_at
		FROM appointments
		WHERE id = $1
	`

	var appt domain.Appointment
	err := r.db.QueryRow(ctx, query, appointmentID).Scan(
		&appt.ID,
		&appt.PatientID,
		&appt.DoctorID,
		&appt.SlotID,
		&appt.Status,
		&appt.Type,
		&appt.Notes,
		&appt.DoctorNotes,
		&appt.Amount,
		&appt.CreatedAt,
		&appt.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Appointment{}, domain.ErrAppointmentNotFound
		}
		return domain.Appointment{}, fmt.Errorf("getting appointment: %w", err)
	}

	return appt, nil
}

func (r *AppointmentRepository) UpdateStatus(ctx context.Context, appointmentID, status string, doctorNotes *string) (domain.Appointment, error) {
	query := `
		UPDATE appointments
		SET status = $1, doctor_notes = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, patient_id, doctor_id, slot_id, status, type, notes, doctor_notes, amount, created_at, updated_at
	`

	var appt domain.Appointment
	err := r.db.QueryRow(ctx, query, status, doctorNotes, appointmentID).Scan(
		&appt.ID,
		&appt.PatientID,
		&appt.DoctorID,
		&appt.SlotID,
		&appt.Status,
		&appt.Type,
		&appt.Notes,
		&appt.DoctorNotes,
		&appt.Amount,
		&appt.CreatedAt,
		&appt.UpdatedAt,
	)
	if err != nil {
		return domain.Appointment{}, fmt.Errorf("updating appointment status: %w", err)
	}

	return appt, nil
}
