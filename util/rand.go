package util

import (
	"crypto/rand"
	"math/big"
)

var (
	runes        = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	runes_length = len(runes)
)

func NewRandId(n int) (string, error) {
	data := make([]rune, n)
	for i := range data {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(runes_length)))
		if err != nil {
			return "", err
		}
		data[i] = runes[num.Int64()]
	}
	return string(data), nil
}

func NewSessionId() (string, error) {
	return NewRandId(24)
}

func NewCSRFToken() (string, error) {
	return NewRandId(24)
}
