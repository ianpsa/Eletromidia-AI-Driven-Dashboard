package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Overload(); err != nil {
			log.Printf("warning: could not load .env file: %v", err)
		}
	}

	dateStr := flag.String("date", "", "data para agregar (formato: YYYY-MM-DD). default: ontem")
	flag.Parse()

	var targetDate time.Time
	if *dateStr == "" {
		targetDate = time.Now().UTC().AddDate(0, 0, -1)
	} else {
		var err error
		targetDate, err = time.Parse("2006-01-02", *dateStr)
		if err != nil {
			log.Fatalf("formato de data invalido (use YYYY-MM-DD): %v", err)
		}
	}

	bucket := getEnv("GCS_BUCKET", "kafka-backup-eletromidia")
	credentials := getEnv("GCS_CREDENTIALS", "")
	basePath := getEnv("GCS_BASE_PATH", "kafka-backup")
	topic := getEnv("KAFKA_TOPIC", "geodata")

	ctx := context.Background()

	var opts []option.ClientOption
	if credentials != "" {
		credBytes, err := os.ReadFile(credentials)
		if err != nil {
			log.Fatalf("erro ao ler arquivo de credenciais: %v", err)
		}
		opts = append(opts, option.WithCredentialsJSON(credBytes))
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		log.Fatalf("erro ao criar cliente GCS: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("error closing gcs client: %v", err)
		}
	}()

	srcPrefix := fmt.Sprintf("%s/topics/%s/year=%04d/month=%02d/day=%02d/",
		basePath, topic,
		targetDate.Year(), targetDate.Month(), targetDate.Day(),
	)

	dstPath := fmt.Sprintf("daily/%s/year=%04d/month=%02d/day=%02d/%s.json",
		topic,
		targetDate.Year(), targetDate.Month(), targetDate.Day(),
		topic,
	)

	log.Printf("agregando fragmentos de gs://%s/%s", bucket, srcPrefix)

	bkt := client.Bucket(bucket)
	it := bkt.Objects(ctx, &storage.Query{Prefix: srcPrefix})

	var totalFiles int
	var totalBytes int64

	writer := bkt.Object(dstPath).NewWriter(ctx)
	writer.ContentType = "application/x-ndjson"

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			if closeErr := writer.Close(); closeErr != nil {
				log.Printf("error closing gcs writer: %v", closeErr)
			}
			log.Fatalf("erro ao listar objetos: %v", err)
		}

		reader, err := bkt.Object(attrs.Name).NewReader(ctx)
		if err != nil {
			if closeErr := writer.Close(); closeErr != nil {
				log.Printf("error closing gcs writer: %v", closeErr)
			}
			log.Fatalf("erro ao ler %s: %v", attrs.Name, err)
		}

		n, err := io.Copy(writer, reader)
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("error closing gcs reader: %v", closeErr)
		}
		if err != nil {
			if closeErr := writer.Close(); closeErr != nil {
				log.Printf("error closing gcs writer: %v", closeErr)
			}
			log.Fatalf("erro ao copiar %s: %v", attrs.Name, err)
		}

		totalFiles++
		totalBytes += n
		log.Printf("  lido: %s (%d bytes)", attrs.Name, n)
	}

	if totalFiles == 0 {
		if err := writer.Close(); err != nil {
			log.Printf("error closing gcs writer: %v", err)
		}
		log.Printf("nenhum fragmento encontrado para %s", targetDate.Format("2006-01-02"))
		return
	}

	if err := writer.Close(); err != nil {
		log.Fatalf("erro ao finalizar escrita: %v", err)
	}

	log.Printf("agregacao concluida: %d fragmentos (%d bytes) → gs://%s/%s",
		totalFiles, totalBytes, bucket, dstPath)
}
