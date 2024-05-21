package tools

import (
	"log"
	"os"
)

var requiredEnvVars = []string{
	//"GITHUB_PAT",
	//"REPO_NAME",
}

func CheckEnvVars() {
	for _, envVar := range requiredEnvVars {
		if _, ok := os.LookupEnv(envVar); !ok {
			log.Fatalf("Environment variable %s is required", envVar)
		}
	}
}
