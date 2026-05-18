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

type DoctorService struct {
	repo         *repository.DoctorRepository
	cfg          *config.Config
	tokenStore   *store.TokenStore
	emailService *EmailService
}

func NewDoctorService(repo *repository.DoctorRepository, cfg *config.Config, tokenStore *store.TokenStore, emailService *EmailService) *DoctorService {
	return &DoctorService{
		repo:         repo,
		cfg:          cfg,
		tokenStore:   tokenStore,
		emailService: emailService,
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

	token, err := jwt.GenerateToken(newDoctor.ID, newDoctor.Email, "doctor", s.cfg.JWTSecret, s.cfg.JWTExpiryMinutes)

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("generating token: %w", err)
	}

	refreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("generating refresh token: %w", err)
	}

	if err := s.tokenStore.SaveRefreshToken(ctx, newDoctor.ID, refreshToken, s.cfg.JWTExpiryMinutes); err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("saving refresh token: %w", err)
	}

	go func() {
		code, err := otp.Generate()
		if err != nil {
			return
		}
		s.tokenStore.SaveOTP(context.Background(), "verify_email", newDoctor.Email, code)
		s.emailService.SendOTP(newDoctor.Email, newDoctor.Name, code, "verify_email")
	}()

	return domain.DoctorAuthResponse{
		Doctor:       newDoctor,
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil

}

func (s *DoctorService) Login(ctx context.Context, req domain.LoginDto) (domain.DoctorAuthResponse, error) {

	attempts, err := s.tokenStore.IncrementLoginAttempts(ctx, req.Email)

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("checking login attempts: %w", err)
	}

	if attempts > 5 {
		return domain.DoctorAuthResponse{}, domain.ErrTooManyAttempts
	}

	doctor, hashedPassword, err := s.repo.FindByEmail(ctx, req.Email)

	if err != nil {
		return domain.DoctorAuthResponse{}, domain.ErrDoctorInvalidPass
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return domain.DoctorAuthResponse{}, domain.ErrDoctorInvalidPass
	}

	token, err := jwt.GenerateToken(doctor.ID, doctor.Email, "doctor", s.cfg.JWTSecret, s.cfg.JWTExpiryMinutes)

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("generating token: %w", err)
	}

	refreshToken, err := jwt.GenerateRefreshToken()

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("generating refresh token: %w", err)
	}

	if err := s.tokenStore.SaveRefreshToken(ctx, doctor.ID, refreshToken, s.cfg.JWTExpiryMinutes); err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("saving refresh token: %w", err)
	}

	return domain.DoctorAuthResponse{
		Doctor:       doctor,
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *DoctorService) RefreshToken(ctx context.Context, req domain.RefreshTokenRequest) (domain.DoctorAuthResponse, error) {
	storedToken, err := s.tokenStore.GetRefreshToken(ctx, req.UserID)

	if err != nil {
		return domain.DoctorAuthResponse{}, domain.ErrInvalidToken
	}

	if storedToken != req.RefreshToken {
		return domain.DoctorAuthResponse{}, domain.ErrInvalidToken
	}

	doctor, _, err := s.repo.FindByID(ctx, req.UserID)

	if err != nil {
		return domain.DoctorAuthResponse{}, domain.ErrDoctorNotFound
	}

	newAccessToken, err := jwt.GenerateToken(doctor.ID, doctor.Email, "doctor", s.cfg.JWTSecret, s.cfg.JWTExpiryMinutes)

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("generating access token: %w", err)
	}

	newRefreshToken, err := jwt.GenerateRefreshToken()

	if err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("getting refresh token: %w", err)
	}

	if err := s.tokenStore.SaveRefreshToken(ctx, doctor.ID, newRefreshToken, s.cfg.JWTRefreshDays); err != nil {
		return domain.DoctorAuthResponse{}, fmt.Errorf("saving refresh token: %w", err)
	}

	return domain.DoctorAuthResponse{
		Doctor:       doctor,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *DoctorService) SendVerificationEmail(ctx context.Context, email string) error {
	doctor, _, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return domain.ErrDoctorNotFound
	}

	if doctor.IsEmailVerified {
		return domain.ErrAlreadyVerified
	}

	code, err := otp.Generate()
	if err != nil {
		return fmt.Errorf("generating otp: %w", err)
	}

	if err := s.tokenStore.SaveOTP(ctx, "verify_email", doctor.Email, code); err != nil {
		return fmt.Errorf("saving otp: %w", err)
	}

	if err := s.emailService.SendOTP(doctor.Email, doctor.Name, code, "verify_email"); err != nil {
		return fmt.Errorf("sending otp: %w", err)
	}

	return nil
}

func (s *DoctorService) VerifyEmail(ctx context.Context, req domain.VerifyEmailRequest) error {
	storedOTP, err := s.tokenStore.GetOTP(ctx, "verify-email", req.Email)

	if err != nil {
		return domain.ErrInvalidOTP
	}

	if storedOTP != req.OTP {
		return domain.ErrInvalidOTP
	}

	if err := s.repo.MarkEmailVerified(ctx, req.Email); err != nil {
		return fmt.Errorf("marking email verified: %w", err)
	}

	s.tokenStore.DeleteOTP(ctx, "verify-email", req.Email)

	return nil
}

func (s *DoctorService) ForgotPassword(ctx context.Context, req domain.ForgotPasswordRequest) error {
	doctor, _, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil
	}

	code, err := otp.Generate()
	if err != nil {
		return fmt.Errorf("generating otp: %w", err)
	}

	if err := s.tokenStore.SaveOTP(ctx, "reset_password", req.Email, code); err != nil {
		return fmt.Errorf("saving otp: %w", err)
	}

	if err := s.emailService.SendOTP(req.Email, doctor.Name, code, "reset_password"); err != nil {
		return fmt.Errorf("sending reset email: %w", err)
	}

	return nil
}

func (s *DoctorService) ResetPassword(ctx context.Context, req domain.ResetPasswordRequest) error {
	storedOTP, err := s.tokenStore.GetOTP(ctx, "reset_password", req.Email)
	if err != nil {
		return domain.ErrInvalidOTP
	}

	if storedOTP != req.OTP {
		return domain.ErrInvalidOTP
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, req.Email, string(hashedPassword)); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	s.tokenStore.DeleteOTP(ctx, "reset_password", req.Email)

	doctor, _, _ := s.repo.FindByEmail(ctx, req.Email)
	s.tokenStore.DeleteRefreshToken(ctx, doctor.ID)

	return nil
}

func (s *DoctorService) ChangePassword(ctx context.Context, doctorID string, req domain.ChangePasswordRequest) error {
	doctor, hashedPassword, err := s.repo.FindByID(ctx, doctorID)
	if err != nil {
		return domain.ErrDoctorNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.CurrentPassword)); err != nil {
		return domain.ErrDoctorInvalidPass
	}

	newHashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, doctor.Email, string(newHashed)); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	s.tokenStore.DeleteRefreshToken(ctx, doctorID)

	return nil
}

func (s *DoctorService) Logout(ctx context.Context, doctorID string) error {
	if err := s.tokenStore.DeleteRefreshToken(ctx, doctorID); err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	return nil
}

func (s *DoctorService) GetMe(ctx context.Context, doctorID string) (domain.Doctor, error) {
	doctor, _, err := s.repo.FindByID(ctx, doctorID)

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
