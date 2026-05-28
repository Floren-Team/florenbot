package helpers

import (
	"crypto/rand"
	"math/big"
)

func GenerateCode() string {
	const alphaNumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 32)

	for i := 0; i < 32; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphaNumeric))))
		if err != nil {
			return ""
		}
		result[i] = alphaNumeric[num.Int64()]
	}

	return string(result)
}
