package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	// Requires a fully qualified module path
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/api"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/logger"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	//If using an underscore, it imports the package as well as runs its init() function
)

func main() {
	//Logging config
	os.Setenv("APPENV", "development")

	logger.Log.Info("Starting API...")
	logger.Log.Debug("Debugging active")

	// Load .env by walking up from cwd until we find it
	if cwd, err := os.Getwd(); err == nil {
		for dir := cwd; ; dir = filepath.Dir(dir) {
			envPath := filepath.Join(dir, ".env")
			if _, statErr := os.Stat(envPath); statErr == nil {
				if loadErr := godotenv.Load(envPath); loadErr == nil {
					log.Printf("Loaded .env from %s", envPath)
					break
				}
			}
			if parent := filepath.Dir(dir); parent == dir {
				break
			}
		}
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
	if os.Getenv("OPENAI_API_KEY") != "" || os.Getenv("ANTHROPIC_API_KEY") != "" {
		log.Printf("LLM API key configured")
	} else {
		log.Printf("WARNING: No LLM API key (OPENAI_API_KEY or ANTHROPIC_API_KEY) - rubric parse will fail")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
		logger.Log.Warn("JWT_SECRET not set; using default (do not use in production)")
	}

	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "uploads"
		logger.Log.Warn("UPLOADS_DIR not set; using default ./uploads (set to an absolute path in production)")
	}
	uploadsMaxBytes := int64(25 << 20) // 25 MiB
	if v := os.Getenv("UPLOADS_MAX_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			uploadsMaxBytes = n
		} else {
			logger.Log.Warnf("Invalid UPLOADS_MAX_BYTES=%q; using default", v)
		}
	}
	store := storage.NewLocalStore(uploadsDir)

	deps := api.Dependencies{
		Queries:         queries,
		LLMService:      llmService,
		TxBeginner:      conn,
		JWTSecret:       jwtSecret,
		Storage:         store,
		UploadsMaxBytes: uploadsMaxBytes,
	}
	srv := api.NewServer(deps)

	addr := ":8080"
	log.Printf("listening on %s", addr)

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}

}
