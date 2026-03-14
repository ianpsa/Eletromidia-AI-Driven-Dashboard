package storage

import (
	"bff-storage/internal/models"
	"context"
	"errors"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Client struct {
	BucketName string
	bucket     *storage.BucketHandle
	client     *storage.Client
}

type FileResult struct {
	Reader      io.ReadCloser
	ContentType string
	Size        int64
}

func NewClient(ctx context.Context, bucketName, credentialsFile string) (*Client, error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, err
	}

	return &Client{
		BucketName: bucketName,
		bucket:     client.Bucket(bucketName),
		client:     client,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) CheckBucket(ctx context.Context) error {
	_, err := c.bucket.Attrs(ctx)
	return err
}

func (c *Client) ListObjects(ctx context.Context, prefix string) ([]models.ObjectItem, error) {
	query := &storage.Query{}
	if prefix != "" {
		query.Prefix = prefix
	}

	it := c.bucket.Objects(ctx, query)
	items := make([]models.ObjectItem, 0)
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

		items = append(items, models.ObjectItem{
			ID:          attrs.Name,
			Size:        attrs.Size,
			ContentType: attrs.ContentType,
			UpdatedAt:   attrs.Updated,
		})
	}

	return items, nil
}

func (c *Client) ListLevel(ctx context.Context, prefix string) (*models.FolderListing, error) {
	query := &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	}

	it := c.bucket.Objects(ctx, query)
	listing := &models.FolderListing{
		Items:   make([]models.ObjectItem, 0),
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

		listing.Items = append(listing.Items, models.ObjectItem{
			ID:          attrs.Name,
			Size:        attrs.Size,
			ContentType: attrs.ContentType,
			UpdatedAt:   attrs.Updated,
		})
	}

	return listing, nil
}

func (c *Client) GetFile(ctx context.Context, id string) (*FileResult, error) {
	reader, err := c.bucket.Object(id).NewReader(ctx)
	if err != nil {
		return nil, err
	}

	contentType := reader.Attrs.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return &FileResult{
		Reader:      reader,
		ContentType: contentType,
		Size:        reader.Attrs.Size,
	}, nil
}
