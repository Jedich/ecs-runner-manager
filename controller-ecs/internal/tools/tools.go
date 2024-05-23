package tools

import (
	"log"
	"math/rand"
	"os"
)

var requiredEnvVars = []string{
	"GITHUB_PAT",
	"REPO",
}

func CheckEnvVars() {
	for _, envVar := range requiredEnvVars {
		if _, ok := os.LookupEnv(envVar); !ok {
			log.Fatalf("Environment variable %s is required", envVar)
		}
	}
}

const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
