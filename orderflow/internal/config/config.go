package config

import (
	// "log"
	"os"
	// "strconv"
)

type Config struct {
	Port		string
	DBHost		string
	DBPort		string
	DBUser		string
	DBPassword	string
	DBName		string
	DBSSLMode	string 
}

func Load() *Config {
	return &Config{
		Port:		getEnv("PORT", "8080"),
		DBHost:		getEnv("DB_HOST", "localhost"),
		DBPort:		getEnv("DB_PORT", "5432"),
		DBUser:		getEnv("DB_USER", "admin"),
		DBPassword:	getEnv("DB_PASSWORD", "admin123"),
		DBName:		getEnv("DB_NAME", "orderflow_db"), 
		DBSSLMode:	getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v 
	}
	return fallback
}
