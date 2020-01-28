package util

import (
	"crypto/rand"
	"math/big"
)

var (
	runes        = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	runes_length = len(runes)
)

func NewRandID(n int) (string, error) {
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

func NewSessionID() (string, error) {
	return NewRandID(24)
}

func NewCSRFToken() (string, error) {
	return NewRandID(24)
}
