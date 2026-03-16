package main

import (
	"bucket-consumer/internal/config"
	"bucket-consumer/internal/consumer"
	"bucket-consumer/internal/storage"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

type HealthAssistant struct {
	BucketWriter *storage.Writer
	Consumer     *consumer.Consumer
	Context      context.Context
}

func (ha *HealthAssistant) readnessProbe(w http.ResponseWriter, r *http.Request) {
	// log.Println("Opa, fizeram chamada no Readness Probe!")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	err := consumer.KafkaReadinessProbe(ha.Consumer.Cfg.KafkaBrokers[0], ha.Consumer.Cfg.KafkaReadTimeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = ha.BucketWriter.BucketReadnessProbe(ha.Context)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Overload(); err != nil {
			log.Printf("warning: could not load .env file: %v", err)
		}
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
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("error closing gcs writer: %v", err)
		}
	}()

	c := consumer.New(cfg, writer)

	ha := HealthAssistant{BucketWriter: writer, Consumer: c, Context: ctx}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", ha.readnessProbe)

	server := &http.Server{
		Addr:              ":" + "8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v\n", err)
		}
	}()

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
