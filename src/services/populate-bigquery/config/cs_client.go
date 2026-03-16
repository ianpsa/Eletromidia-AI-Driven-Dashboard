package config

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type CloudStorageConfig struct {
	Bucket string
	File   string
	Key    string
}

func newCloudStorageConfig() CloudStorageConfig {
	return CloudStorageConfig{Bucket: getEnv("BUCKET_NAME", "default_bucket"), File: getEnv("FILE_NAME", "default_file"), Key: getEnv("CS_SA_CREDENTIALS", "default_key")}
}

func InitCloudStorage(ctx context.Context) (CloudStorageConfig, *storage.Client, *storage.ObjectHandle, error) {
	csConfig := newCloudStorageConfig()

	fmt.Printf("Configuração nome do bucket: %s\n", csConfig.Bucket)

	credBytes, err := os.ReadFile(csConfig.Key)
	if err != nil {
		return CloudStorageConfig{}, nil, nil, fmt.Errorf("[config/cs_client] Erro ao ler credenciais: %w", err)
	}

	csClient, err := storage.NewClient(ctx, option.WithCredentialsJSON(credBytes))
	if err != nil {
		return CloudStorageConfig{}, nil, nil, fmt.Errorf("[config/cs_client] Erro ao criar cliente CloudStorage %v", err)
	}

	file := csClient.Bucket(csConfig.Bucket).Object(csConfig.File)

	return csConfig, csClient, file, nil
}
