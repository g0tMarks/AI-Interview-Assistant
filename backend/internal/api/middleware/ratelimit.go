package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type visitor struct {
	count int
	reset time.Time
}

// RateLimitIP returns middleware that limits each client IP to at most `limit`
// requests per `window`. When the limit is exceeded, it responds with 429.
func RateLimitIP(limit int, window time.Duration) func(http.Handler) http.Handler {
	if limit <= 0 || window <= 0 {
		// Degenerate configuration: no-op middleware.
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			now := time.Now()

			mu.Lock()
			v, ok := visitors[ip]
			if !ok || now.After(v.reset) {
				v = &visitor{
					count: 1,
					reset: now.Add(window),
				}
				visitors[ip] = v
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if v.count >= limit {
				retryAfter := int(time.Until(v.reset).Seconds())
				if retryAfter < 0 {
					retryAfter = 0
				}
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				mu.Unlock()
				return
			}

			v.count++
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the best-effort client IP from headers or RemoteAddr.
func clientIP(r *http.Request) string {
	// Prefer standard proxy headers when present.
	for _, header := range []string{"True-Client-IP", "X-Real-IP", "X-Forwarded-For"} {
		if v := strings.TrimSpace(r.Header.Get(header)); v != "" {
			// X-Forwarded-For can contain a list; take the first.
			if header == "X-Forwarded-For" {
				parts := strings.Split(v, ",")
				if len(parts) > 0 {
					return strings.TrimSpace(parts[0])
				}
			}
			return v
		}
	}

	// Fallback to RemoteAddr.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

