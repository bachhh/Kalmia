package middleware

import (
	"net/http"

	"git.difuse.io/Difuse/kalmia/handlers"
	"git.difuse.io/Difuse/kalmia/services"
)

func AuthenticateDocument(authSrv *services.AuthService, docSrv *services.DocService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. If user has a valid viewToken cookies -> user is an  editor -> automatically get a pass
			token, err := handlers.GetTokenFromHeader(r)
			if err == nil && authSrv.VerifyTokenInDb(token, false) {
				next.ServeHTTP(w, r)
				return
			}

			// 2. if user has Cookies token -> check against coresponding doc

			// 2. if not, check on ?token=... in paramter
			// - validate jwt token
			// - generate an auth cookies from secret from secret
			// - set said cookies

			next.ServeHTTP(w, r)
		})
	}
}
