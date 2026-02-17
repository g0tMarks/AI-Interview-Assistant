package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/storage"
)

type UploadHandler struct {
	store    storage.Store
	maxBytes int64
}

func NewUploadHandler(store storage.Store, maxBytes int64) *UploadHandler {
	return &UploadHandler{store: store, maxBytes: maxBytes}
}

type UploadResponse struct {
	storage.StoredFile
	DownloadURL string `json:"downloadUrl"`
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		http.Error(w, "uploads not configured", http.StatusNotImplemented)
		return
	}

	maxFileBytes := h.maxBytes
	if maxFileBytes <= 0 {
		maxFileBytes = 25 << 20 // 25 MiB default
	}

	// Allow some multipart overhead beyond the file itself.
	r.Body = http.MaxBytesReader(w, r.Body, maxFileBytes+(1<<20))

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	f, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer func() { _ = f.Close() }()

	// Sniff content type if not provided.
	buf := make([]byte, 512)
	n, _ := io.ReadFull(f, buf)
	sniffedType := http.DetectContentType(buf[:n])

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = sniffedType
	}

	reader := io.MultiReader(bytes.NewReader(buf[:n]), f)

	stored, err := h.store.Save(r.Context(), reader, storage.SaveOptions{
		OriginalName: header.Filename,
		ContentType:  contentType,
		MaxBytes:     maxFileBytes,
	})
	if err != nil {
		if errors.Is(err, storage.ErrTooLarge) {
			http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
			return
		}
		if errors.Is(err, storage.ErrEmptyFile) {
			http.Error(w, "empty file", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to store upload", http.StatusInternalServerError)
		return
	}

	resp := UploadResponse{
		StoredFile:  stored,
		DownloadURL: "/uploads/" + stored.Key,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *UploadHandler) Download(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		http.Error(w, "uploads not configured", http.StatusNotImplemented)
		return
	}

	key := chi.URLParam(r, "key")
	meta, rc, err := h.store.Open(r.Context(), key)
	if err != nil {
		if errors.Is(err, storage.ErrBadKey) {
			http.Error(w, "invalid key", http.StatusBadRequest)
			return
		}
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to read upload", http.StatusInternalServerError)
		return
	}
	defer func() { _ = rc.Close() }()

	contentType := meta.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(meta.SizeBytes, 10))
	if meta.SHA256 != "" {
		w.Header().Set("ETag", `"`+meta.SHA256+`"`)
	}
	if meta.OriginalName != "" {
		if v := mime.FormatMediaType("attachment", map[string]string{"filename": meta.OriginalName}); v != "" {
			w.Header().Set("Content-Disposition", v)
		}
	}

	_, _ = io.Copy(w, rc)
}
