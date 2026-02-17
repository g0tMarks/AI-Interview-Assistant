package storage

import (
	"context"
	"io"
	"time"
)

type SaveOptions struct {
	OriginalName string
	ContentType  string
	MaxBytes     int64 // if > 0, reject files larger than this
}

type StoredFile struct {
	Key          string    `json:"key"`
	OriginalName string    `json:"originalName"`
	ContentType  string    `json:"contentType"`
	SizeBytes    int64     `json:"sizeBytes"`
	SHA256       string    `json:"sha256"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Store interface {
	Save(ctx context.Context, r io.Reader, opts SaveOptions) (StoredFile, error)
	Open(ctx context.Context, key string) (StoredFile, io.ReadCloser, error)
}
