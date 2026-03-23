package handler

import (
	"bff-storage/internal/models"
	"bff-storage/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gcs "cloud.google.com/go/storage"
)

// --------------- Grupo 1: TestHealth ---------------

func TestHealth(t *testing.T) {
	h := New(&mockStorage{})

	t.Run("GET returns 200 with status ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		h.Health(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if body["status"] != "ok" {
			t.Fatalf("expected status ok, got %q", body["status"])
		}
	})

	t.Run("POST returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
		rec := httptest.NewRecorder()
		h.Health(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", rec.Code)
		}
	})
}

// --------------- Grupo 2: TestStartUpProbe ---------------

func TestStartUpProbe(t *testing.T) {
	t.Run("bucket connected returns 200", func(t *testing.T) {
		h := New(&mockStorage{
			checkBucketFn: func(ctx context.Context) error { return nil },
		})
		req := httptest.NewRequest(http.MethodGet, "/probe/startup", nil)
		rec := httptest.NewRecorder()
		h.StartUpProbe(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if !strings.Contains(body["message"], "todo vapor") {
			t.Fatalf("unexpected message: %q", body["message"])
		}
	})

	t.Run("bucket error returns 500", func(t *testing.T) {
		h := New(&mockStorage{
			checkBucketFn: func(ctx context.Context) error { return errors.New("connection refused") },
		})
		req := httptest.NewRequest(http.MethodGet, "/probe/startup", nil)
		rec := httptest.NewRecorder()
		h.StartUpProbe(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("POST returns 405", func(t *testing.T) {
		h := New(&mockStorage{})
		req := httptest.NewRequest(http.MethodPost, "/probe/startup", nil)
		rec := httptest.NewRecorder()
		h.StartUpProbe(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", rec.Code)
		}
	})
}

// --------------- Grupo 3: TestListItems ---------------

func TestListItems(t *testing.T) {
	now := time.Now()

	t.Run("returns 2 items", func(t *testing.T) {
		h := New(&mockStorage{
			bucketName: "test-bucket",
			listObjectsFn: func(ctx context.Context, prefix string) ([]models.ObjectItem, error) {
				return []models.ObjectItem{
					{ID: "file1.txt", Size: 100, ContentType: "text/plain", UpdatedAt: now},
					{ID: "file2.json", Size: 200, ContentType: "application/json", UpdatedAt: now},
				}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items", nil)
		rec := httptest.NewRecorder()
		h.ListItems(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var body map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if body["bucket"] != "test-bucket" {
			t.Fatalf("expected bucket test-bucket, got %v", body["bucket"])
		}
		if int(body["count"].(float64)) != 2 {
			t.Fatalf("expected count 2, got %v", body["count"])
		}
		items := body["items"].([]any)
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		h := New(&mockStorage{
			bucketName: "test-bucket",
			listObjectsFn: func(ctx context.Context, prefix string) ([]models.ObjectItem, error) {
				return []models.ObjectItem{}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items", nil)
		rec := httptest.NewRecorder()
		h.ListItems(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var body map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if int(body["count"].(float64)) != 0 {
			t.Fatalf("expected count 0, got %v", body["count"])
		}
		items := body["items"].([]any)
		if len(items) != 0 {
			t.Fatalf("expected 0 items, got %d", len(items))
		}
	})

	t.Run("storage error returns 500", func(t *testing.T) {
		h := New(&mockStorage{
			listObjectsFn: func(ctx context.Context, prefix string) ([]models.ObjectItem, error) {
				return nil, errors.New("storage unavailable")
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items", nil)
		rec := httptest.NewRecorder()
		h.ListItems(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("POST returns 405", func(t *testing.T) {
		h := New(&mockStorage{})
		req := httptest.NewRequest(http.MethodPost, "/bucket/items", nil)
		rec := httptest.NewRecorder()
		h.ListItems(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", rec.Code)
		}
	})
}

// --------------- Grupo 4: TestListItemsByFolder ---------------

func TestListItemsByFolder(t *testing.T) {
	now := time.Now()

	t.Run("folder=data/images returns folders and items", func(t *testing.T) {
		var capturedPrefix string
		h := New(&mockStorage{
			bucketName: "test-bucket",
			listLevelFn: func(ctx context.Context, prefix string) (*models.FolderListing, error) {
				capturedPrefix = prefix
				return &models.FolderListing{
					Folders: []string{"data/images/thumbs/"},
					Items: []models.ObjectItem{
						{ID: "data/images/photo.png", Size: 500, ContentType: "image/png", UpdatedAt: now},
					},
				}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items/by-folder?folder=data/images", nil)
		rec := httptest.NewRecorder()
		h.ListItemsByFolder(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if capturedPrefix != "data/images/" {
			t.Fatalf("expected prefix data/images/, got %q", capturedPrefix)
		}

		var body map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if body["bucket"] != "test-bucket" {
			t.Fatalf("expected bucket test-bucket, got %v", body["bucket"])
		}
		folders := body["folders"].([]any)
		if len(folders) != 1 {
			t.Fatalf("expected 1 folder, got %d", len(folders))
		}
		if int(body["count"].(float64)) != 1 {
			t.Fatalf("expected count 1, got %v", body["count"])
		}
	})

	t.Run("empty folder normalizes to empty string", func(t *testing.T) {
		var capturedPrefix string
		h := New(&mockStorage{
			bucketName: "test-bucket",
			listLevelFn: func(ctx context.Context, prefix string) (*models.FolderListing, error) {
				capturedPrefix = prefix
				return &models.FolderListing{
					Folders: []string{},
					Items:   []models.ObjectItem{},
				}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items/by-folder?folder=", nil)
		rec := httptest.NewRecorder()
		h.ListItemsByFolder(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if capturedPrefix != "" {
			t.Fatalf("expected empty prefix, got %q", capturedPrefix)
		}
	})

	t.Run("folder with leading slash is normalized", func(t *testing.T) {
		var capturedPrefix string
		h := New(&mockStorage{
			bucketName: "test-bucket",
			listLevelFn: func(ctx context.Context, prefix string) (*models.FolderListing, error) {
				capturedPrefix = prefix
				return &models.FolderListing{
					Folders: []string{},
					Items:   []models.ObjectItem{},
				}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items/by-folder?folder=/data/images", nil)
		rec := httptest.NewRecorder()
		h.ListItemsByFolder(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if capturedPrefix != "data/images/" {
			t.Fatalf("expected prefix data/images/, got %q", capturedPrefix)
		}
	})

	t.Run("storage error returns 500", func(t *testing.T) {
		h := New(&mockStorage{
			listLevelFn: func(ctx context.Context, prefix string) (*models.FolderListing, error) {
				return nil, errors.New("storage error")
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items/by-folder?folder=data", nil)
		rec := httptest.NewRecorder()
		h.ListItemsByFolder(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("POST returns 405", func(t *testing.T) {
		h := New(&mockStorage{})
		req := httptest.NewRequest(http.MethodPost, "/bucket/items/by-folder", nil)
		rec := httptest.NewRecorder()
		h.ListItemsByFolder(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", rec.Code)
		}
	})
}

// --------------- Grupo 5: TestGetFileByID ---------------

func TestGetFileByID(t *testing.T) {
	t.Run("existing file returns content with headers", func(t *testing.T) {
		fileContent := "hello world"
		h := New(&mockStorage{
			getFileFn: func(ctx context.Context, id string) (*storage.FileResult, error) {
				return &storage.FileResult{
					Reader:      io.NopCloser(strings.NewReader(fileContent)),
					ContentType: "text/plain",
					Size:        int64(len(fileContent)),
				}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items/file?id=docs/readme.txt", nil)
		rec := httptest.NewRecorder()
		h.GetFileByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "text/plain" {
			t.Fatalf("expected Content-Type text/plain, got %q", ct)
		}
		if cd := rec.Header().Get("Content-Disposition"); cd != `attachment; filename="readme.txt"` {
			t.Fatalf("unexpected Content-Disposition: %q", cd)
		}
		if cl := rec.Header().Get("Content-Length"); cl != "11" {
			t.Fatalf("expected Content-Length 11, got %q", cl)
		}
		if rec.Body.String() != fileContent {
			t.Fatalf("expected body %q, got %q", fileContent, rec.Body.String())
		}
	})

	t.Run("empty id returns 400", func(t *testing.T) {
		h := New(&mockStorage{})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items/file?id=", nil)
		rec := httptest.NewRecorder()
		h.GetFileByID(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if body["error"] != "query param 'id' is required" {
			t.Fatalf("unexpected error message: %q", body["error"])
		}
	})

	t.Run("file not found returns 404", func(t *testing.T) {
		h := New(&mockStorage{
			getFileFn: func(ctx context.Context, id string) (*storage.FileResult, error) {
				return nil, gcs.ErrObjectNotExist
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items/file?id=missing.txt", nil)
		rec := httptest.NewRecorder()
		h.GetFileByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}

		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if body["error"] != "file not found" {
			t.Fatalf("unexpected error: %q", body["error"])
		}
	})

	t.Run("generic storage error returns 500", func(t *testing.T) {
		h := New(&mockStorage{
			getFileFn: func(ctx context.Context, id string) (*storage.FileResult, error) {
				return nil, errors.New("disk failure")
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items/file?id=file.txt", nil)
		rec := httptest.NewRecorder()
		h.GetFileByID(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("id with leading slash is trimmed", func(t *testing.T) {
		var capturedID string
		h := New(&mockStorage{
			getFileFn: func(ctx context.Context, id string) (*storage.FileResult, error) {
				capturedID = id
				return &storage.FileResult{
					Reader:      io.NopCloser(strings.NewReader("data")),
					ContentType: "application/octet-stream",
					Size:        4,
				}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items/file?id=/path/to/file.json", nil)
		rec := httptest.NewRecorder()
		h.GetFileByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if capturedID != "path/to/file.json" {
			t.Fatalf("expected id path/to/file.json, got %q", capturedID)
		}
	})

	t.Run("file with negative size omits Content-Length", func(t *testing.T) {
		h := New(&mockStorage{
			getFileFn: func(ctx context.Context, id string) (*storage.FileResult, error) {
				return &storage.FileResult{
					Reader:      io.NopCloser(strings.NewReader("data")),
					ContentType: "application/octet-stream",
					Size:        -1,
				}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/bucket/items/file?id=file.bin", nil)
		rec := httptest.NewRecorder()
		h.GetFileByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if cl := rec.Header().Get("Content-Length"); cl != "" {
			t.Fatalf("expected no Content-Length header, got %q", cl)
		}
	})

	t.Run("POST returns 405", func(t *testing.T) {
		h := New(&mockStorage{})
		req := httptest.NewRequest(http.MethodPost, "/bucket/items/file?id=file.txt", nil)
		rec := httptest.NewRecorder()
		h.GetFileByID(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", rec.Code)
		}
	})
}

// --------------- Grupo 6: TestNormalizeFolderPrefix ---------------

func TestNormalizeFolderPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"   ", ""},
		{"data", "data/"},
		{"data/", "data/"},
		{"/data", "data/"},
		{"/data/", "data/"},
		{"a/b/c", "a/b/c/"},
	}

	for _, tc := range tests {
		t.Run("input="+tc.input, func(t *testing.T) {
			got := normalizeFolderPrefix(tc.input)
			if got != tc.expected {
				t.Fatalf("normalizeFolderPrefix(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

// --------------- Grupo 7: TestWriteJSON ---------------

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusCreated, map[string]string{"key": "value"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["key"] != "value" {
		t.Fatalf("expected key=value, got %q", body["key"])
	}
}

// --------------- Grupo 8: TestWriteError ---------------

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusForbidden, "access denied")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["error"] != "access denied" {
		t.Fatalf("expected error=access denied, got %q", body["error"])
	}
}
