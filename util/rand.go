package util

import (
	"crypto/rand"
	"encoding/base64"
)

var enc = base64.URLEncoding

func NewRandID(n int) (string, error) {
	data := make([]byte, enc.DecodedLen(n))
	_, err := rand.Read(data)
	if err != nil {
		return "", err
	}
	return enc.EncodeToString(data), nil
}

func NewSessionID() (string, error) {
	return NewRandID(24)
}

func NewCSRFToken() (string, error) {
	return NewRandID(24)
}
