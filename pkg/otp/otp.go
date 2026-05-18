package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func Generate() (string, error) {
	max := big.NewInt(999999)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("generating otp: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
