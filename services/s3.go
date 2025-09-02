package services

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"git.difuse.io/Difuse/kalmia/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
)

type UploadFileResult struct {
	S3Key      string
	PublicURL  string // only valid
	PrivateURL string
}

// TODO: update the parameters to accept services.UploadedFileData{}
func UploadToS3Storage(
	file io.Reader,
	originalFilename, contentType string,
	parsedConfig *config.Config,
) (string, string, error) {
	key, err := uuid.NewV7()
	if err != nil {
		return "", "", err
	}
	ext := filepath.Ext(originalFilename)
	filename := fmt.Sprintf("upload-%s%s", key.String(), ext)
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"", originalFilename)

	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(parsedConfig.S3.Endpoint),
		Region:   aws.String(parsedConfig.S3.Region),
		Credentials: credentials.NewStaticCredentials(
			parsedConfig.S3.AccessKeyId,
			parsedConfig.S3.SecretAccessKey,
			"",
		),
		S3ForcePathStyle: aws.Bool(parsedConfig.S3.UsePathStyle),
	})
	if err != nil {
		return "", "", fmt.Errorf("error creating AWS session: %v", err)
	}

	svc := newS3Client(sess)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("error reading file: %v", err)
	}

	if ext == "" {
		// TODO: update this part to detect mimetype once on UploadFile()
		detectedMIME := mimetype.Detect(fileBytes)
		ext = detectedMIME.Extension()
		if contentType == "" {
			contentType = detectedMIME.String()
		}
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(parsedConfig.S3.Bucket),
		Key:                aws.String(filename),
		Body:               bytes.NewReader(fileBytes),
		ContentLength:      aws.Int64(int64(len(fileBytes))),
		ContentType:        aws.String(contentType),
		ContentDisposition: aws.String(contentDisposition), // <-- This is the new field
	})
	if err != nil {
		return "", "", fmt.Errorf("error uploading to S3-compatible storage: %v", err)
	}

	// NOTE: depending on system setting, can be private / public URL
	// - private object can only be proxy via API
	// - public object is accessed directly via S3 public URL
	var accessURL string
	if parsedConfig.AssetStorage == "local" {
		accessURL = fmt.Sprintf("http://localhost:%d/kal-api/file/get/%s", parsedConfig.Port, filename)
	} else {
		// TODO: use private URL
		accessURL = fmt.Sprintf(parsedConfig.S3.PublicUrlFormat, filename)
	}

	return filename, accessURL, nil
}

var newS3Client = func(sess *session.Session) s3iface.S3API {
	return s3.New(sess)
}
