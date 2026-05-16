package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenStore struct {
	redis *redis.Client
}

func NewTokenStore(redis *redis.Client) *TokenStore {
	return &TokenStore{redis: redis}
}

func (ts *TokenStore) SaveRefreshToken(ctx context.Context, userID, token string, days int) error {
	key := fmt.Sprintf("refresh:%s", userID)
	ttl := time.Duration(days) * 24 * time.Hour

	if err := ts.redis.Set(ctx, key, token, ttl).Err(); err != nil {
		return fmt.Errorf("saving refresh token: %w", err)
	}
	return nil
}

func (ts *TokenStore) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("refresh:%s", userID)
	token, err := ts.redis.Get(ctx, key).Result()

	if err != nil {
		return "", fmt.Errorf("getting refresh token: %w", err)
	}
	return token, nil
}

func (ts *TokenStore) DeleteRefreshToken(ctx context.Context, userID string) error {
	key := fmt.Sprintf("refresh:%s", userID)
	if err := ts.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("deleting refresh token: %w", err)
	}
	return nil
}

func (ts *TokenStore) SaveOTP(ctx context.Context, purpose, email string, otp string) error {
	key := fmt.Sprintf("otp:%s:%s", purpose, email)
	ttl := 10 * time.Minute

	if err := ts.redis.Set(ctx, key, otp, ttl).Err(); err != nil {
		return fmt.Errorf("saving otp: %w", err)
	}

	return nil
}

func (ts *TokenStore) GetOTP(ctx context.Context, purpose, email string) (string, error) {
	key := fmt.Sprintf("otp:%s:%s", purpose, email)

	otp, err := ts.redis.Get(ctx, key).Result()

	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("otp expired or not found")
		}
		return "", fmt.Errorf("getting otp: %w", err)
	}

	return otp, nil
}

func (s *TokenStore) DeleteOTP(ctx context.Context, purpose, email string) error {
	key := fmt.Sprintf("otp:%s:%s", purpose, email)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("deleting otp: %w", err)
	}
	return nil
}

func (s *TokenStore) IncrementLoginAttempts(ctx context.Context, email string) (int64, error) {
	key := fmt.Sprintf("login_attempts:%s", email)
	count, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("incrementing login attempts: %w", err)
	}

	if count == 1 {
		s.redis.Expire(ctx, key, 15*time.Minute)
	}
	return count, nil
}

func (s *TokenStore) ResetLoginAttempts(ctx context.Context, email string) error {
	key := fmt.Sprintf("login_attempts:%s", email)
	return s.redis.Del(ctx, key).Err()
}
