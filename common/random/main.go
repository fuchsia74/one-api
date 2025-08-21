package random

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

// GetUUID generates a UUID and returns it as a string without hyphens.
// It uses github.com/google/uuid for UUID generation.
func GetUUID() string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	return code
}

const keyChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const keyNumbers = "0123456789"

// GenerateKey creates a 48-character key consisting of 16 random characters
// followed by a modified UUID. This provides a unique, secure identifier
// suitable for authentication tokens or similar purposes.
func GenerateKey() string {
	key := make([]byte, 48)
	for i := range 16 {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(keyChars))))
		if err != nil {
			// This is unlikely to result in an error, especially on Linux, so it's safe to keep as is.
			panic(err)
		}
		key[i] = keyChars[n.Int64()]
	}
	uuid := GetUUID()
	for i := range 32 {
		c := uuid[i]
		if i%2 == 0 && c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}
		key[i+16] = c
	}
	return string(key)
}

// GetRandomString generates a random string of the specified length
// using a mix of numbers and letters (both uppercase and lowercase).
// It uses crypto/rand for secure random number generation.
func GetRandomString(length int) string {
	key := make([]byte, length)
	for i := range length {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(keyChars))))
		if err != nil {
			// This is unlikely to result in an error, especially on Linux, so it's safe to keep as is.
			panic(err)
		}
		key[i] = keyChars[n.Int64()]
	}
	return string(key)
}

// GetRandomNumberString generates a random string of the specified length
// using only numeric characters (0-9). It uses crypto/rand for secure
// random number generation.
func GetRandomNumberString(length int) string {
	key := make([]byte, length)
	for i := range length {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(keyNumbers))))
		if err != nil {
			// This is unlikely to result in an error, especially on Linux, so it's safe to keep as is.
			panic(err)
		}
		key[i] = keyNumbers[n.Int64()]
	}
	return string(key)
}

// RandRange returns a random number between min and max (max is not included)
func RandRange(min, max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		// This is unlikely to result in an error, especially on Linux, so it's safe to keep as is.
		panic(err)
	}
	return min + int(n.Int64())
}
