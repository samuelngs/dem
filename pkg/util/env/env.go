package env

import "os"

// Has checks if the variable present in the shell enviroment
func Has(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}

// GetEnvAsString reads environment variable as string
func GetEnvAsString(key string, defaultValue string) string {
	if s := os.Getenv(key); len(s) > 0 {
		return s
	}
	return defaultValue
}

// GetEnvAsStringWithFallback reads environment variable as string with fallback
func GetEnvAsStringWithFallback(key string, fallbackKey string) string {
	if s := os.Getenv(key); len(s) > 0 {
		return s
	}
	return os.Getenv(fallbackKey)
}
