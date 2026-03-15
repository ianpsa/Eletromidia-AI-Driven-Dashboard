package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	BucketName string
	Key        string
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
		Port:       getEnv("PORT", "8080"),
		BucketName: getEnv("BUCKET_NAME", ""),
		Key:        getEnv("CS_SA_CREDENTIALS", ""),
	}
}

func (c Config) Validate() error {
	var missing []string
	if c.BucketName == "" {
		missing = append(missing, "BUCKET_NAME")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}
