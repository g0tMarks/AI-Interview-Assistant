package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/extraction"
)

type RubricHandler struct {
	q *db.Queries
}

func NewRubricHandler(q *db.Queries) *RubricHandler {
	return &RubricHandler{q: q}
}

type CreateRubricRequest struct {
	TeacherID   uuid.UUID `json:"teacherId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	RawText     string    `json:"rawText"`
}

type RubricResponse struct {
	RubricID    uuid.UUID          `json:"rubricId"`
	TeacherID   uuid.UUID          `json:"teacherId"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	RawText     string             `json:"rawText"`
	IsEnabled   bool               `json:"isEnabled"`
	CreatedAt   pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt   pgtype.Timestamptz `json:"updatedAt"`
}

func (h *RubricHandler) CreateRubric(w http.ResponseWriter, r *http.Request) {
	var req CreateRubricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Minimal validation: we just ensure there's enough info to be useful
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if req.TeacherID == uuid.Nil {
		http.Error(w, "teacherId is required", http.StatusBadRequest)
		return
	}
	if req.RawText == "" {
		http.Error(w, "rawText is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Convert uuid.UUID to pgtype.UUID
	teacherID := pgtype.UUID{
		Bytes: req.TeacherID,
		Valid: true,
	}

	// Convert string to pgtype.Text for description
	description := pgtype.Text{}
	if req.Description != "" {
		description.String = req.Description
		description.Valid = true
	}

	rubric, err := h.q.CreateRubric(ctx, db.CreateRubricParams{
		TeacherID:   teacherID,
		Title:       req.Title,
		Description: description,
		RawText:     req.RawText,
	})
	if err != nil {
		// Log the actual error for debugging (in production, be more careful about exposing errors)
		http.Error(w, "failed to save rubric: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert pgtype.UUID to uuid.UUID for response
	var rubricID uuid.UUID
	if rubric.RubricID.Valid {
		rubricID = rubric.RubricID.Bytes
	}

	var teacherIDResp uuid.UUID
	if rubric.TeacherID.Valid {
		teacherIDResp = rubric.TeacherID.Bytes
	}

	resp := RubricResponse{
		RubricID:    rubricID,
		TeacherID:   teacherIDResp,
		Title:       rubric.Title,
		Description: rubric.Description.String,
		RawText:     rubric.RawText,
		IsEnabled:   rubric.IsEnabled,
		CreatedAt:   rubric.CreatedAt,
		UpdatedAt:   rubric.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RubricHandler) ListRubrics(w http.ResponseWriter, r *http.Request) {
	// Extract teacherId from query parameter
	teacherIdStr := r.URL.Query().Get("teacherId")
	if teacherIdStr == "" {
		http.Error(w, "teacherId is required", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	teacherID, err := uuid.Parse(teacherIdStr)
	if err != nil {
		http.Error(w, "invalid teacherId format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Convert to pgtype.UUID
	teacherIDPgtype := pgtype.UUID{
		Bytes: teacherID,
		Valid: true,
	}

	// Query database
	rubrics, err := h.q.ListRubricsByTeacher(ctx, teacherIDPgtype)
	if err != nil {
		http.Error(w, "failed to retrieve rubrics", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	resp := make([]RubricResponse, len(rubrics))
	for i, rubric := range rubrics {
		var rubricID uuid.UUID
		if rubric.RubricID.Valid {
			rubricID = rubric.RubricID.Bytes
		}

		var teacherIDResp uuid.UUID
		if rubric.TeacherID.Valid {
			teacherIDResp = rubric.TeacherID.Bytes
		}

		resp[i] = RubricResponse{
			RubricID:    rubricID,
			TeacherID:   teacherIDResp,
			Title:       rubric.Title,
			Description: rubric.Description.String,
			RawText:     rubric.RawText,
			IsEnabled:   rubric.IsEnabled,
			CreatedAt:   rubric.CreatedAt,
			UpdatedAt:   rubric.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// UploadRubricFile handles file uploads (PDF/DOCX), extracts text, and creates a rubric.
// Expects multipart/form-data with:
//   - file: the PDF or DOCX file
//   - teacherId: UUID of the teacher
//   - title: title for the rubric (optional, defaults to filename)
//   - description: description for the rubric (optional)
func (h *RubricHandler) UploadRubricFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 25 MB)
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	// Get teacherId from form
	teacherIdStr := r.FormValue("teacherId")
	if teacherIdStr == "" {
		http.Error(w, "teacherId is required", http.StatusBadRequest)
		return
	}

	teacherID, err := uuid.Parse(teacherIdStr)
	if err != nil {
		http.Error(w, "invalid teacherId format", http.StatusBadRequest)
		return
	}

	// Get optional title and description
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		// Default to filename without extension
		filename := header.Filename
		if idx := strings.LastIndex(filename, "."); idx > 0 {
			title = filename[:idx]
		} else {
			title = filename
		}
	}

	description := strings.TrimSpace(r.FormValue("description"))

	// Read file into memory for extraction (needs to be seekable for PDF)
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	// Extract text from file
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect from filename
		filename := strings.ToLower(header.Filename)
		if strings.HasSuffix(filename, ".pdf") {
			contentType = "application/pdf"
		} else if strings.HasSuffix(filename, ".docx") {
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		} else if strings.HasSuffix(filename, ".doc") {
			contentType = "application/msword"
		}
	}

	reader := bytes.NewReader(fileBytes)
	extractedText, err := extraction.ExtractText(reader, contentType, header.Filename)
	if err != nil {
		if errors.Is(err, extraction.ErrUnsupportedFormat) {
			http.Error(w, "unsupported file format. Only PDF and DOCX files are supported", http.StatusBadRequest)
			return
		}
		if errors.Is(err, extraction.ErrEmptyDocument) {
			msg := "file contains no extractable text"
			if strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
				msg += " (PDF may be image-only/scanned; use a PDF with a text layer or run OCR first)"
			}
			msg += "."
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to extract text from file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// Convert uuid.UUID to pgtype.UUID
	teacherIDPgtype := pgtype.UUID{
		Bytes: teacherID,
		Valid: true,
	}

	// Convert string to pgtype.Text for description
	descriptionPgtype := pgtype.Text{}
	if description != "" {
		descriptionPgtype.String = description
		descriptionPgtype.Valid = true
	}

	// Create rubric with extracted text
	rubric, err := h.q.CreateRubric(ctx, db.CreateRubricParams{
		TeacherID:   teacherIDPgtype,
		Title:       title,
		Description: descriptionPgtype,
		RawText:     extractedText,
	})
	if err != nil {
		http.Error(w, "failed to save rubric: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert pgtype.UUID to uuid.UUID for response
	var rubricID uuid.UUID
	if rubric.RubricID.Valid {
		rubricID = rubric.RubricID.Bytes
	}

	var teacherIDResp uuid.UUID
	if rubric.TeacherID.Valid {
		teacherIDResp = rubric.TeacherID.Bytes
	}

	resp := RubricResponse{
		RubricID:    rubricID,
		TeacherID:   teacherIDResp,
		Title:       rubric.Title,
		Description: rubric.Description.String,
		RawText:     rubric.RawText,
		IsEnabled:   rubric.IsEnabled,
		CreatedAt:   rubric.CreatedAt,
		UpdatedAt:   rubric.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
