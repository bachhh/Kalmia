package services

import "git.difuse.io/Difuse/kalmia/db/models"

func (service *AuthService) CreateDocumentToken(docID *uint) (string, error) {
	// TODO
	return "", nil
}

func (service *AuthService) CheckDocumentToken(docID int64, token string) (bool, error) {
	var docToken models.DocumentToken
	if err := service.DB.Where("token = ?", token).First(&docToken).Error; err != nil {
		return false, err
	}
	return false, nil
}
