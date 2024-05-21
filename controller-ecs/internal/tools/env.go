package tools

import (
	"log"
	"os"
)

var requiredEnvVars = []string{
	//"GITHUB_PAT",
	//"REPO_NAME",
	//"AWS_ACCESS_KEY_ID",
	//"AWS_SECRET_ACCESS_KEY",
	//"AWS_REGION",
}

func CheckEnvVars() {
	for _, envVar := range requiredEnvVars {
		if _, ok := os.LookupEnv(envVar); !ok {
			log.Fatalf("Environment variable %s is required", envVar)
		}
	}
}
