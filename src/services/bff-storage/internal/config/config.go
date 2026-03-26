package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	BucketName  string
	Key         string
	BQProjectID string
	BQDatasetID string
	BQKey       string
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		fmt.Println("no .env file found, using environment variables")
	}

	return Config{
		Port:        getEnv("PORT", "8080"),
		BucketName:  getEnv("BUCKET_NAME", ""),
		Key:         getEnv("CS_SA_CREDENTIALS", ""),
		BQProjectID: getEnv("BQ_PROJECT_ID", ""),
		BQDatasetID: getEnv("BQ_DATASET_ID", ""),
		BQKey:       getEnv("BQ_SA_CREDENTIALS", ""),
	}
}

func (c Config) Validate() error {
	var missing []string
	if c.BucketName == "" {
		missing = append(missing, "BUCKET_NAME")
	}
	if c.BQProjectID == "" {
		missing = append(missing, "BQ_PROJECT_ID")
	}
	if c.BQDatasetID == "" {
		missing = append(missing, "BQ_DATASET_ID")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}
