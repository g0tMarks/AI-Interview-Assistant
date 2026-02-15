package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

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
