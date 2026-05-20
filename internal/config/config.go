package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort string
	ServerEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret        string
	JWTExpiryMinutes int
	JWTRefreshDays   int

	RedisHost     string
	RedisPort     string
	RedisPassword string

	ResendAPIKey string
	EmailFrom    string

	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("No .env file found, reading from environment variables")
	}

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_ENV", "development")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("JWT_REFRESH_DAYS", 7)

	cfg := &Config{
		ServerPort:         viper.GetString("SERVER_PORT"),
		ServerEnv:          viper.GetString("SERVER_ENV"),
		DBHost:             viper.GetString("DB_HOST"),
		DBPort:             viper.GetString("DB_PORT"),
		DBUser:             viper.GetString("DB_USER"),
		DBPassword:         viper.GetString("DB_PASSWORD"),
		DBName:             viper.GetString("DB_NAME"),
		JWTSecret:          viper.GetString("JWT_SECRET"),
		JWTExpiryMinutes:   viper.GetInt("JWT_EXPIRY_MINUTES"),
		JWTRefreshDays:     viper.GetInt("JWT_REFRESH_DAYS"),
		RedisHost:          viper.GetString("REDIS_HOST"),
		RedisPort:          viper.GetString("REDIS_PORT"),
		RedisPassword:      viper.GetString("REDIS_PASSWORD"),
		ResendAPIKey:       viper.GetString("RESEND_API_KEY"),
		EmailFrom:          viper.GetString("EMAIL_FROM"),
		CORSAllowedOrigins: strings.Split(viper.GetString("CORS_ALLOWED_ORIGINS"), ","),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"JWT_SECRET": c.JWTSecret,
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	return nil
}
