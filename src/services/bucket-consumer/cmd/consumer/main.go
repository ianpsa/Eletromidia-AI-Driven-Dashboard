package main

import (
	"bucket-consumer/internal/config"
	"bucket-consumer/internal/consumer"
	"bucket-consumer/internal/storage"
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		godotenv.Overload()
	}

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("erro de configuracao: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer, err := storage.NewWriter(
		ctx,
		cfg.GCSBucket,
		cfg.GCSBasePath,
		cfg.GCSCredentials,
		cfg.FlushSize,
	)
	if err != nil {
		log.Fatalf("erro ao criar writer do GCS: %v", err)
	}
	defer writer.Close()

	c := consumer.New(cfg, writer)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Run(ctx)
	}()

	<-sigCh
	log.Println("sinal de encerramento recebido")
	cancel()
	wg.Wait()

	if err := writer.Flush(context.Background()); err != nil {
		log.Printf("erro no flush final: %v", err)
	}

	log.Println("bucket-consumer encerrado")
}
