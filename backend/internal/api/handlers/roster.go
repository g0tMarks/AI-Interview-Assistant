package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xuri/excelize/v2"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
)

type RosterHandler struct {
	q *db.Queries
}

func NewRosterHandler(q *db.Queries) *RosterHandler {
	return &RosterHandler{q: q}
}

type AddToRosterRequest struct {
	StudentID uuid.UUID `json:"studentId"`
}

type RosterEntryResponse struct {
	ClassID            uuid.UUID          `json:"classId"`
	StudentID          uuid.UUID          `json:"studentId"`
	JoinedAt           pgtype.Timestamptz `json:"joinedAt"`
	StudentEmail       string             `json:"studentEmail"`
	StudentDisplayName string             `json:"studentDisplayName"`
}

func (h *RosterHandler) AddToRoster(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "classId")
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}
	var req AddToRosterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.StudentID == uuid.Nil {
		http.Error(w, "studentId is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	// Verify class exists
	cid := pgtype.UUID{Bytes: classID, Valid: true}
	sid := pgtype.UUID{Bytes: req.StudentID, Valid: true}
	_, err = h.q.GetClassByID(ctx, cid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "class not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get class", http.StatusInternalServerError)
		return
	}
	_, err = h.q.GetStudentByID(ctx, sid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get student", http.StatusInternalServerError)
		return
	}

	_, err = h.q.AddToRoster(ctx, db.AddToRosterParams{ClassID: cid, StudentID: sid})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "student is already in this class", http.StatusConflict)
			return
		}
		http.Error(w, "failed to add student to class", http.StatusInternalServerError)
		return
	}

	// Return roster entry with student details
	rows, err := h.q.ListRosterByClass(ctx, cid)
	if err != nil {
		http.Error(w, "failed to list roster", http.StatusInternalServerError)
		return
	}
	for _, row := range rows {
		if row.StudentID.Bytes == req.StudentID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(RosterEntryResponse{
				ClassID:            classID,
				StudentID:          row.StudentID.Bytes,
				JoinedAt:           row.JoinedAt,
				StudentEmail:       row.StudentEmail,
				StudentDisplayName: row.StudentDisplayName,
			})
			return
		}
	}
	// Fallback: just 201
	w.WriteHeader(http.StatusCreated)
}

func (h *RosterHandler) RemoveFromRoster(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "classId")
	studentIDStr := chi.URLParam(r, "studentId")
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	err = h.q.RemoveFromRoster(ctx, db.RemoveFromRosterParams{
		ClassID:   pgtype.UUID{Bytes: classID, Valid: true},
		StudentID: pgtype.UUID{Bytes: studentID, Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to remove from roster", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RosterHandler) ListRoster(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "classId")
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	cid := pgtype.UUID{Bytes: classID, Valid: true}
	rows, err := h.q.ListRosterByClass(ctx, cid)
	if err != nil {
		http.Error(w, "failed to list roster", http.StatusInternalServerError)
		return
	}
	out := make([]RosterEntryResponse, len(rows))
	for i := range rows {
		out[i] = RosterEntryResponse{
			ClassID:            classID,
			StudentID:          rows[i].StudentID.Bytes,
			JoinedAt:           rows[i].JoinedAt,
			StudentEmail:       rows[i].StudentEmail,
			StudentDisplayName: rows[i].StudentDisplayName,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type RosterUploadResponse struct {
	CreatedCount      int      `json:"createdCount"`
	AddedToRosterCount int     `json:"addedToRosterCount"`
	SkippedCount      int      `json:"skippedCount"`
	ErrorCount        int      `json:"errorCount"`
	Errors            []string `json:"errors,omitempty"`
}

func (h *RosterHandler) UploadRoster(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "id")
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	cid := pgtype.UUID{Bytes: classID, Valid: true}

	// Verify class exists
	_, err = h.q.GetClassByID(ctx, cid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "class not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get class", http.StatusInternalServerError)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	// Check file extension
	filename := header.Filename
	if !strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		http.Error(w, "file must be a .xlsx file", http.StatusBadRequest)
		return
	}

	// Read file into memory
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	// Open Excel file
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to open Excel file: %v", err), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			// Log error but don't fail the request
		}
	}()

	// Get the first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		http.Error(w, "Excel file has no sheets", http.StatusBadRequest)
		return
	}

	// Get all rows
	rows, err := f.GetRows(sheetName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read rows: %v", err), http.StatusBadRequest)
		return
	}

	if len(rows) == 0 {
		http.Error(w, "Excel file is empty", http.StatusBadRequest)
		return
	}

	// Find column indices for first name, last name, email
	headerRow := rows[0]
	var firstNameCol, lastNameCol, emailCol int = -1, -1, -1

	for i, cell := range headerRow {
		cellLower := strings.ToLower(strings.TrimSpace(cell))
		switch cellLower {
		case "first name", "firstname", "first_name", "fname":
			firstNameCol = i
		case "last name", "lastname", "last_name", "lname":
			lastNameCol = i
		case "email", "e-mail", "email address":
			emailCol = i
		}
	}

	if firstNameCol == -1 || lastNameCol == -1 || emailCol == -1 {
		http.Error(w, "Excel file must contain columns: first name, last name, email", http.StatusBadRequest)
		return
	}

	// Process rows (skip header)
	summary := RosterUploadResponse{
		Errors: []string{},
	}

	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		if len(row) <= firstNameCol || len(row) <= lastNameCol || len(row) <= emailCol {
			summary.ErrorCount++
			summary.Errors = append(summary.Errors, fmt.Sprintf("Row %d: missing required columns", rowIdx+1))
			continue
		}

		firstName := strings.TrimSpace(row[firstNameCol])
		lastName := strings.TrimSpace(row[lastNameCol])
		email := strings.TrimSpace(row[emailCol])

		// Validate email
		if email == "" {
			summary.ErrorCount++
			summary.Errors = append(summary.Errors, fmt.Sprintf("Row %d: email is required", rowIdx+1))
			continue
		}

		// Build display name
		displayName := strings.TrimSpace(firstName + " " + lastName)
		if displayName == "" {
			displayName = email // Fallback to email if no name provided
		}

		// Get or create student
		student, err := h.q.GetStudentByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Create new student
				student, err = h.q.CreateStudent(ctx, db.CreateStudentParams{
					Email:       email,
					DisplayName: displayName,
				})
				if err != nil {
					summary.ErrorCount++
					summary.Errors = append(summary.Errors, fmt.Sprintf("Row %d (%s): failed to create student: %v", rowIdx+1, email, err))
					continue
				}
				summary.CreatedCount++
			} else {
				summary.ErrorCount++
				summary.Errors = append(summary.Errors, fmt.Sprintf("Row %d (%s): failed to get student: %v", rowIdx+1, email, err))
				continue
			}
		}

		// Add to roster
		sid := pgtype.UUID{Bytes: student.StudentID.Bytes, Valid: true}
		_, err = h.q.AddToRoster(ctx, db.AddToRosterParams{
			ClassID:   cid,
			StudentID: sid,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				// Already in roster, skip
				summary.SkippedCount++
				continue
			}
			summary.ErrorCount++
			summary.Errors = append(summary.Errors, fmt.Sprintf("Row %d (%s): failed to add to roster: %v", rowIdx+1, email, err))
			continue
		}

		summary.AddedToRosterCount++
	}

	// Return summary
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summary)
}
