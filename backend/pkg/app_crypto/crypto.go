package app_crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
	"io"
)

func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func Verify(hashed, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

// GenerateAPIKey generates a new API key of the given length in bytes.
func GenerateAPIKey(userID string, secretKey string, length int) (string, error) {
	key := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}

	finalKey := bytes.Join([][]byte{[]byte(userID), key}, []byte("."))

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(finalKey)
	return hex.EncodeToString(h.Sum(nil)), nil
}
