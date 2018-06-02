package util

import (
	"math/rand"
	"time"
)

// From: http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyz")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
