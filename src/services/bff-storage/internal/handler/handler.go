package handler

import (
	"bff-storage/internal/models"
	"bff-storage/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	gcs "cloud.google.com/go/storage"
)

type StorageService interface {
	CheckBucket(ctx context.Context) error
	ListObjects(ctx context.Context, prefix string) ([]models.ObjectItem, error)
	ListLevel(ctx context.Context, prefix string) (*models.FolderListing, error)
	GetFile(ctx context.Context, id string) (*storage.FileResult, error)
	GetBucketName() string
}

type Handler struct {
	Storage StorageService
}

func New(s StorageService) *Handler {
	return &Handler{Storage: s}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) StartUpProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := h.Storage.CheckBucket(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": fmt.Sprintf("Bucket não conectado: %v", err),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "BFF Storage funcionando a todo vapor!!!",
	})
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	items, err := h.Storage.ListObjects(ctx, "")
	if err != nil {
		log.Printf("error listing bucket items: %v", err)
		writeError(w, http.StatusInternalServerError, "error listing bucket items")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"bucket": h.Storage.GetBucketName(),
		"count":  len(items),
		"items":  items,
	})
}

func (h *Handler) ListItemsByFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	raw := r.URL.Query().Get("folder")
	folder := normalizeFolderPrefix(raw)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	listing, err := h.Storage.ListLevel(ctx, folder)
	if err != nil {
		log.Printf("error listing bucket items by folder: %v", err)
		writeError(w, http.StatusInternalServerError, "error listing bucket items by folder")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"bucket":  h.Storage.GetBucketName(),
		"folder":  folder,
		"folders": listing.Folders,
		"count":   len(listing.Items),
		"items":   listing.Items,
	})
}

func (h *Handler) GetFileByID(w http.ResponseWriter, r *http.Request) {
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

	file, err := h.Storage.GetFile(ctx, id)
	if err != nil {
		if errors.Is(err, gcs.ErrObjectNotExist) {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}

		log.Printf("error reading file %s: %v", id, err)
		writeError(w, http.StatusInternalServerError, "error reading file")
		return
	}
	defer file.Reader.Close()

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+path.Base(id)+"\"")
	if file.Size >= 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	}

	if _, err := io.Copy(w, file.Reader); err != nil {
		log.Printf("error streaming file %s: %v", id, err)
	}
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
