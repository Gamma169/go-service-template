package main

import (
	"log"
	"os"
)

func getRequiredEnv(envVar string) string {
	val, found := os.LookupEnv(envVar)
	if !found || val == "" {
		panic("PLEASE SET " + envVar + " ENVIRONMENT VARIABLE")
	}
	return val
}

func getOptionalEnv(envVar string, defaultVal string) string {
	val, found := os.LookupEnv(envVar)
	if !found || val == "" {
		log.Printf("Env var: '%s' not found or empty.  Setting to default value: '%s'", envVar, defaultVal)
		return defaultVal
	}
	return val
}

func checkRequiredEnvs() {
	// Not sure if I should use the getOptionalEnv function here or just os.LookupEnv
	// Because if I use getOptionalEnv and it doesn't exist, we output the logs for it twice
	// I think that's fine, but I need to think on it
	if getOptionalEnv("DATABASE_URL", "") == "" {
		getRequiredEnv("DATABASE_NAME")
		getRequiredEnv("DATABASE_HOST")
		getRequiredEnv("DATABASE_USER")
	}
}

func initDebug() {
	if releaseMode = getOptionalEnv("RELEASE_MODE", "dev"); releaseMode != "production" {
		debug = true
	} else {
		debug = false
	}
}
