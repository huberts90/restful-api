package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// LoggingMiddleware creates a middleware that logs each HTTP request
// It captures the method, path, status code, and response time
func LoggingMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a wrapper for the response writer to capture the status code
			ww := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default to 200 OK
			}

			// Call the next handler
			next.ServeHTTP(ww, r)

			// Log the request
			duration := time.Since(start)
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Int("status", ww.statusCode),
				zap.Duration("duration", duration),
				zap.String("user_agent", r.UserAgent()),
			)
		})
	}
}

// responseWriterWrapper is a wrapper around http.ResponseWriter that captures the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before passing it to the wrapped ResponseWriter
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
