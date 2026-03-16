package config

import (
	"cloud.google.com/go/bigquery"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"os"
)

type BigQueryConfig struct {
	ProjectID string
	DatasetID string
	Key       string
}

func newBigQueryConfig() BigQueryConfig {
	projectId := getEnv("PROJECT_ID", "default_project")
	datasetId := getEnv("DATASET_ID", "default_dataset")
	keyPath := getEnv("BQ_SA_CREDENTIALS", "default_key")

	bqConfig := BigQueryConfig{ProjectID: projectId, DatasetID: datasetId, Key: keyPath}

	return bqConfig
}

func getEnv(key, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func initClient(ctx context.Context) (*bigquery.Client, BigQueryConfig, error) {
	bqConfig := newBigQueryConfig()

	credBytes, err := os.ReadFile(bqConfig.Key)
	if err != nil {
		return nil, BigQueryConfig{}, fmt.Errorf("[config/bigquery_client] Erro ao ler credenciais: %w", err)
	}

	client, err := bigquery.NewClient(ctx, bqConfig.ProjectID, option.WithCredentialsJSON(credBytes))
	if err != nil {
		return nil, BigQueryConfig{}, fmt.Errorf("[config/bigquery_client] Erro ao criar o cliente BigQuery: %w", err)
	}

	return client, bqConfig, nil
}

func InitBigQuery(ctx context.Context) (*bigquery.Client, *bigquery.Dataset, error) {
	bqClient, bqConfig, err := initClient(ctx)

	if err != nil {
		return nil, nil, err
	}

	dataset := bqClient.Dataset(bqConfig.DatasetID)

	return bqClient, dataset, nil
}
