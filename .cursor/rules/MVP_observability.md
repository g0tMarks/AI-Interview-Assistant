# Observability & MVP Rule

## Purpose
Ensure the codebase is designed so that **metrics and tracing can be added later with minimal refactoring**, even though the MVP does **not** implement full observability.

---

## Rules

### 1. Context Propagation
- All handlers, services, and database methods **must accept `context.Context`**.
- Never create empty contexts (`context.Background()`) inside business logic.
- Pass the existing context through all layers to allow future tracing.

### 2. Middleware-First Architecture
- All HTTP requests must flow through a **central middleware stack**.
- Keep cross-cutting concerns (logging, request IDs, recovery) in middleware.
- This enables future drop-in middleware for:
  - OpenTelemetry tracing
  - Prometheus metrics
  - Request timing
  - Error tracking

### 3. Structured Logging
- Use a **structured logger** (zap, zerolog, slog, or similar).
- Logging should occur through a single logger interface where possible.
- Include request IDs and contextual metadata from `context.Context`.
- Avoid `fmt.Printf` or unstructured logs.

### 4. Instrumentation-Ready Database Layer
- Use a **single pgx pool (`*pgxpool.Pool`)** entrypoint.
- Do not open ad-hoc connections or embed connection logic inside handlers.
- This allows future instrumentation via otelsql or pgx tracing hooks.

### 5. Minimal MVP Requirements
Even without metrics/tracing in the MVP, ensure:
- Structured logs exist.
- Every request gets a generated request ID.
- Context is passed through the stack.
- Observability-related code lives in its own package (e.g. `internal/observability/`) even if thin.

### 6. Future Expansion Notes (for Cursor)
When evolving the project, Cursor should prefer:
- Adding OpenTelemetry SDK setup in `internal/observability/otel.go`
- Adding Prometheus metrics via middleware in `internal/http/middleware/metrics.go`
- Wrapping the pgx connection with instrumentation in `internal/db/instrumentation.go`

---

## Why These Rules Exist
The MVP intentionally avoids full metrics + tracing, but these practices ensure that adding them later requires **no structural changes**, only additional dependencies and middleware.

This rule helps maintain long-term scalability while keeping MVP development fast and simple.
