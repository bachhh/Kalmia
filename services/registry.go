package services

import (
	"git.difuse.io/Difuse/kalmia/config"
	"gorm.io/gorm"
)

type ServiceRegistry struct {
	AuthService *AuthService
	DocService  *DocService
}

func NewServiceRegistry(db *gorm.DB, logSubCmd bool, secret config.Secret) *ServiceRegistry {
	return &ServiceRegistry{
		AuthService: NewAuthService(db, secret.JwtSecretKey),
		DocService:  NewDocService(db, logSubCmd),
	}
}
