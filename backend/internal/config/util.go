package config

import "os"

// Getenv retrieves an environment variable, returning the given fallback if not found
func Getenv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}
