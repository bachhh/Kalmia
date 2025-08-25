package handlers

import (
	"encoding/json"
	"net/http"

	"git.difuse.io/Difuse/kalmia/config"
	"git.difuse.io/Difuse/kalmia/services"
)

func ImportGitbook(services *services.ServiceRegistry, w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	var request struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "invalid_url", "error": err.Error()})
		return
	}

	if request.URL == "" {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "url_required"})
		return
	}

	jsonString, err := services.DocService.ImportGitbook(request.URL, request.Username, request.Password, cfg)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": "gitbook_proccessing_failed", "error": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, jsonString)
}

const maxBody = 50 << 20

func ImportFolder(services *services.ServiceRegistry, w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// 100MB limit
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxBody); err != nil {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{
			"status":  "error",
			"message": "invalid_form_data",
			"error":   err.Error(),
		})
		return
	}

	// Retrieve file from form field "file"
	file, _, err := r.FormFile("file")
	if err != nil {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{
			"status":  "error",
			"message": "file_required",
			"error":   err.Error(),
		})
		return
	}
	defer file.Close()

	// Call your service with the uploaded file reader
	jsonString, err := services.DocService.ImportGitbookFolder(file)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{
			"status":  "error",
			"message": "gitbook_folder_processing_failed",
			"error":   err.Error(),
		})
		return
	}

	// Success
	SendJSONResponse(http.StatusOK, w, jsonString)
}
