package models

import "time"

type ObjectItem struct {
	ID          string    `json:"id"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FolderListing struct {
	Items   []ObjectItem `json:"items"`
	Folders []string     `json:"folders"`
}
