package source

import (
	"time"
)

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
	ID        string     `json:"id"`
	Origin    Origin     `json:"origin"`
	Blob      Blob       `json:"blob"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
