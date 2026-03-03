package middleware

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"git.difuse.io/Difuse/kalmia/config"
	"git.difuse.io/Difuse/kalmia/consts"
	"git.difuse.io/Difuse/kalmia/db"
	"git.difuse.io/Difuse/kalmia/handlers"
	"git.difuse.io/Difuse/kalmia/logger"
	"git.difuse.io/Difuse/kalmia/services"
	"git.difuse.io/Difuse/kalmia/utils"
	"go.uber.org/zap"
)

func RsPressMiddleware(srv *services.ServiceRegistry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			urlPath := r.URL.Path
			// TODO:  disable doc's auth should be opt-out
			docData, found, err := srv.DocService.GetRsPressFromURL(urlPath)
			if err != nil {
				logger.Error("error while fetching doc rspress from url", zap.Error(err))
				handlers.SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
				return
			} else if !found {
				next.ServeHTTP(w, r)
				return
			}

			isAuthorized := false
			// 1. check if user has admin cookie
			for _, cookie := range r.CookiesNamed(consts.COOKIE_NAME_ADMIN_TOKEN) {
				if cookie.Value != "" {
					// NOTE: here we validate admin cookie with the global jwt key
					if _, err = utils.ValidateDocJWT(cookie.Value, srv.AuthService.JwtSecretKey); err == nil {
						isAuthorized = true
						logger.Debug("access granted via adminToken")
						break
					} else {
						logger.Error(fmt.Sprintf("invalid cookie token for doc %d", docData.ID), zap.Error(err))
						continue
					}
				}
			}

			// 2. check if user has visitor cookie
			if !isAuthorized {
				for _, cookie := range r.CookiesNamed(consts.COOKIE_NAME_VISITOR_TOKEN) {
					if cookie.Value != "" {
						// NOTE: here we validate admin cookie with the respective key jwt key
						if _, err := srv.AuthService.VerifyDocumentJWT(cookie.Value, docData.ID); err == nil {
							isAuthorized = true
							logger.Debug("access granted via visitorToken")
							break
						} else {
							logger.Error(fmt.Sprintf("invalid cookie token for doc %d", docData.ID), zap.Error(err))
							continue
						}
					}
				}
			}

			// 3. if we are still not authorized, check and validate jwt_token in url param
			if !isAuthorized {
				if jwtToken := r.URL.Query().Get("jwt_token"); jwtToken != "" {
					// NOTE: here we validate the jwt to
					if _, err := srv.AuthService.VerifyDocumentJWT(jwtToken, docData.ID); err == nil {
						// NOTE: convert to visitor cookie so that user don't have to use jwt again in future

						//nolint:errcheck
						_ = handlers.SetDocVisitorCookie(w, config.ParsedConfig.Security.CookieConfig, docData.TokenSecret)
						isAuthorized = true
						logger.Debug("access granted via jwt_token")

					} else {
						logger.Error("validate visit jwt token failed",
							zap.Uint("doc_id", docData.ID),
							zap.Error(err))
					}
				}
			}

			if docData.RequireAuth && !isAuthorized {
				// TODO: redirect user to a predefined URL set for each documentation
				logger.Error("attempted unauthorized access to document",
					zap.Uint("doc_id", docData.ID))
				handlers.SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"error": "user_unauthorized_route"})
				return
			}

			fileKey := strings.TrimPrefix(urlPath, docData.BaseURL)
			fullPath := filepath.Join(docData.Path, fileKey)

			if strings.HasPrefix(fullPath, filepath.Join(docData.Path, "build")+string(filepath.Separator)) {
				fileKey = strings.TrimPrefix(fullPath, filepath.Join(docData.Path, "build")+string(filepath.Separator))
			}

			if strings.HasSuffix(fileKey, "guides.html") {
				fileKey = strings.TrimSuffix(fileKey, ".html")
			}

			if filepath.Ext(fileKey) == "" {
				fileKey = filepath.Join(fileKey, "index.html")
			}

			fileKey = fmt.Sprintf("rs|doc_%d|%s", docData.ID, utils.TrimFirstRune(fileKey))
			value, err := db.GetValue([]byte(fileKey))
			if err == nil {
				w.Header().Set("Content-Type", value.ContentType)
				//nolint:errcheck
				_, _ = w.Write(value.Data)
				return
			}

			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				fullPath = filepath.Join(docData.Path, "build", "index.html")
			}

			http.ServeFile(w, r, fullPath)
		})
	}
}
