package config

import "os"

// Get Env
func GetEnv(key, fallback string) string {
	e := os.Getenv(key)
	if e == "" {
		return fallback
	}
	return e
}
