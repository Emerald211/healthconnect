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

type DoctorService struct {
	repo *repository.DoctorRepository
	cfg  *config.Config
}

func NewDoctorService(repo *repository.DoctorRepository, cfg *config.Config) *DoctorService {
	return &DoctorService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *DoctorService) RegisterDoctor(ctx context.Context, req domain.RegisterDoctorDTO) (domain.DoctorAuthResponse, error) {
	emailExists, err := s.repo.EmailExists(ctx, req.Email)

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("registering doctor: %w", err)
	}

	if emailExists {
		return domain.DoctorAuthResponse{}, domain.ErrDoctorEmailTaken
	}

	licenseExists, err := s.repo.LicenseExists(ctx, req.LicenseNumber)

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("registering doctor: %w", err)
	}

	if licenseExists {
		return domain.DoctorAuthResponse{}, domain.ErrLicenseTaken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("hashing password: %w", err)
	}

	doctor := domain.Doctor{
		Name:            req.Name,
		Email:           req.Email,
		Phone:           req.Phone,
		Specialty:       req.Specialty,
		LicenseNumber:   req.LicenseNumber,
		YearsExperience: req.YearsExperience,
		ConsultationFee: req.ConsultationFee,
		Bio:             req.Bio,
	}

	newDoctor, err := s.repo.Create(ctx, doctor, string(hashedPassword))

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("doctor service register: %w", err)
	}

	token, err := jwt.GenerateToken(newDoctor.ID, newDoctor.Email, "doctor",s.cfg.JWTSecret, s.cfg.JWTExpiryHours)

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("generating token: %w", err)
	}

	return domain.DoctorAuthResponse{
		Doctor:      newDoctor,
		AccessToken: token,
	}, nil

}

func (s *DoctorService) Login(ctx context.Context, req domain.LoginDto) (domain.DoctorAuthResponse, error) {
	doctor, hashedPassword, err := s.repo.FindByEmail(ctx, req.Email)

	if err != nil {
		return domain.DoctorAuthResponse{}, domain.ErrDoctorInvalidPass
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return domain.DoctorAuthResponse{}, domain.ErrDoctorInvalidPass
	}

	token, err := jwt.GenerateToken(doctor.ID, doctor.Email, "doctor",s.cfg.JWTSecret, s.cfg.JWTExpiryHours)

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("generating token: %w", err)
	}

	return domain.DoctorAuthResponse{
		Doctor:      doctor,
		AccessToken: token,
	}, nil
}


func (s *DoctorService) GetMe(ctx context.Context, doctorID string) (domain.Doctor, error){
	doctor, err := s.repo.FindByID(ctx, doctorID)

	if err != nil {
		return domain.Doctor{}, fmt.Errorf("doctor service get me: %w", err)
	}

	return doctor, nil
}

func (s *DoctorService) GetAllDoctors(ctx context.Context, specialty string) ([]domain.Doctor, error) {
	doctors, err := s.repo.FindAll(ctx, specialty)

	if err != nil {
		return nil, fmt.Errorf("doctor service get all doctors: %w", err)
	}

	return doctors, nil	
}