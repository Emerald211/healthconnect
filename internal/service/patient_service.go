package service

import (
	"context"
	"fmt"

	"github.com/Emerald211/healthconnect/internal/config"
	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/Emerald211/healthconnect/internal/repository"
	"github.com/Emerald211/healthconnect/internal/store"
	"github.com/Emerald211/healthconnect/pkg/jwt"
	"github.com/Emerald211/healthconnect/pkg/otp"

	"golang.org/x/crypto/bcrypt"
)

type PatientService struct {
	repo         *repository.PatientRepository
	cfg          *config.Config
	tokenStore   *store.TokenStore
	emailService *EmailService
}

func NewPatientService(repo *repository.PatientRepository, cfg *config.Config, tokenStore *store.TokenStore, emailService *EmailService) *PatientService {
	return &PatientService{repo: repo, cfg: cfg, tokenStore: tokenStore, emailService: emailService}
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

	token, err := jwt.GenerateToken(newUser.ID, newUser.Email, "patient", s.cfg.JWTSecret, s.cfg.JWTExpiryMinutes)
	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("generating token: %w", err)
	}

	refreshToken, err := jwt.GenerateRefreshToken()

	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("generating refresh token: %w", err)
	}

	if err := s.tokenStore.SaveRefreshToken(ctx, newUser.ID, refreshToken, s.cfg.JWTRefreshDays); err != nil {
		return domain.AuthResponse{}, fmt.Errorf("saving refresh token: %w", err)
	}

	go func() {
		code, err := otp.Generate()
		if err != nil {
			return
		}
		s.tokenStore.SaveOTP(context.Background(), "verify_email", newUser.Email, code)
		s.emailService.SendOTP(newUser.Email, newUser.Name, code, "verify_email")
	}()

	return domain.AuthResponse{
		Patient:      newUser,
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil

}

func (s *PatientService) LoginPatient(ctx context.Context, req domain.LoginDto) (domain.AuthResponse, error) {

	attempts, err := s.tokenStore.IncrementLoginAttempts(ctx, req.Email)
	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("checking login attempts: %w", err)
	}
	if attempts > 5 {
		return domain.AuthResponse{}, domain.ErrTooManyAttempts
	}

	patient, hashedPassword, err := s.repo.FindByEmail(ctx, req.Email)

	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("login patient: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return domain.AuthResponse{}, domain.ErrInvalidPassword

	}

	s.tokenStore.ResetLoginAttempts(ctx, req.Email)

	token, err := jwt.GenerateToken(patient.ID, patient.Email, "patient", s.cfg.JWTSecret, s.cfg.JWTExpiryMinutes)
	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("generating token: %w", err)
	}

	refreshToken, err := jwt.GenerateRefreshToken()

	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("generating refresh token: %w", err)
	}

	if err := s.tokenStore.SaveRefreshToken(ctx, patient.ID, refreshToken, s.cfg.JWTRefreshDays); err != nil {
		return domain.AuthResponse{}, fmt.Errorf("saving refresh token: %w", err)
	}

	return domain.AuthResponse{
		Patient:      patient,
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *PatientService) GetMe(ctx context.Context, patientID string) (domain.Patient, error) {
	patient, _, err := s.repo.FindByID(ctx, patientID)

	if err != nil {
		return domain.Patient{}, fmt.Errorf("patient service get me: %w", err)
	}

	return patient, nil
}

