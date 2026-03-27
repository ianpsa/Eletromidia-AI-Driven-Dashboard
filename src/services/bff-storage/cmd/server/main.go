package main

import (
	"bff-storage/internal/bigquery"
	"bff-storage/internal/config"
	"bff-storage/internal/handler"
	"bff-storage/internal/storage"
	"context"
	"errors"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx, cfg.BucketName, cfg.Key)
	if err != nil {
		log.Fatalf("error creating cloud storage client: %v", err)
	}
	defer func() {
		if err := storageClient.Close(); err != nil {
			log.Printf("error closing storage client: %v", err)
		}
	}()

	bqClient, err := bigquery.NewClient(ctx, cfg.BQProjectID, cfg.BQDatasetID, cfg.BQKey)
	if err != nil {
		log.Fatalf("error creating bigquery client: %v", err)
	}
	defer func() {
		if err := bqClient.Close(); err != nil {
			log.Printf("error closing bigquery client: %v", err)
		}
	}()

	h := handler.New(storageClient, bqClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.Health)
	mux.HandleFunc("/bucket/items", h.ListItems)
	mux.HandleFunc("/bucket/items/by-folder", h.ListItemsByFolder)
	mux.HandleFunc("/bucket/items/file", h.GetFileByID)
	mux.HandleFunc("/probe/startup", h.StartUpProbe)
	mux.HandleFunc("/geodata/points", h.GetGeoPoints)
	mux.HandleFunc("/geodata/demographics", h.GetDemographics)
	mux.HandleFunc("/geodata/filter-options", h.GetFilterOptions)
	mux.HandleFunc("/geodata/compare", h.GetCompare)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("bff-storage listening on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
