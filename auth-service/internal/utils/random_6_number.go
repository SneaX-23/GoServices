package utils

import (
	"crypto/rand"
	"math/big"
)

func SecureRandom6Digit() (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return 0, err
	}

	// retutn 6 digit number
	return n.Int64() + 100000, nil
}
