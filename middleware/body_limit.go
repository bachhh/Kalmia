package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

// BodyLimit creates a middleware that limits the request body size.
func BodyLimit(limitMB int64) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Calculate the limit in bytes.
			limitBytes := limitMB * 1024 * 1024

			// http.MaxBytesReader is the standard way to limit request body size.
			// It wraps the original request body and returns a new ReadCloser.
			// If the limit is exceeded during a read, it will return an error.
			r.Body = http.MaxBytesReader(w, r.Body, limitBytes)

			// Call the next handler in the chain.
			next.ServeHTTP(w, r)
		})
	}
}
