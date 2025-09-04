package middleware

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ErrorLoggingMiddleware logs non-successful API calls.
func ErrorLoggingMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a response writer to capture the response
			rw := newResponseWriter(w)

			// Call the next handler
			next.ServeHTTP(rw, r)

			// Log only if the status code is 400 or greater
			if rw.statusCode >= 400 {
				logger.Error("API Error",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("statusCode", rw.statusCode),
					zap.String("responseBody", rw.body.String()),
					zap.String("userAgent", r.UserAgent()),
					zap.String("remoteAddr", r.RemoteAddr),
				)
			}
		})
	}
}

// responseWriter is a wrapper for http.ResponseWriter that allows the
// written HTTP status code and body to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

// newResponseWriter creates a new responseWriter.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

// WriteHeader captures the status code and calls the original WriteHeader.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body and calls the original Write.
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}
