// internal/api/router.go
package api

import (
	"net/http"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/api/handlers"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/api/middleware"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/engine"
	"github.com/go-chi/chi/v5"
)

// NewRouter builds the chi router and registers all routes.
func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	// Middlewares (logging, recover, etc.) can go here later.
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)

	healthHandler := handlers.NewHealthHandler()
	rubricHandler := handlers.NewRubricHandler(deps.Queries, deps.LLMService, deps.TxBeginner)
	teacherHandler := handlers.NewTeacherHandler(deps.Queries)
	templateHandler := handlers.NewInterviewTemplateHandler(deps.Queries, deps.LLMService)
	interviewEngine := engine.NewEngine(deps.Queries, deps.LLMService)
	interviewHandler := handlers.NewInterviewHandler(deps.Queries, interviewEngine)
	studentHandler := handlers.NewStudentHandler(deps.Queries)
	classHandler := handlers.NewClassHandler(deps.Queries)
	rosterHandler := handlers.NewRosterHandler(deps.Queries)
	authHandler := handlers.NewAuthHandler(deps.Queries, deps.JWTSecret)
	uploadHandler := handlers.NewUploadHandler(deps.Storage, deps.UploadsMaxBytes)

	r.Get("/health", healthHandler.Health)
	r.Post("/rubrics", rubricHandler.CreateRubric)
	r.Post("/rubrics/upload", rubricHandler.UploadRubricFile)
	r.Get("/rubrics", rubricHandler.ListRubrics)
	r.Patch("/rubrics/{id}", rubricHandler.PatchRubric)
	r.Post("/rubrics/{id}/parse", rubricHandler.ParseRubric)
	r.Put("/rubrics/{id}/criteria-and-plan", rubricHandler.PutCriteriaAndPlan)
	r.Post("/teachers/register", teacherHandler.RegisterTeacher)
	r.Post("/interview-templates", templateHandler.CreateInterviewTemplate)
	r.Post("/interviews", interviewHandler.CreateInterview)
	r.Get("/interviews/{id}", interviewHandler.GetInterview)
	r.Post("/interviews/{id}/messages", interviewHandler.CreateMessage)
	r.Get("/interviews/{id}/messages", interviewHandler.ListMessages)
	r.Get("/interviews/{id}/next", interviewHandler.GetNext)
	r.Post("/interviews/{id}/next", interviewHandler.PostNext)
	r.Post("/uploads", uploadHandler.Upload)
	r.Get("/uploads/{key}", uploadHandler.Download)

	// Student auth (class code + email → JWT)
	r.Post("/auth/student/login", authHandler.StudentLogin)

	// Student-facing routes (require valid student JWT)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireStudentAuth(deps.JWTSecret))
		r.Get("/student/me", studentHandler.GetMe)
	})

	// Students (admin/teacher CRUD; unauthenticated for now)
	r.Post("/students", studentHandler.CreateStudent)
	r.Get("/students", studentHandler.ListStudents)
	r.Get("/students/{id}", studentHandler.GetStudent)
	r.Patch("/students/{id}", studentHandler.UpdateStudent)

	// Classes (teacher-scoped list via ?teacherId=)
	r.Post("/classes", classHandler.CreateClass)
	r.Get("/classes", classHandler.ListClasses)
	r.Get("/classes/{id}", classHandler.GetClass)
	r.Patch("/classes/{id}", classHandler.UpdateClass)
	r.Delete("/classes/{id}", classHandler.DeleteClass)

	// Roster (students in a class)
	r.Get("/classes/{classId}/roster", rosterHandler.ListRoster)
	r.Post("/classes/{classId}/roster", rosterHandler.AddToRoster)
	r.Delete("/classes/{classId}/roster/{studentId}", rosterHandler.RemoveFromRoster)
	r.Post("/classes/{id}/roster/upload", rosterHandler.UploadRoster)

	return r
}
