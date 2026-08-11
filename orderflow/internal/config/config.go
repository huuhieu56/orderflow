package config

import (
	// "log"
	"os"
	"time"
	// "strconv"
)

type Config struct {
	Port			string

	DBHost			string
	DBPort			string
	DBUser			string
	DBPassword		string
	DBName			string
	DBSSLMode		string 

	JWTSecret		string 
	JWTExpiration	time.Duration
	JWTRefreshExpiration	time.Duration
}

func Load() *Config {
	return &Config{
		Port:			getEnv("PORT", "8080"),

		DBHost:			getEnv("DB_HOST", "localhost"),
		DBPort:			getEnv("DB_PORT", "5432"),
		DBUser:			getEnv("DB_USER", "admin"),
		DBPassword:		getEnv("DB_PASSWORD", "admin123"),
		DBName:			getEnv("DB_NAME", "orderflow_db"), 
		DBSSLMode:		getEnv("DB_SSLMODE", "disable"),

		JWTSecret: 		getEnv("JWT_SECRET", "orderflow-dev-secret"),
		JWTExpiration:	time.Hour, 
		JWTRefreshExpiration: time.Hour * 24 * 7,
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v 
	}
	return fallback
}
