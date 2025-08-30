package services

import "gorm.io/gorm"

type ServiceRegistry struct {
	AuthService *AuthService
	DocService  *DocService
}

func NewServiceRegistry(db *gorm.DB, logSubCmd bool) *ServiceRegistry {
	return &ServiceRegistry{
		AuthService: NewAuthService(db),
		DocService:  NewDocService(db, logSubCmd),
	}
}
