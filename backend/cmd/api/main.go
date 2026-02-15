package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Requires a fully qualified module path
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/api"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/logger"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	//If using an underscore, it imports the package as well as runs its init() function
)

func main() {
	//Logging config
	os.Setenv("APPENV", "development")

	logger.Log.Info("Starting API...")
	logger.Log.Debug("Debugging active")

	// Load .env file manually
	// Adjust path as needed because by default it looks in the current working directory for .env
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("No .env file found, using environment variables")
		fmt.Printf("%s\n", err)
	}

	dbURI := os.Getenv("DATABASE_URL")
	if dbURI == "" {
		dbURI = "postgres://postgres:mysecretpassword@localhost:5432/test-db?sslmode=disable"
	}

	fmt.Println("DB URI:", dbURI)

	// Connect to DB using pgx/v5
	// pgx.Connect() returns a *pgx.Conn which implements the DBTX interface required by db.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURI)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	//defer the close to the end of main()
	defer conn.Close(ctx)

	// Ping to verify connection
	if err := conn.Ping(ctx); err != nil {
		log.Fatal("db ping failed:", err)
	}

	fmt.Println("Connected to DB")

	//db.New() is from sqlc generated code in internal/db/db.go
	// It expects a DBTX interface which *pgx.Conn implements
	queries := db.New(conn)

	// Example: Use queries for database operations
	// You can now use queries to execute database queries
	logger.Log.Infof("Database queries initialised successfully: %v", queries != nil)

	// Initialize LLM service
	llmService := services.NewOpenAIService()
	logger.Log.Info("LLM service initialized")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
		logger.Log.Warn("JWT_SECRET not set; using default (do not use in production)")
	}

	deps := api.Dependencies{
		Queries:    queries,
		LLMService: llmService,
		JWTSecret:  jwtSecret,
	}
	srv := api.NewServer(deps)

	addr := ":8080"
	log.Printf("listening on %s", addr)

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}

}
