package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateOTP returns a cryptographically random numeric code of the given length.
func GenerateOTP(totalDigits int) (string, error) {
	otp := ""
	for i := 0; i < totalDigits; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp += fmt.Sprintf("%d", num.Int64())
	}
	return otp, nil
}
