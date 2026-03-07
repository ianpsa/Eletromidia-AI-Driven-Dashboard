package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type config struct {
	Port       string
	BucketName string
	Key		   string
}

type api struct {
	bucketName string
	bucket     *storage.BucketHandle
}

type objectItem struct {
	ID          string    `json:"id"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type heatlhAssistant struct {
	Api 		*api
	Context 	context.Context
}

type folderListing struct {
	Items   []objectItem `json:"items"`
	Folders []string     `json:"folders"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	cfg := loadConfig()
	if cfg.BucketName == "" {
		log.Fatal("missing env var: BUCKET_NAME")
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(cfg.Key))
	if err != nil {
		log.Fatalf("error creating cloud storage client: %v", err)
	}
	defer client.Close()

	h := &api{
		bucketName: cfg.BucketName,
		bucket:     client.Bucket(cfg.BucketName),
	}

	ha := heatlhAssistant{
		Api: h,
		Context: ctx,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.health)
	mux.HandleFunc("/bucket/items", h.listItems)
	mux.HandleFunc("/bucket/items/by-folder", h.listItemsByFolder)
	mux.HandleFunc("/bucket/items/file", h.getFileByID)
	mux.HandleFunc("/probe/startup", ha.startUpProbe)

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

func (ha *heatlhAssistant) startUpProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, err := ha.Api.bucket.Attrs(ha.Context)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": fmt.Sprintf("Bucket não conectado: %v", err),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "BFF Storage funcionando a todo vapor!!!",
	})
	return

}

func (a *api) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func loadConfig() config {
	return config{
		Port:       getEnv("PORT", "8080"),
		BucketName: getEnv("BUCKET_NAME", ""),
		Key:		getEnv("CS_SA_CREDENTIALS", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}



func (a *api) listItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	items, err := a.listObjects(ctx, "")
	if err != nil {
		log.Printf("error listing bucket items: %v", err)
		writeError(w, http.StatusInternalServerError, "error listing bucket items")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"bucket": a.bucketName,
		"count":  len(items),
		"items":  items,
	})
}

func (a *api) listItemsByFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	folder := normalizeFolderPrefix(r.URL.Query().Get("folder"))
	if folder == "" {
		writeError(w, http.StatusBadRequest, "query param 'folder' is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	listing, err := a.listLevel(ctx, folder)
	if err != nil {
		log.Printf("error listing bucket items by folder: %v", err)
		writeError(w, http.StatusInternalServerError, "error listing bucket items by folder")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"bucket":  a.bucketName,
		"folder":  folder,
		"count":   len(listing.Items),
		"items":   listing.Items,
	})
}

func (a *api) getFileByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := strings.TrimPrefix(strings.TrimSpace(r.URL.Query().Get("id")), "/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "query param 'id' is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	reader, err := a.bucket.Object(id).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}

		log.Printf("error reading file %s: %v", id, err)
		writeError(w, http.StatusInternalServerError, "error reading file")
		return
	}
	defer reader.Close()

	contentType := reader.Attrs.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+path.Base(id)+"\"")
	if reader.Attrs.Size >= 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(reader.Attrs.Size, 10))
	}

	if _, err := io.Copy(w, reader); err != nil {
		log.Printf("error streaming file %s: %v", id, err)
	}
}

func (a *api) listLevel(ctx context.Context, prefix string) (*folderListing, error) {
	query := &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	}

	it := a.bucket.Objects(ctx, query)
	listing := &folderListing{
		Items:   make([]objectItem, 0),
		Folders: make([]string, 0),
	}

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}

		if attrs.Prefix != "" {
			listing.Folders = append(listing.Folders, attrs.Prefix)
			continue
		}

		if strings.HasSuffix(attrs.Name, "/") {
			continue
		}

		listing.Items = append(listing.Items, objectItem{
			ID:          attrs.Name,
			Size:        attrs.Size,
			ContentType: attrs.ContentType,
			UpdatedAt:   attrs.Updated,
		})
	}

	return listing, nil
}

func (a *api) listObjects(ctx context.Context, prefix string) ([]objectItem, error) {
	query := &storage.Query{}
	if prefix != "" {
		query.Prefix = prefix
	}

	it := a.bucket.Objects(ctx, query)
	items := make([]objectItem, 0)
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(attrs.Name, "/") {
			continue
		}

		items = append(items, objectItem{
			ID:          attrs.Name,
			Size:        attrs.Size,
			ContentType: attrs.ContentType,
			UpdatedAt:   attrs.Updated,
		})
	}

	return items, nil
}

func normalizeFolderPrefix(folder string) string {
	folder = strings.TrimPrefix(strings.TrimSpace(folder), "/")
	if folder == "" {
		return ""
	}
	if !strings.HasSuffix(folder, "/") {
		folder += "/"
	}
	return folder
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("error writing response: %v", err)
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
