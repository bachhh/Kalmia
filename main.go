package main

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"git.difuse.io/Difuse/kalmia/cmd"
	"git.difuse.io/Difuse/kalmia/config"
	"git.difuse.io/Difuse/kalmia/db"
	"git.difuse.io/Difuse/kalmia/handlers"
	"git.difuse.io/Difuse/kalmia/logger"
	"git.difuse.io/Difuse/kalmia/middleware"
	"git.difuse.io/Difuse/kalmia/services"
	muxHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

//go:embed web/build
var adminFS embed.FS

func main() {
	cfgPath := cmd.ParseFlags()
	cfg := config.ParseConfig(cfgPath)
	logger.InitializeLogger(cfg.Environment, cfg.LogLevel, cfg.DataPath)

	/* Setup database */
	d := db.SetupDatabase(cfg.Environment, cfg.Database, cfg.DataPath)
	db.SetupBasicData(d, cfg.Admins)

	db.InitCache()

	serviceRegistry := services.NewServiceRegistry(d, cfg.LogSubCmd, cfg.Secret)
	authSrvc := serviceRegistry.AuthService
	docSrvc := serviceRegistry.DocService

	go func() {
		if err := docSrvc.StartupCheck(); err != nil {
			logger.Error("doc service failed startup check", zap.Error(err))
		}
		// start delete job and build job process every 10 seconds
		for {
			docSrvc.DeleteJob()
			docSrvc.BuildJob()
			time.Sleep(10 * time.Second)
		}
	}()

	/* Setup router */
	router := mux.NewRouter()
	router.Use(middleware.RecoverWithLog(logger.Logger))
	router.Use(middleware.BodyLimit(config.ParsedConfig.BodyLimitMb))
	router.Use(muxHandlers.CORS(config.ParsedConfig.Security.CORSConfig.ToMuxCORSOptions()...))

	kRouter := router.PathPrefix("/kal-api").Subrouter()

	// INFO: files could be fetched without authentication
	fileRouter := kRouter.PathPrefix("/file").Subrouter()

	fileRouter.HandleFunc("/get/{filename}", func(w http.ResponseWriter, r *http.Request) { handlers.GetFile(d, w, r, config.ParsedConfig) }).Methods("GET")

	/* Health endpoints */
	healthRouter := kRouter.PathPrefix("/health").Subrouter()
	healthRouter.HandleFunc("/ping", handlers.HealthPing).Methods("GET")
	healthRouter.HandleFunc("/last-trigger", func(w http.ResponseWriter, r *http.Request) { handlers.TriggerCheck(docSrvc, w, r) }).Methods("GET")

	oAuthRouter := kRouter.PathPrefix("/oauth").Subrouter()
	oAuthRouter.HandleFunc("/github", func(w http.ResponseWriter, r *http.Request) { handlers.GithubLogin(authSrvc, w, r) }).Methods("GET")
	oAuthRouter.HandleFunc("/github/callback", func(w http.ResponseWriter, r *http.Request) { handlers.GithubCallback(authSrvc, w, r) }).Methods("GET")
	oAuthRouter.HandleFunc("/microsoft", func(w http.ResponseWriter, r *http.Request) { handlers.MicrosoftLogin(authSrvc, w, r) }).Methods("GET")
	oAuthRouter.HandleFunc("/microsoft/callback", func(w http.ResponseWriter, r *http.Request) { handlers.MicrosoftCallback(authSrvc, w, r) }).Methods("GET")
	oAuthRouter.HandleFunc("/google", func(w http.ResponseWriter, r *http.Request) { handlers.GoogleLogin(authSrvc, w, r) }).Methods("GET")
	oAuthRouter.HandleFunc("/google/callback", func(w http.ResponseWriter, r *http.Request) { handlers.GoogleCallback(authSrvc, w, r) }).Methods("GET")
	oAuthRouter.HandleFunc("/providers", func(w http.ResponseWriter, r *http.Request) { handlers.GetOAuthProviders(authSrvc, w, r) }).Methods("GET")

	authRouter := kRouter.PathPrefix("/auth").Subrouter()
	authRouter.Use(middleware.EnsureAuthenticated(authSrvc))

	authRouter.HandleFunc("/user/create", func(w http.ResponseWriter, r *http.Request) { handlers.CreateUser(authSrvc, w, r) }).Methods("POST")
	authRouter.HandleFunc("/user/edit", func(w http.ResponseWriter, r *http.Request) { handlers.EditUser(authSrvc, w, r) }).Methods("POST")
	authRouter.HandleFunc("/user/delete", func(w http.ResponseWriter, r *http.Request) { handlers.DeleteUser(authSrvc, w, r) }).Methods("POST")

	authRouter.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) { handlers.GetUsers(authSrvc, w, r) }).Methods("GET")
	authRouter.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) { handlers.GetUser(authSrvc, w, r) }).Methods("POST")

	authRouter.HandleFunc("/user/upload-file", func(w http.ResponseWriter, r *http.Request) {
		handlers.UploadFile(serviceRegistry, d, w, r, config.ParsedConfig)
	}).Methods("POST")
	authRouter.HandleFunc("/user/assets/upload-file", func(w http.ResponseWriter, r *http.Request) {
		handlers.UploadAssetsFile(serviceRegistry, d, w, r, config.ParsedConfig)
	}).Methods("POST")

	authRouter.HandleFunc("/jwt/create", func(w http.ResponseWriter, r *http.Request) { handlers.CreateJWT(authSrvc, w, r) }).Methods("POST")
	authRouter.HandleFunc("/jwt/refresh", func(w http.ResponseWriter, r *http.Request) { handlers.RefreshJWT(authSrvc, w, r) }).Methods("POST")
	authRouter.HandleFunc("/jwt/validate", func(w http.ResponseWriter, r *http.Request) { handlers.ValidateJWT(authSrvc, w, r) }).Methods("POST")
	authRouter.HandleFunc("/jwt/revoke", func(w http.ResponseWriter, r *http.Request) { handlers.RevokeJWT(authSrvc, w, r) }).Methods("POST")

	docsRouter := kRouter.PathPrefix("/docs").Subrouter()
	docsRouter.Use(middleware.EnsureAuthenticated(authSrvc))
	docsRouter.HandleFunc("/documentations", func(w http.ResponseWriter, r *http.Request) { handlers.GetDocumentations(docSrvc, w, r) }).Methods("GET")
	docsRouter.HandleFunc("/documentation", func(w http.ResponseWriter, r *http.Request) { handlers.GetDocumentation(docSrvc, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/documentation/create", func(w http.ResponseWriter, r *http.Request) { handlers.CreateDocumentation(serviceRegistry, w, r) }).Methods("POST")

	docsRouter.HandleFunc("/documentation/{docID}/check_jwt", func(w http.ResponseWriter, r *http.Request) { handlers.CheckJWTToken(serviceRegistry, w, r) }).Methods("POST")

	docsRouter.HandleFunc("/documentation/edit", func(w http.ResponseWriter, r *http.Request) { handlers.EditDocumentation(serviceRegistry, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/documentation/delete", func(w http.ResponseWriter, r *http.Request) { handlers.DeleteDocumentation(docSrvc, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/documentation/version", func(w http.ResponseWriter, r *http.Request) { handlers.CreateDocumentationVersion(docSrvc, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/documentation/reorder-bulk", func(w http.ResponseWriter, r *http.Request) { handlers.BulkReorderPageOrPageGroup(docSrvc, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/documentation/root-parent-id", func(w http.ResponseWriter, r *http.Request) { handlers.GetRootParentId(docSrvc, w, r) }).Methods("GET")

	importRouter := docsRouter.PathPrefix("/import").Subrouter()
	importRouter.Use(middleware.EnsureAuthenticated(authSrvc))
	importRouter.HandleFunc("/gitbook", func(w http.ResponseWriter, r *http.Request) {
		handlers.ImportGitbook(serviceRegistry, w, r, config.ParsedConfig)
	}).Methods("POST")

	docsRouter.HandleFunc("/pages", func(w http.ResponseWriter, r *http.Request) { handlers.GetPages(docSrvc, w, r) }).Methods("GET")
	docsRouter.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) { handlers.GetPage(docSrvc, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/page/create", func(w http.ResponseWriter, r *http.Request) { handlers.CreatePage(serviceRegistry, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/page/edit", func(w http.ResponseWriter, r *http.Request) { handlers.EditPage(serviceRegistry, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/page/delete", func(w http.ResponseWriter, r *http.Request) { handlers.DeletePage(docSrvc, w, r) }).Methods("POST")

	docsRouter.HandleFunc("/page-groups", func(w http.ResponseWriter, r *http.Request) { handlers.GetPageGroups(docSrvc, w, r) }).Methods("GET")
	docsRouter.HandleFunc("/page-group", func(w http.ResponseWriter, r *http.Request) { handlers.GetPageGroup(docSrvc, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/page-group/create", func(w http.ResponseWriter, r *http.Request) { handlers.CreatePageGroup(serviceRegistry, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/page-group/edit", func(w http.ResponseWriter, r *http.Request) { handlers.EditPageGroup(serviceRegistry, w, r) }).Methods("POST")
	docsRouter.HandleFunc("/page-group/delete", func(w http.ResponseWriter, r *http.Request) { handlers.DeletePageGroup(docSrvc, w, r) }).Methods("POST")

	rsPressMiddleware := middleware.RsPressMiddleware(serviceRegistry)
	router.Use(rsPressMiddleware)

	spaHandler := createSPAHandler()
	router.PathPrefix("/").HandlerFunc(spaHandler)

	logger.Info("Starting server", zap.Int("port", cfg.Port))

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		logger.Fatal("srv.ListenAndServe", zap.Error(err))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	logger.Info("Shutting down server")
	os.Exit(0)
}

func createSPAHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/" {
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}

		if config.ParsedConfig.Environment != "dev" {
			file, err := adminFS.Open("web/build" + path)
			if err != nil {
				file, err = adminFS.Open("web/build/index.html")
				if err != nil {
					http.Error(w, "File not found", http.StatusNotFound)
					return
				}
			}
			defer file.Close()

			stat, err := file.Stat()
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if stat.IsDir() {
				file, err = adminFS.Open("web/build/index.html")
				if err != nil {
					http.Error(w, "File not found", http.StatusNotFound)
					return
				}
				defer file.Close()
			}

			if rs, ok := file.(io.ReadSeeker); ok {
				http.ServeContent(w, r, path, stat.ModTime(), rs)
			} else {
				http.Error(w, "file is not seekable", http.StatusInternalServerError)
			}
		} else {
			filePath := "web/build/" + path

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				filePath = "web/build/index.html"
			}

			http.ServeFile(w, r, filePath)
		}
	}
}
