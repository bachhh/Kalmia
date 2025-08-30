package services

import (
	"sync"

	"gorm.io/gorm"
)

type DocService struct {
	DB          *gorm.DB
	UWBMutexMap sync.Map
	logSubCmd   bool
}

func NewDocService(db *gorm.DB, logSubCmd bool) *DocService {
	return &DocService{DB: db, logSubCmd: logSubCmd}
}
