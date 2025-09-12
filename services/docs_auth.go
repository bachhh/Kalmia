package services

import "git.difuse.io/Difuse/kalmia/db/models"

func (service *DocService) CreateDocumentToken(docID int64) (token string, error) {
	// TODO
	return "", nil
}

func (service *DocService) CheckDocumentToken(docID int64, token string) (bool, error) {
	// TODO
	return false, nil
}
