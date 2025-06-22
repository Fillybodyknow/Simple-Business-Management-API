package utility

import (
	"math/rand"
	"time"
)

func GenerateTrackingNumber() string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())

	code := make([]byte, 6)
	for i := range code {
		code[i] = letters[rand.Intn(len(letters))]
	}

	return "TH" + string(code)
}
