package config

import(
    "cloud.google.com/go/bigquery"
    "os"
    "context"
    "fmt"
    "google.golang.org/api/option"
)

type BigQueryConfig struct {
    ProjectID       string
    DatasetID       string
    Key             string
}

func newBigQueryConfig() BigQueryConfig {
    projectId := getEnv("PROJECT_ID", "default_project")
    datasetId := getEnv("DATASET_ID", "default_dataset")
    keyPath := getEnv("BQ_SA_CREDENTIALS", "default_key")

    bqConfig := BigQueryConfig{  ProjectID: projectId, DatasetID: datasetId, Key: keyPath,  }

    return bqConfig
}

func getEnv(key, def string) string {
    value := os.Getenv(key)
    if value == "" {
        return def
    }
    return value
}


func initClient(ctx context.Context) (*bigquery.Client, error, BigQueryConfig) {
    bqConfig := newBigQueryConfig()

    client, err := bigquery.NewClient(ctx, bqConfig.ProjectID, option.WithCredentialsFile(bqConfig.Key))
    if err != nil {
        return nil, fmt.Errorf("[config/bigquery_client] Erro ao criar o cliente BigQuery: %v", err), BigQueryConfig{}
    }

    return client, nil, bqConfig
}

func InitBigQuery(ctx context.Context) (*bigquery.Client, *bigquery.Dataset, error) {
    bqClient, err, bqConfig := initClient(ctx)

    if err != nil {
        return nil, nil, err
    }

    dataset := bqClient.Dataset(bqConfig.DatasetID)

    return bqClient, dataset, nil
}