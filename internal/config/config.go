package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ApiPort               string
	Port                  string
	DatabaseDSN           string
	JWTAccessSecret       string
	JWTRefreshSecret      string
	APIBaseURL            string
	FirebaseCredsFile     string
	FirebaseCredsJSON     string
	OpenAIKey             string
	RedisAddr             string
	AppleAppSharedSecret  string
	AppleBundleID         string
	AppleTeamID           string
	DailyMessageLimitFree int
	CronAuthKey           string
	EnableDevRoutes       bool
}

func LoadConfig() (*Config, error) {
	apiPort := getEnv("PORT", "")
	if apiPort == "" {
		apiPort = getEnv("API_PORT", "8164")
	}
	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")
	postgresSSLMode := os.Getenv("POSTGRES_SSL_MODE")

	databaseDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		postgresHost,
		postgresPort,
		postgresUser,
		postgresPassword,
		postgresDB,
		postgresSSLMode,
	)
	config := &Config{
		ApiPort:               apiPort,
		Port:                  postgresPort,
		DatabaseDSN:           databaseDSN,
		JWTAccessSecret:       os.Getenv("ACCESS_SECRET"),
		JWTRefreshSecret:      os.Getenv("REFRESH_SECRET"),
		FirebaseCredsFile:     os.Getenv("FIREBASE_CREDS_FILE"),
		FirebaseCredsJSON:     os.Getenv("FIREBASE_CREDS_JSON"),
		OpenAIKey:             os.Getenv("OPENAI_KEY"),
		RedisAddr:             os.Getenv("REDIS_ADDR"),
		AppleAppSharedSecret:  getEnv("APPLE_APP_SHARED_SECRET", ""),
		AppleBundleID:         getEnv("APPLE_BUNDLE_ID", "com.soulwi.app"),
		AppleTeamID:           getEnv("APPLE_TEAM_ID", ""),
		DailyMessageLimitFree: getEnvAsInt("DAILY_MESSAGE_LIMIT_FREE", 10),
		CronAuthKey:           getEnv("CRON_AUTH_KEY", ""),
		EnableDevRoutes:       getEnvAsBool("ENABLE_DEV_ROUTES", false),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
