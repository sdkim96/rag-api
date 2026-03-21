package source

import (
	"time"
)

type IndexingStatus string

const (
	StatusNotStarted IndexingStatus = "not_started"
	StatusInProgress IndexingStatus = "in_progress"
	StatusCompleted  IndexingStatus = "completed"
	StatusFailed     IndexingStatus = "failed"
)

type SourceMeta struct {
	Description string `json:"description,omitempty"`
	Author      string `json:"author,omitempty"`
	Version     int    `json:"version,omitempty"`
}

type Origin struct {
	URI      string            `json:"uri"`
	MimeType string            `json:"mime_type"`
	Headers  map[string]string `json:"headers,omitempty"`
}

type Blob struct {
	Key       string    `json:"key"`
	URI       string    `json:"uri"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type Source struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	URI       string    `json:"uri"`
	MimeType  string    `json:"mime_type"`
	Name      string    `json:"name"`
	Origin    Origin    `json:"origin"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type SourceIndexing struct {
	Source
	Status IndexingStatus `json:"status"`
}
