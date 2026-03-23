package handler

import (
	"bff-storage/internal/models"
	"bff-storage/internal/storage"
	"context"
)

type mockStorage struct {
	bucketName    string
	checkBucketFn func(ctx context.Context) error
	listObjectsFn func(ctx context.Context, prefix string) ([]models.ObjectItem, error)
	listLevelFn   func(ctx context.Context, prefix string) (*models.FolderListing, error)
	getFileFn     func(ctx context.Context, id string) (*storage.FileResult, error)
}

func (m *mockStorage) GetBucketName() string {
	return m.bucketName
}

func (m *mockStorage) CheckBucket(ctx context.Context) error {
	return m.checkBucketFn(ctx)
}

func (m *mockStorage) ListObjects(ctx context.Context, prefix string) ([]models.ObjectItem, error) {
	return m.listObjectsFn(ctx, prefix)
}

func (m *mockStorage) ListLevel(ctx context.Context, prefix string) (*models.FolderListing, error) {
	return m.listLevelFn(ctx, prefix)
}

func (m *mockStorage) GetFile(ctx context.Context, id string) (*storage.FileResult, error) {
	return m.getFileFn(ctx, id)
}