func (s *PatientService) RefreshToken(ctx context.Context, req domain.RefreshTokenRequest) (domain.AuthResponse, error) {

	storedToken, err := s.tokenStore.GetRefreshToken(ctx, req.UserID)
	if err != nil {
		return domain.AuthResponse{}, domain.ErrInvalidToken
	}

	if storedToken != req.RefreshToken {
		return domain.AuthResponse{}, domain.ErrInvalidToken
	}

	patient, _, err := s.repo.FindByID(ctx, req.UserID)
	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("finding patient: %w", err)
	}

	accessToken, err := jwt.GenerateToken(patient.ID, patient.Email, "patient", s.cfg.JWTSecret, s.cfg.JWTExpiryMinutes)
	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("generating access token: %w", err)
	}

	newRefreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf("generating refresh token: %w", err)
	}

	if err := s.tokenStore.SaveRefreshToken(ctx, patient.ID, newRefreshToken, s.cfg.JWTRefreshDays); err != nil {
		return domain.AuthResponse{}, fmt.Errorf("saving refresh token: %w", err)
	}

	return domain.AuthResponse{
		Patient:      patient,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *PatientService) SendVerificationOTP(ctx context.Context, email string) error {
	patient, _, err := s.repo.FindByEmail(ctx, email)

	if err != nil {
		return domain.ErrPatientNotFound
	}

	if patient.IsEmailVerified {
		return domain.ErrAlreadyVerified
	}

	otpCode, err := otp.Generate()

	if err != nil {
		return fmt.Errorf("generating otp: %w", err)
	}

	if err := s.tokenStore.SaveOTP(ctx, "verify_email", email, otpCode); err != nil {
		return fmt.Errorf("saving otp: %w", err)
	}

	if err := s.emailService.SendOTP(email, patient.Name, otpCode, "verify_email"); err != nil {
		return fmt.Errorf("sending otp: %w", err)
	}

	return nil
}

func (s *PatientService) VerifyEmail(ctx context.Context, req domain.VerifyEmailRequest) error {
	storedOTP, err := s.tokenStore.GetOTP(ctx, "verify_email", req.Email)

	if err != nil {
		return domain.ErrInvalidOTP
	}

	if storedOTP != req.OTP {
		return domain.ErrInvalidOTP
	}

	if err := s.repo.MarkEmailVerified(ctx, req.Email); err != nil {
		return fmt.Errorf("marking email verified: %w", err)
	}

	if err := s.tokenStore.DeleteOTP(ctx, "verify_email", req.Email); err != nil {
		return fmt.Errorf("deleting otp: %w", err)
	}

	return nil
}

func (s *PatientService) ForgotPassword(ctx context.Context, req domain.ForgotPasswordRequest) error {
	patient, _, err := s.repo.FindByEmail(ctx, req.Email)

	if err != nil {
		return domain.ErrPatientNotFound
	}

	code, err := otp.Generate()

	if err != nil {
		return fmt.Errorf("generating otp: %w", err)
	}

	if err := s.tokenStore.SaveOTP(ctx, "reset_password", req.Email, code); err != nil {
		return fmt.Errorf("saving otp: %w", err)
	}

	if err := s.emailService.SendOTP(req.Email, patient.Name, code, "reset_password"); err != nil {
		return fmt.Errorf("sending otp: %w", err)
	}

	return nil
}

func (s *PatientService) ResetPassword(ctx context.Context, req domain.ResetPasswordRequest) error {
	storedOTP, err := s.tokenStore.GetOTP(ctx, "reset_password", req.Email)

	if err != nil {
		return domain.ErrInvalidOTP
	}

	if storedOTP != req.OTP {
		return domain.ErrInvalidOTP
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)

	if err := s.repo.UpdatePassword(ctx, req.Email, string(hashedPassword)); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	if err := s.tokenStore.DeleteOTP(ctx, "reset_password", req.Email); err != nil {
		return fmt.Errorf("deleting otp: %w", err)
	}

	patient, _, _ := s.repo.FindByEmail(ctx, req.Email)
	s.tokenStore.DeleteRefreshToken(ctx, patient.ID)

	return nil
}

func (s *PatientService) ChangePassword(ctx context.Context, patientID string, req domain.ChangePasswordRequest) error {
	patient, hashedPassword, err := s.repo.FindByID(ctx, patientID)
	if err != nil {
		return domain.ErrPatientNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.CurrentPassword)); err != nil {
		return domain.ErrInvalidPassword
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, patient.Email, string(hashedNewPassword)); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	s.tokenStore.DeleteRefreshToken(ctx, patientID)

	return nil

}

func (s *PatientService) Logout(ctx context.Context, patientID string) error {
	if err := s.tokenStore.DeleteRefreshToken(ctx, patientID); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}
