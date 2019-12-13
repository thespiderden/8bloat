package util

import (
	"math/rand"
)

var (
	runes        = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	runes_length = len(runes)
)

func NewRandId(n int) string {
	data := make([]rune, n)
	for i := range data {
		data[i] = runes[rand.Intn(runes_length)]
	}
	return string(data)
}

func NewSessionId() string {
	return NewRandId(24)
}
