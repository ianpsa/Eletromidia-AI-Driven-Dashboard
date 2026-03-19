package main

import (
	"bigquery-writer/internal/config"
	"bigquery-writer/internal/consumer"
	"bigquery-writer/internal/metrics"
	"bigquery-writer/internal/writer"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type healthAssistant struct {
	writer   *writer.Writer
	consumer *consumer.Consumer
	ctx      context.Context
}

func (ha *healthAssistant) healthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(ha.ctx, 5*time.Second)
	defer cancel()

	if err := ha.writer.HealthCheck(ctx); err != nil {
		log.Printf("health check failed (bigquery): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := consumer.KafkaReadinessProbe(ha.consumer.Cfg.KafkaBrokers[0], 5*time.Second); err != nil {
		log.Printf("health check failed (kafka): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w, err := writer.NewWriter(ctx, cfg.GCPProjectID, cfg.BQDatasetID, cfg.FlushSize)
	if err != nil {
		log.Fatalf("failed to create writer: %v", err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Printf("error closing writer: %v", err)
		}
	}()

	c := consumer.New(cfg, w)

	ha := &healthAssistant{writer: w, consumer: c, ctx: ctx}
	reg := prometheus.NewRegistry()
	m := metrics.NewFlushMetrics(reg)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", ha.healthCheck)
	mux.HandleFunc("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Run(ctx, m)
	}()

	<-sigCh
	log.Println("shutdown signal received")
	cancel()
	wg.Wait()

	if err := w.Flush(context.Background(), m); err != nil {
		log.Printf("final flush error: %v", err)
	}

	log.Println("bigquery-writer stopped")
}
