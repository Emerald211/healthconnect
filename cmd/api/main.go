package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Emerald211/healthconnect/internal/config"
	"github.com/Emerald211/healthconnect/internal/db"
	"github.com/Emerald211/healthconnect/internal/handler"
	"github.com/Emerald211/healthconnect/internal/repository"
	"github.com/Emerald211/healthconnect/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {

	// load config first
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// connect to DB

	pool, err := db.NewPool(cfg)

	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	defer pool.Close()

	fmt.Println("Database connected successfully")

	patientRepo := repository.NewPatientRepository(pool)
	patientService := service.NewPatientService(patientRepo, cfg)
	patientHandler := handler.NewPatientHandler(patientService)

	// if in production mode, set gin to release mode
	if cfg.ServerEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin Router
	r := gin.New()

	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

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
		}

	
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
		fmt.Printf("HealthConnect API running on :%s (env: %s)\n",
			cfg.ServerPort, cfg.ServerEnv)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}

	}()

	// Graceful Shutdown

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	fmt.Println("Server stopped cleanly")

}



