// internal/api/router.go
package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/api/handlers"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/api/middleware"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/engine"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/evaluation"
)

// NewRouter builds the chi router and registers all routes.
func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	// Global middlewares.
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)
	// Basic IP-based rate limit to protect the API and LLM-backed endpoints.
	r.Use(middleware.RateLimitIP(100, time.Minute))

	healthHandler := handlers.NewHealthHandler()
	rubricHandler := handlers.NewRubricHandler(deps.Queries, deps.LLMService, deps.TxBeginner)
	teacherHandler := handlers.NewTeacherHandler(deps.Queries)
	templateHandler := handlers.NewInterviewTemplateHandler(deps.Queries, deps.LLMService)
	interviewEngine := engine.NewEngine(deps.Queries, deps.LLMService)
	evalRunner := evaluation.NewRunner(deps.Queries, deps.LLMService)
	interviewHandler := handlers.NewInterviewHandler(deps.Queries, interviewEngine, evalRunner)
	submissionHandler := handlers.NewSubmissionHandler(deps.Queries, deps.LLMService, interviewHandler)
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
	r.Get("/teachers/{id}/results", teacherHandler.ListResults)
	r.Post("/interview-templates", templateHandler.CreateInterviewTemplate)
	r.Post("/interviews", interviewHandler.CreateInterview)
	r.Get("/interviews/{id}", interviewHandler.GetInterview)
	r.Post("/interviews/{id}/messages", interviewHandler.CreateMessage)
	r.Get("/interviews/{id}/messages", interviewHandler.ListMessages)
	r.Get("/interviews/{id}/next", interviewHandler.GetNext)
	r.Post("/interviews/{id}/next", interviewHandler.PostNext)
	r.Get("/interviews/{id}/results", interviewHandler.GetResults)
	r.Get("/interviews/{id}/summary", interviewHandler.GetSummary)
	r.Post("/uploads", uploadHandler.Upload)
	r.Get("/uploads/{key}", uploadHandler.Download)

	// Submissions (authorship workflow)
	r.Post("/submissions", submissionHandler.CreateSubmission)
	r.Get("/submissions/{id}", submissionHandler.GetSubmission)
	r.Post("/submissions/{id}/artifacts", submissionHandler.CreateArtifact)
	r.Get("/submissions/{id}/artifacts", submissionHandler.ListArtifacts)
	r.Post("/submissions/{id}/viva/start", submissionHandler.StartViva)
	r.Get("/submissions/{id}/viva", submissionHandler.GetViva)
	r.Post("/submissions/{id}/viva/messages", submissionHandler.VivaMessages)
	r.Get("/submissions/{id}/viva/messages", submissionHandler.ListVivaMessages)
	r.Post("/submissions/{id}/authorship/run", submissionHandler.RunAuthorship)
	r.Get("/submissions/{id}/authorship", submissionHandler.GetAuthorship)

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
	r.Post("/classes/{id}/interviews/bulk", classHandler.BulkCreateInterviews)

	// Roster (students in a class)
	r.Get("/classes/{classId}/roster", rosterHandler.ListRoster)
	r.Post("/classes/{classId}/roster", rosterHandler.AddToRoster)
	r.Delete("/classes/{classId}/roster/{studentId}", rosterHandler.RemoveFromRoster)
	r.Post("/classes/{id}/roster/upload", rosterHandler.UploadRoster)

	return r
}
