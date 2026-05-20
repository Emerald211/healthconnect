package service

import (
	"context"
	"fmt"

	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/Emerald211/healthconnect/internal/repository"
)

type AppointmentService struct {
	repo       *repository.AppointmentRepository
	doctorRepo *repository.DoctorRepository
}

func NewAppointmentService(repo *repository.AppointmentRepository, doctorRepo *repository.DoctorRepository) *AppointmentService {
	return &AppointmentService{
		repo:       repo,
		doctorRepo: doctorRepo,
	}
}

func (s *AppointmentService) SetAvailability(ctx context.Context, doctorID string, req domain.SetAvailabilityRequest) (domain.DoctorAvailability, error) {
	availability, err := s.repo.SetAvailability(ctx, doctorID, req)

	if err != nil {
		return domain.DoctorAvailability{}, fmt.Errorf("failed to set availability: %w", err)
	}

	return availability, nil
}

func (s *AppointmentService) GetAvailability(ctx context.Context, doctorID string) ([]domain.DoctorAvailability, error) {
	availabities, err := s.repo.GetAvailability(ctx, doctorID)

	if err != nil {
		return nil, fmt.Errorf("failed to get availability: %w", err)
	}

	return availabities, nil
}

func (s *AppointmentService) GenerateBookableSlots(ctx context.Context, doctorID string, date string) ([]domain.AppointmentSlot, error) {
	slots, err := s.repo.GenerateSlots(ctx, doctorID, date)

	if err != nil {
		return nil, fmt.Errorf("failed to generate bookable slots: %w", err)
	}

	return slots, nil
}

func (s *AppointmentService) GetAvailableSlots(ctx context.Context, doctorID, date string) ([]domain.AppointmentSlot, error) {
	slots, err := s.repo.GetAvailableSlots(ctx, doctorID, date)
	if err != nil {
		return nil, fmt.Errorf("appointment service get available slots: %w", err)
	}
	return slots, nil
}

func (s *AppointmentService) getDoctorIdbySlot(ctx context.Context, slotID string) (domain.Doctor, error) {
	doctorID, err := s.repo.GetDoctorIDBySlot(ctx, slotID)

	if err != nil {
		return domain.Doctor{}, fmt.Errorf("failed to get doctor id by slot: %w", err)
	}

	doctor, _, err := s.doctorRepo.FindByID(ctx, doctorID)
	if err != nil {
		return domain.Doctor{}, fmt.Errorf("failed to get doctor by id: %w", err)
	}

	return doctor, nil
}

func (s *AppointmentService) BookAppointment(ctx context.Context, patientId string, req domain.BookAppointmentRequest) (domain.Appointment, error) {
	doctor, err := s.getDoctorIdbySlot(ctx, req.SlotID)
	if err != nil {
		return domain.Appointment{}, fmt.Errorf("failed to get doctor by slot: %w", err)
	}

	appt, err := s.repo.BookAppointment(ctx, patientId, req, doctor.ConsultationFee)

	if err != nil {
		return domain.Appointment{}, fmt.Errorf("failed to book appointment: %w", err)
	}

	return appt, nil
}

func (s *AppointmentService) GetPatientAppointments(ctx context.Context, patientID string) ([]domain.Appointment, error) {
	appointments, err := s.repo.GetPatientAppointments(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("appointment service get patient appointments: %w", err)
	}
	return appointments, nil
}

func (s *AppointmentService) GetDoctorAppointments(ctx context.Context, doctorID string) ([]domain.Appointment, error) {
	appointments, err := s.repo.GetDoctorAppointments(ctx, doctorID)
	if err != nil {
		return nil, fmt.Errorf("appointment service get doctor appointments: %w", err)
	}
	return appointments, nil
}

func (s *AppointmentService) UpdateAppointmentStatus(ctx context.Context, doctorID, appointmentID string, req domain.UpdateAppointmentStatusRequest) (domain.Appointment, error) {
	appointment, err := s.repo.GetByID(ctx, appointmentID)

	if err != nil {
		return domain.Appointment{}, fmt.Errorf("failed to get appointment: %w", err)
	}

	if appointment.DoctorID != doctorID {
		return domain.Appointment{}, fmt.Errorf("appointment does not belong to doctor")
	}

	updatedAppointment, err := s.repo.UpdateStatus(ctx, appointmentID, req.Status, req.DoctorNotes)
	if err != nil {
		return domain.Appointment{}, fmt.Errorf("failed to update appointment status: %w", err)
	}

	return updatedAppointment, nil
}
