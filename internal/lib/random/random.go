package random

import (
	"math/rand"
	"time"
)

func GenerateRandomString(length int) string {
	rnd := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	symbols := []rune("QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnm1234567890")

	bytes := make([]rune, length)
	for i := range bytes {
		bytes[i] = symbols[rnd.Intn(len(symbols))]
	}

	return string(bytes)
}
