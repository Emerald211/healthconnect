package service

import (
	"context"
	"fmt"

	"github.com/Emerald211/healthconnect/internal/config"
	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/Emerald211/healthconnect/internal/repository"
	"github.com/Emerald211/healthconnect/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type PatientService struct {
	repo *repository.PatientRepository
	cfg  *config.Config
}

func NewPatientService(repo *repository.PatientRepository, cfg *config.Config) *PatientService {
	return &PatientService{repo: repo, cfg: cfg}
}

func (s *PatientService) RegisterPatient(ctx context.Context, req domain.RegisterPatientDTO) (domain.AuthResponse, error) {

	exists, err := s.repo.EmailExists(ctx, req.Email)

	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("register patient: %w", err)
	}

	if exists {
		return domain.AuthResponse{}, domain.ErrEmailAlreadyTaken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)

	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("hashing password: %w", err)
	}

	dob, err := repository.ParseDate(req.DateOfBirth)

	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("parsing date of birth: %w", err)
	}

	patient := domain.Patient{
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		DateOfBirth: dob,
		Gender:      req.Gender,
		Address:     req.Address,
	}

	newUser, err := s.repo.Create(ctx, patient, string(hashedPassword))

	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("creating patient: %w", err)
	}

	token, err := jwt.GenerateToken(newUser.ID, newUser.Email, "patient", s.cfg.JWTSecret, s.cfg.JWTExpiryHours)
	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("generating token: %w", err)
	}

	return domain.AuthResponse{
		Patient:     newUser,
		AccessToken: token,
	}, nil

}

func (s *PatientService) LoginPatient(ctx context.Context, req domain.LoginDto) (domain.AuthResponse, error) {

	patient, hashedPassword, err := s.repo.FindByEmail(ctx, req.Email)

	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("login patient: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return domain.AuthResponse{}, domain.ErrInvalidPassword
	}

	token, err := jwt.GenerateToken(patient.ID, patient.Email, "patient", s.cfg.JWTSecret, s.cfg.JWTExpiryHours)
	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("generating token: %w", err)
	}

	return domain.AuthResponse{
		Patient:     patient,
		AccessToken: token,
	}, nil
}

func (s *PatientService) GetMe(ctx context.Context, patientID string) (domain.Patient, error) {
	patient, err := s.repo.FindByID(ctx, patientID)

	if err != nil {
		return domain.Patient{}, fmt.Errorf("patient service get me: %w", err)
	}

	return patient, nil
}
