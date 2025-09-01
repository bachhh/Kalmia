package models

import (
	"gorm.io/gorm"
)

// File stores metadata for files uploaded to S3.
// It embeds gorm.Model to include ID, CreatedAt, UpdatedAt, and DeletedAt fields automatically.
type File struct {
	gorm.Model
	FileName string `gorm:"index" json:"fileName,omitempty"`    // Original file name from the user.
	S3Key    string `gorm:"uniqueIndex" json:"s3Key,omitempty"` // The unique path to the object in the S3 bucket.
	URL      string `json:"url,omitempty"`                      // The public or pre-signed URL to access the file. Note that this is not valid if file access is set to "private" ( file can only be accessed via API )
	MIMEType string `json:"mimeType,omitempty"`                 // The content type, e.g., "image/png", "application/pdf".
	Size     int64  `json:"size,omitempty"`                     // File size in bytes.

	UploaderID uint `json:"uploaderId,omitempty"`
	Uploader   User `gorm:"foreignKey:UploaderID" json:"uploader,omitempty"`
}
