package middleware

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"git.difuse.io/Difuse/kalmia/db"
	"git.difuse.io/Difuse/kalmia/services"
	"git.difuse.io/Difuse/kalmia/utils"
)

func RsPressMiddleware(dS *services.DocService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			urlPath := r.URL.Path
			docId, docPath, baseURL, reqAuth, err := dS.GetRsPress(urlPath)
			cookieToken := ""
			fmt.Println("_____1______")

			for _, cookie := range r.Cookies() {
				if cookie.Name == "viewToken" {
					cookieToken = cookie.Value
					break
				}
			}

			fmt.Println("_____2______")
			if err != nil {
				next.ServeHTTP(w, r)
				fmt.Println(err)
				return
			}

			fmt.Println("_____3______")
			fileKey := strings.TrimPrefix(urlPath, baseURL)
			fullPath := filepath.Join(docPath, fileKey)

			if strings.HasPrefix(fullPath, filepath.Join(docPath, "build")+string(filepath.Separator)) {
				fileKey = strings.TrimPrefix(fullPath, filepath.Join(docPath, "build")+string(filepath.Separator))
			}

			fmt.Println("_____4______")
			if strings.HasSuffix(fileKey, "guides.html") {
				fileKey = strings.TrimSuffix(fileKey, ".html")
			}

			fmt.Println("_____5______")
			if filepath.Ext(fileKey) == "" {
				fileKey = filepath.Join(fileKey, "index.html")
			}

			fmt.Println("_____6______")
			fileKey = fmt.Sprintf("rs|doc_%d|%s", docId, utils.TrimFirstRune(fileKey))
			value, err := db.GetValue([]byte(fileKey))
			if err == nil {
				if reqAuth && cookieToken == "" {
					http.Redirect(w, r, "/admin/login?docAuth="+utils.ToBase64(r.URL.Path), http.StatusTemporaryRedirect)
					return
				}
				w.Header().Set("Content-Type", value.ContentType)
				w.Write(value.Data)
				return
			}

			fmt.Println("_____7______")
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				fullPath = filepath.Join(docPath, "build", "index.html")
			}

			if reqAuth && cookieToken == "" {
				http.Redirect(w, r, "/admin/login?docAuth="+utils.ToBase64(r.URL.Path), http.StatusTemporaryRedirect)
				return
			} else {
				http.ServeFile(w, r, fullPath)
			}
		})
	}
}
