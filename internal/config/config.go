package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Phone      string
	AppID      string
	AppHash    string
	UserTarget []string

	MinioHost      string
	MinioPort      string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioSSL       bool
	MinioRegion    string
	MinioEndpoint  string

	AUTO_REMOVE_MEDIA  bool
	WORKER_POOL        string
	SEND_INFO_UPLOADED bool
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() Config {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		// Log error but continue - we might be using environment variables directly
		fmt.Fprintf(os.Stderr, "Warning: Error loading .env file: %v\n", err)
	}

	// Split USER_TARGET by comma to convert string to []string
	userTargets := strings.Split(os.Getenv("USER_TARGET"), ",")

	// Trim spaces from each target
	for i, target := range userTargets {
		userTargets[i] = strings.TrimSpace(target)
	}

	return Config{
		Phone:      os.Getenv("TG_PHONE"),
		AppID:      os.Getenv("APP_ID"),
		AppHash:    os.Getenv("APP_HASH"),
		UserTarget: userTargets,

		MinioHost:      os.Getenv("MINIO_HOST"),
		MinioPort:      os.Getenv("MINIO_PORT"),
		MinioAccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:    os.Getenv("MINIO_BUCKET"),
		MinioSSL:       os.Getenv("MINIO_SSL") == "true",
		MinioRegion:    os.Getenv("MINIO_REGION"),
		MinioEndpoint:  os.Getenv("MINIO_ENDPOINT"),

		AUTO_REMOVE_MEDIA:  os.Getenv("AUTO_REMOVE_MEDIA") == "true",
		WORKER_POOL:        os.Getenv("WORKER_POOL"),
		SEND_INFO_UPLOADED: os.Getenv("SEND_INFO_UPLOADED") == "true",
	}
}
