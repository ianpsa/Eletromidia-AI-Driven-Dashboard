package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	KafkaBrokers     []string
	KafkaTopic       string
	KafkaGroupID     string
	KafkaMinBytes    int
	KafkaMaxBytes    int
	KafkaReadTimeout time.Duration
	KafkaMaxWait     time.Duration

	GCPProjectID string
	BQDatasetID  string

	FlushInterval time.Duration
	FlushSize     int
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func Load() Config {
	brokers := getEnv("KAFKA_BROKERS", "my-cluster-kafka-bootstrap.kafka.svc:9092")
	return Config{
		KafkaBrokers:     strings.Split(brokers, ","),
		KafkaTopic:       getEnv("KAFKA_TOPIC", "geodata"),
		KafkaGroupID:     getEnv("KAFKA_GROUP_ID", "bigquery-writer-group"),
		KafkaMinBytes:    getEnvInt("KAFKA_MIN_BYTES", 1000),
		KafkaMaxBytes:    getEnvInt("KAFKA_MAX_BYTES", 10_000_000),
		KafkaReadTimeout: getEnvDuration("KAFKA_READ_TIMEOUT", 5*time.Second),
		KafkaMaxWait:     getEnvDuration("KAFKA_MAX_WAIT", 500*time.Millisecond),

		GCPProjectID: getEnv("GCP_PROJECT_ID", ""),
		BQDatasetID:  getEnv("BQ_DATASET_ID", ""),

		FlushInterval: getEnvDuration("FLUSH_INTERVAL", 30*time.Second),
		FlushSize:     getEnvInt("FLUSH_SIZE", 500),
	}
}

func (c Config) Validate() error {
	var missing []string
	if c.GCPProjectID == "" {
		missing = append(missing, "GCP_PROJECT_ID")
	}
	if c.BQDatasetID == "" {
		missing = append(missing, "BQ_DATASET_ID")
	}
	if len(c.KafkaBrokers) == 0 || c.KafkaBrokers[0] == "" {
		missing = append(missing, "KAFKA_BROKERS")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}
