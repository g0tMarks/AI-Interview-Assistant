// internal/api/router.go
package api

import (
	"net/http"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/api/handlers"
	"github.com/go-chi/chi/v5"
)

// NewRouter builds the chi router and registers all routes.
func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	// Middlewares (logging, recover, etc.) can go here later.
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)

	healthHandler := handlers.NewHealthHandler()
	rubricHandler := handlers.NewRubricHandler(deps.Queries)
	teacherHandler := handlers.NewTeacherHandler(deps.Queries)
	templateHandler := handlers.NewInterviewTemplateHandler(deps.Queries, deps.LLMService)
	interviewHandler := handlers.NewInterviewHandler(deps.Queries)
	studentHandler := handlers.NewStudentHandler(deps.Queries)
	classHandler := handlers.NewClassHandler(deps.Queries)
	rosterHandler := handlers.NewRosterHandler(deps.Queries)

	r.Get("/health", healthHandler.Health)
	r.Post("/rubrics", rubricHandler.CreateRubric)
	r.Get("/rubrics", rubricHandler.ListRubrics)
	r.Post("/teachers/register", teacherHandler.RegisterTeacher)
	r.Post("/interview-templates", templateHandler.CreateInterviewTemplate)
	r.Post("/interviews", interviewHandler.CreateInterview)
	r.Get("/interviews/{id}", interviewHandler.GetInterview)

	// Students
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

	return r
}
