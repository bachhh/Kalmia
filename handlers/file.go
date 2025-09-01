package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"git.difuse.io/Difuse/kalmia/config"
	"git.difuse.io/Difuse/kalmia/db/models"
	"git.difuse.io/Difuse/kalmia/logger"
	"git.difuse.io/Difuse/kalmia/services"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func GetFile(service *gorm.DB, w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	if len(filename) == 0 {
		SendJSONResponse(http.StatusNotFound, w, map[string]string{"status": "error", "message": "empty filename"})
		return
	}

	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(config.ParsedConfig.S3.Endpoint),
		Region:           aws.String(config.ParsedConfig.S3.Region),
		Credentials:      credentials.NewStaticCredentials(config.ParsedConfig.S3.AccessKeyId, config.ParsedConfig.S3.SecretAccessKey, ""),
		S3ForcePathStyle: aws.Bool(config.ParsedConfig.S3.UsePathStyle),
	})
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": "error creating s3 session: " + err.Error()})
		return
	}

	// create a new s3 client
	svc := s3.New(sess)

	// fileKey := strings.Split(strings.Split(filename, "-")[1], ".")[0]

	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("uploads"),
		Key:    aws.String(filename),
	})
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": "error getting object: " + err.Error()})
		return
	}

	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": "error reading object body: " + err.Error()})
		return
	}

	http.ServeContent(w, r, filename, time.Now(), bytes.NewReader(body))
	logger.Info("Successfully sent object file: " + filename)

	// INFO: use this if the above doesn't work
	// deletectedMime := mimetype.Detect(body)
	// w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	// w.Header().Set("Content-Type", deletectedMime.String())
	// w.WriteHeader(http.StatusOK)
	// w.Write(body)
}

func UploadFile(srv *services.ServiceRegistry, db *gorm.DB, w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	var err error
	token, err := GetTokenFromHeader(r)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_token"})
		return
	}

	user, err := srv.AuthService.GetUserFromToken(token)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_token"})
		return
	}

	// Capped at MaxFileSize set by the user
	err = r.ParseMultipartForm(cfg.MaxFileSize << 20)
	if err != nil {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "failed_to_parse_form"})
		return
	}

	file, header, err := r.FormFile("upload")
	if err != nil {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "failed_to_get_file"})
		return
	}
	defer file.Close()

	if header.Size > cfg.MaxFileSize<<20 {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "file_too_large"})
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": "failed_to_read_file"})
		return
	}

	contentType := http.DetectContentType(fileBytes)

	// Your S3 upload service should return the unique S3 key and the final URL.
	s3Key, fileURL, err := services.UploadToS3Storage(bytes.NewReader(fileBytes), header.Filename, contentType, cfg)
	if err != nil {
		logger.Error("error uploading to s3", zap.Error(err))
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": "failed_to_upload_file"})
		return
	}

	newFile := models.File{
		FileName:   header.Filename,
		S3Key:      s3Key,
		URL:        fileURL,
		MIMEType:   contentType,
		Size:       header.Size,
		UploaderID: user.ID,
	}

	// Create the record in the 'files' table
	result := db.Create(&newFile)
	if result.Error != nil {
		logger.Error("failed to create File record in database", zap.Error(result.Error))
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": "failed_to_save_file_metadata"})
		return
	}

	// Return the full database object for the newly created file.
	SendJSONResponse(http.StatusOK, w, map[string]interface{}{
		"status":  "success",
		"message": "file_uploaded",
		"file":    fileURL, // newFile object now includes the ID, CreatedAt, etc.
	})
}

// UploadAssetsFile handles uploading generic asset files, authenticates the user,
// and saves the file metadata to the database.
func UploadAssetsFile(srv *services.ServiceRegistry, db *gorm.DB, w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// 1. Authenticate the user (Copied from UploadFile)
	token, err := GetTokenFromHeader(r)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_token"})
		return
	}

	user, err := srv.AuthService.GetUserFromToken(token)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_token"})
		return
	}

	// 2. Parse the multipart form
	err = r.ParseMultipartForm(cfg.MaxFileSize << 20)
	if err != nil {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "failed_to_parse_form"})
		return
	}

	// This function retains its ability to use a dynamic form field name
	uploadTagName := r.FormValue("upload_tag_name")
	if uploadTagName == "" {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "missing_upload_tag_name"})
		return
	}

	file, header, err := r.FormFile(uploadTagName)
	if err != nil {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "failed_to_get_file"})
		return
	}
	defer file.Close()

	if header.Size > cfg.MaxFileSize<<20 {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "file_too_large"})
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": "failed_to_read_file"})
		return
	}

	contentType := http.DetectContentType(fileBytes)

	// 3. Upload to S3, capturing both the key and the URL
	s3Key, fileURL, err := services.UploadToS3Storage(bytes.NewReader(fileBytes), header.Filename, contentType, cfg)
	if err != nil {
		logger.Error("ERROR uploading file to S3", zap.Error(err))
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": "failed_to_upload_file"})
		return
	}

	// 4. Create the file record in the database (Copied from UploadFile)
	newFile := models.File{
		FileName:   header.Filename,
		S3Key:      s3Key,
		URL:        fileURL,
		MIMEType:   contentType,
		Size:       header.Size,
		UploaderID: user.ID,
	}

	result := db.Create(&newFile)
	if result.Error != nil {
		logger.Error("failed to create File record in database", zap.Error(result.Error))
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": "failed_to_save_file_metadata"})
		return
	}

	// strip file name from the bucket url and we will only need that here
	filePathSlices := strings.Split(fileURL, "/")
	bucketFileName := filePathSlices[len(filePathSlices)-1]

	// 5. Return the full database object as a structured response
	SendJSONResponse(http.StatusOK, w, map[string]interface{}{
		"status":  "success",
		"message": "file_uploaded",
		"file":    bucketFileName,
	})
}
