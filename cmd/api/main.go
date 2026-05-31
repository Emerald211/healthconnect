package main

import (
	"context"

	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Emerald211/healthconnect/internal/config"
	"github.com/Emerald211/healthconnect/internal/db"
	"github.com/Emerald211/healthconnect/internal/handler"
	"github.com/Emerald211/healthconnect/internal/middleware"
	"github.com/Emerald211/healthconnect/internal/repository"
	"github.com/Emerald211/healthconnect/internal/service"
	"github.com/Emerald211/healthconnect/internal/store"
	"github.com/Emerald211/healthconnect/pkg/logger"
	"github.com/gin-gonic/gin"

	_ "github.com/Emerald211/healthconnect/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           HealthConnect API
// @version         1.0
// @description     A production-grade healthcare API for Nigeria — appointment booking, medical records, and telemedicine
// @termsOfService  http://swagger.io/terms/

// @contact.name   Emerald
// @contact.email  your-email@gmail.com

// @license.name  MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token

func main() {

	// load config first
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// init logger
	logger.Init(cfg.ServerEnv)

	// connect to DB

	pool, err := db.NewPool(cfg)

	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	defer pool.Close()

	slog.Info("Database connected successfully")

	redisClient, err := db.NewRedisClient(cfg)

	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	defer redisClient.Close()

	slog.Info("Redis connected successfully")

	tokenStore := store.NewTokenStore(redisClient)
	emailService := service.NewEmailService(cfg)

	patientRepo := repository.NewPatientRepository(pool)
	patientService := service.NewPatientService(patientRepo, cfg, tokenStore, emailService)
	patientHandler := handler.NewPatientHandler(patientService)

	doctorRepo := repository.NewDoctorRepository(pool)
	doctorService := service.NewDoctorService(doctorRepo, cfg, tokenStore, emailService)
	doctorHandler := handler.NewDoctorHandler(doctorService)

	// Appointment layers
	appointmentRepo := repository.NewAppointmentRepository(pool)
	appointmentService := service.NewAppointmentService(appointmentRepo, doctorRepo)
	appointmentHandler := handler.NewAppointmentHandler(appointmentService)

	// payment layers
	paymentRepo := repository.NewPaymentRepository(pool)
	paymentService := service.NewPaymentService(paymentRepo, appointmentRepo, patientRepo, cfg, emailService)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// if in production mode, set gin to release mode
	if cfg.ServerEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin Router
	r := gin.New()

	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.GinLogger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RequestSizeLimit(10 << 20))
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database Unreachable",
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"service":  "HealthConnect",
			"env":      cfg.ServerEnv,
			"database": "connected",
			"time":     time.Now().UTC(),
		})
	})

	v1 := r.Group("/api/v1")
	{

		auth := v1.Group("/auth")

		{
			auth.POST("/register", patientHandler.Register)
			auth.POST("/login", patientHandler.Login)
			auth.POST("/refresh", patientHandler.RefreshToken)
			auth.POST("/verify-email", patientHandler.VerifyEmail)
			auth.POST("/resend-otp", patientHandler.ResendOTP)
			auth.POST("/forgot-password", patientHandler.ForgotPassword)
			auth.POST("/reset-password", patientHandler.ResetPassword)

		}

		doctorAuth := v1.Group("/doctors/auth")
		{
			doctorAuth.POST("/register", doctorHandler.RegisterDoctor)
			doctorAuth.POST("/login", doctorHandler.Login)
			doctorAuth.POST("/refresh", doctorHandler.RefreshToken)
			doctorAuth.POST("/verify-email", doctorHandler.VerifyEmail)
			doctorAuth.POST("/resend-otp", doctorHandler.ResendOTP)
			doctorAuth.POST("/forgot-password", doctorHandler.ForgotPassword)
			doctorAuth.POST("/reset-password", doctorHandler.ResetPassword)
		}

		patients := v1.Group("/patients")
		patients.Use(middleware.AuthMiddleware(cfg))
		patients.Use(middleware.RoleMiddleware("patient"))

		{
			patients.GET("/me", patientHandler.GetMe)
			patients.POST("/logout", patientHandler.Logout)
			patients.POST("/change-password", patientHandler.ChangePassword)
		}

		doctors := v1.Group("/doctors")
		doctors.Use(middleware.AuthMiddleware(cfg))
		doctors.Use(middleware.RoleMiddleware("doctor"))

		{
			doctors.GET("/me", doctorHandler.GetMe)
			doctors.POST("/logout", doctorHandler.Logout)
			doctors.POST("/change-password", doctorHandler.ChangePassword)
		}

		doctorList := v1.Group("/doctors")
		doctorList.Use(middleware.AuthMiddleware(cfg))
		{
			doctorList.GET("", doctorHandler.GetAll)
		}

		// Appointment routes
		appointments := v1.Group("/appointments")
		appointments.Use(middleware.AuthMiddleware(cfg))
		{
			// Any authenticated user can view slots and availability
			appointments.GET("/slots/:doctor_id", appointmentHandler.GetAvailableSlots)
			appointments.GET("/availability/:doctor_id", appointmentHandler.GetAvailability)

			// Patient only routes
			patientAppts := appointments.Group("")
			patientAppts.Use(middleware.RoleMiddleware("patient"))
			{
				patientAppts.POST("", appointmentHandler.BookAppointment)
				patientAppts.GET("/my", appointmentHandler.GetPatientAppointments)
			}

			// Doctor only routes
			doctorAppts := appointments.Group("")
			doctorAppts.Use(middleware.RoleMiddleware("doctor"))
			{
				doctorAppts.POST("/availability", appointmentHandler.SetAvailability)
				doctorAppts.POST("/slots/:doctor_id/generate", appointmentHandler.GenerateSlots)
				doctorAppts.GET("/doctor/my", appointmentHandler.GetDoctorAppointments)
				doctorAppts.PATCH("/:id/status", appointmentHandler.UpdateAppointmentStatus)
			}
		}

		payments := v1.Group("/payments")
		payments.Use(middleware.AuthMiddleware(cfg))
		payments.Use(middleware.RoleMiddleware("patient"))
		{
			payments.POST("/initialize", paymentHandler.InitiliazePayment)
			payments.GET("/:appointment_id", paymentHandler.GetPaymentStatus)
		}

		v1.POST("/payments/webhook", paymentHandler.HandleWebhook)

	}

	server := &http.Server{

		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in go routine

	go func() {
		slog.Info("HealthConnect API starting",
			"port", cfg.ServerPort,
			"env", cfg.ServerEnv)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}

	}()

	// Graceful Shutdown

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped cleanly")

}
