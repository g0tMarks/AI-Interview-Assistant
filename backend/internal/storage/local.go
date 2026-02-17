package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("file not found")
	ErrTooLarge  = errors.New("file too large")
	ErrBadKey    = errors.New("invalid key")
	ErrEmptyFile = errors.New("empty file")
)

type LocalStore struct {
	baseDir string
}

func NewLocalStore(baseDir string) *LocalStore {
	return &LocalStore{baseDir: baseDir}
}

func (s *LocalStore) Save(ctx context.Context, r io.Reader, opts SaveOptions) (StoredFile, error) {
	_ = ctx

	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return StoredFile{}, fmt.Errorf("create uploads dir: %w", err)
	}

	key := uuid.NewString()
	dataPath := s.dataPath(key)
	metaPath := s.metaPath(key)
	tmpPath := dataPath + ".tmp"

	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return StoredFile{}, fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	var reader io.Reader = io.TeeReader(r, h)
	if opts.MaxBytes > 0 {
		reader = &io.LimitedReader{R: reader, N: opts.MaxBytes + 1}
	}

	n, err := io.Copy(f, reader)
	if err != nil {
		_ = os.Remove(tmpPath)
		return StoredFile{}, fmt.Errorf("write file: %w", err)
	}

	if opts.MaxBytes > 0 && n > opts.MaxBytes {
		_ = os.Remove(tmpPath)
		return StoredFile{}, ErrTooLarge
	}
	if n == 0 {
		_ = os.Remove(tmpPath)
		return StoredFile{}, ErrEmptyFile
	}

	meta := StoredFile{
		Key:          key,
		OriginalName: sanitizeFilename(opts.OriginalName),
		ContentType:  strings.TrimSpace(opts.ContentType),
		SizeBytes:    n,
		SHA256:       hex.EncodeToString(h.Sum(nil)),
		CreatedAt:    time.Now().UTC(),
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return StoredFile{}, fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, dataPath); err != nil {
		_ = os.Remove(tmpPath)
		return StoredFile{}, fmt.Errorf("commit data file: %w", err)
	}

	if err := writeJSONFile(metaPath, meta, 0o600); err != nil {
		_ = os.Remove(dataPath)
		return StoredFile{}, fmt.Errorf("write metadata: %w", err)
	}

	return meta, nil
}

func (s *LocalStore) Open(ctx context.Context, key string) (StoredFile, io.ReadCloser, error) {
	_ = ctx

	if _, err := uuid.Parse(key); err != nil {
		return StoredFile{}, nil, ErrBadKey
	}

	metaPath := s.metaPath(key)
	dataPath := s.dataPath(key)

	meta, err := readJSONFile[StoredFile](metaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return StoredFile{}, nil, ErrNotFound
		}
		return StoredFile{}, nil, fmt.Errorf("read metadata: %w", err)
	}

	f, err := os.Open(dataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return StoredFile{}, nil, ErrNotFound
		}
		return StoredFile{}, nil, fmt.Errorf("open data: %w", err)
	}

	return meta, f, nil
}

func (s *LocalStore) dataPath(key string) string {
	return filepath.Join(s.baseDir, key+".data")
}

func (s *LocalStore) metaPath(key string) string {
	return filepath.Join(s.baseDir, key+".json")
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "\x00", "")
	if name == "." || name == string(filepath.Separator) {
		return ""
	}
	// Avoid absurd names (defense-in-depth)
	if len(name) > 255 {
		return name[:255]
	}
	return name
}

func writeJSONFile(path string, v any, perm os.FileMode) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readJSONFile[T any](path string) (T, error) {
	var zero T
	b, err := os.ReadFile(path)
	if err != nil {
		return zero, err
	}
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		return zero, err
	}
	return v, nil
}
