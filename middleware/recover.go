package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func RecoverWithLog(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// A panic occurred. Log it with zap.
					logger.Error("Panic recovered",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						// zap.Stack("stacktrace") captures the stack trace automatically.
						zap.Stack("stacktrace"),
					)

					// Respond with a 500 Internal Server Error.
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			// Call the next handler in the chain.
			next.ServeHTTP(w, r)
		})
	}
}
