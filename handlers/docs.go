package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.difuse.io/Difuse/kalmia/db/models"
	"git.difuse.io/Difuse/kalmia/logger"
	"git.difuse.io/Difuse/kalmia/services"
	"git.difuse.io/Difuse/kalmia/utils"
	"github.com/gorilla/mux"
)

func GetDocumentations(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	docs, err := service.GetDocumentations()
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	SendJSONResponse(http.StatusOK, w, docs)
}

func GetDocumentation(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID uint `json:"id" validate:"required"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	doc, err := service.GetDocumentation(req.ID)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	SendJSONResponse(http.StatusOK, w, doc)
}

func CreateDocumentation(service *services.ServiceRegistry, w http.ResponseWriter, r *http.Request) {
	token, err := GetTokenFromHeader(r)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_token"})
		return
	}

	user, err := service.AuthService.GetUserFromToken(token)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_token"})
		return
	}

	type Request struct {
		Name             string `json:"name" validate:"required"`
		Description      string `json:"description" validate:"required"`
		Version          string `json:"version" validate:"required"`
		URL              string `json:"url" validate:"required"`
		OrganizationName string `json:"organizationName" validate:"required"`
		LanderDetails    string `json:"landerDetails"`
		ProjectName      string `json:"projectName" validate:"required"`
		BaseURL          string `json:"baseURL" validate:"required"`
		Favicon          string `json:"favicon"`
		MetaImage        string `json:"metaImage"`
		NavImage         string `json:"navImage"`
		NavImageDark     string `json:"navImageDark"`
		CustomCSS        string `json:"customCSS" validate:"required"`
		FooterLabelLinks string `json:"footerLabelLinks"`
		MoreLabelLinks   string `json:"moreLabelLinks"`
		CopyrightText    string `json:"copyrightText" validate:"required"`
		RequireAuth      bool   `json:"requireAuth"`
		GitRepo          string `json:"gitRepo"`
		GitBranch        string `json:"gitBranch"`
		GitUser          string `json:"gitUser"`
		GitPassword      string `json:"gitPassword"`
		GitEmail         string `json:"gitEmail"`

		BucketFavicon      string `json:"bucketFavicon"`
		BucketMetaImage    string `json:"bucketMetaImage"`
		BucketNavImage     string `json:"bucketNavImage"`
		BucketNavImageDark string `json:"bucketNavImageDark"`
		TokenSecret        string `json:"tokenSecret"`
		RedirectURL        string `json:"RedirectURL"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	documentation := &models.Documentation{
		Name:             req.Name,
		Description:      req.Description,
		URL:              req.URL,
		OrganizationName: req.OrganizationName,
		LanderDetails:    req.LanderDetails,
		ProjectName:      req.ProjectName,
		BaseURL:          req.BaseURL,
		AuthorID:         user.ID,
		Author:           user,
		Editors:          []models.User{user},
		LastEditorID:     &user.ID,
		Version:          req.Version,
		Favicon:          req.Favicon,
		MetaImage:        req.MetaImage,
		NavImage:         req.NavImage,
		NavImageDark:     req.NavImageDark,
		CustomCSS:        req.CustomCSS,
		FooterLabelLinks: req.FooterLabelLinks,
		MoreLabelLinks:   req.MoreLabelLinks,
		CopyrightText:    req.CopyrightText,
		RequireAuth:      req.RequireAuth,
		GitRepo:          req.GitRepo,
		GitBranch:        req.GitBranch,
		GitUser:          req.GitUser,
		GitPassword:      req.GitPassword,
		GitEmail:         req.GitEmail,
		TokenSecret:      req.TokenSecret,
		RedirectURL:      req.RedirectURL,
	}

	err = service.DocService.CreateDocumentation(documentation, user, map[string]string{
		"favicon":      req.BucketFavicon,
		"metaImage":    req.BucketMetaImage,
		"navImage":     req.BucketNavImage,
		"navImageDark": req.BucketNavImageDark,
	})
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "documentation_created", "id": fmt.Sprint(documentation.ID)})
}

// special function for checking jwt token
func CheckJWTToken(srv *services.ServiceRegistry, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docIDStr, ok := vars["docID"]
	if !ok {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "doc_id not found"})
		return
	}

	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "cannot parse doc_id" + docIDStr + "as uint"})
		return
	}

	jwtToken, err := io.ReadAll(r.Body)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	doc, err := srv.DocService.GetDocumentation(uint(docID))
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	claims, err := utils.ValidateDocJWT(string(jwtToken), doc.TokenSecret)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}
	if claims.ExpiresAt.Time.After(time.Now()) {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "expired JWT token"})
		return
	}

	SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "success", "message": "none"})
}

func EditDocumentation(srv *services.ServiceRegistry, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID               uint   `json:"id" validate:"required"`
		Name             string `json:"name" validate:"required"`
		Description      string `json:"description" validate:"required"`
		URL              string `json:"url" validate:"required"`
		OrganizationName string `json:"organizationName" validate:"required"`
		LanderDetails    string `json:"landerDetails"`
		ProjectName      string `json:"projectName" validate:"required"`
		BaseURL          string `json:"baseURL" validate:"required"`
		Version          string `json:"version" validate:"required"`
		Favicon          string `json:"favicon"`
		MetaImage        string `json:"metaImage"`
		NavImage         string `json:"navImage"`
		NavImageDark     string `json:"navImageDark"`
		CustomCSS        string `json:"customCSS" validate:"required"`
		FooterLabelLinks string `json:"footerLabelLinks"`
		MoreLabelLinks   string `json:"moreLabelLinks"`
		CopyrightText    string `json:"copyrightText" validate:"required"`
		RequireAuth      bool   `json:"requireAuth"`
		GitRepo          string `json:"gitRepo"`
		GitBranch        string `json:"gitBranch"`
		GitEmail         string `json:"gitEmail"`
		GitUser          string `json:"gitUser"`
		GitPassword      string `json:"gitPassword"`

		BucketFavicon      string `json:"bucketFavicon"`
		BucketMetaImage    string `json:"bucketMetaImage"`
		BucketNavImage     string `json:"bucketNavImage"`
		BucketNavImageDark string `json:"bucketNavImageDark"`

		TokenSecret string `json:"tokenSecret"`
		RedirectURL string `json:"redirectURL"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	token, err := GetTokenFromHeader(r)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_request"})
		return
	}

	user, err := srv.AuthService.GetUserFromToken(token)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_request"})
		return
	}

	err = srv.DocService.EditDocumentation(
		services.EditDocumentationParams{
			User:             user,
			ID:               req.ID,
			Name:             req.Name,
			Description:      req.Description,
			Version:          req.Version,
			URL:              req.URL,
			BaseURL:          req.BaseURL,
			OrganizationName: req.OrganizationName,
			ProjectName:      req.ProjectName,
			RequireAuth:      req.RequireAuth,
			LanderDetails:    req.LanderDetails,
			CopyrightText:    req.CopyrightText,
			GitRepo:          req.GitRepo,
			GitBranch:        req.GitBranch,
			GitUser:          req.GitUser,
			GitPassword:      req.GitPassword,
			GitEmail:         req.GitEmail,
			Favicon:          req.Favicon,
			MetaImage:        req.MetaImage,
			NavImage:         req.NavImage,
			NavImageDark:     req.NavImageDark,
			CustomCSS:        req.CustomCSS,
			FooterLabelLinks: req.FooterLabelLinks,
			MoreLabelLinks:   req.MoreLabelLinks,
			BucketUploadedFiles: map[string]string{
				"favicon":      req.BucketFavicon,
				"metaImage":    req.BucketMetaImage,
				"navImage":     req.BucketNavImage,
				"navImageDark": req.BucketNavImageDark,
			},
			TokenSecret: req.TokenSecret,
			RedirectURL: req.RedirectURL,
		})
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "documentation_updated", "id": fmt.Sprint(req.ID)})
}

func DeleteDocumentation(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID uint `json:"id" validate:"required"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	err = service.DeleteDocumentation(req.ID)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "documentation_deleted", "id": fmt.Sprint(req.ID)})
}

func CreateDocumentationVersion(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		OriginalDocID uint   `json:"originalDocId" validate:"required"`
		NewVersion    string `json:"version" validate:"required"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	err = service.CreateDocumentationVersion(req.OriginalDocID, req.NewVersion)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "version_created"})
}

func GetPages(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	pages, err := service.GetPages()
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, pages)
}

func GetPage(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID uint `json:"id" validate:"required"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	page, err := service.GetPage(req.ID)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, page)
}

func CreatePage(services *services.ServiceRegistry, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Title           string `json:"title" validate:"required"`
		Slug            string `json:"slug" validate:"required"`
		Content         string `json:"content" validate:"required"`
		DocumentationID uint   `json:"documentationId" validate:"required"`
		PageGroupID     *uint  `json:"pageGroupId"`
		Order           *uint  `json:"order"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	token, err := GetTokenFromHeader(r)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_request"})
		return
	}

	user, err := services.AuthService.GetUserFromToken(token)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_request"})
		return
	}

	page := models.Page{
		Title:           req.Title,
		Slug:            req.Slug,
		Content:         req.Content,
		DocumentationID: req.DocumentationID,
		AuthorID:        user.ID,
		Author:          user,
		Editors:         []models.User{user},
		LastEditorID:    &user.ID,
	}

	if req.PageGroupID != nil {
		page.PageGroupID = req.PageGroupID
	}

	if req.Order != nil {
		page.Order = req.Order
	}

	err = services.DocService.CreatePage(&page)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "page_created", "id": fmt.Sprint(page.ID)})
}

func EditPage(services *services.ServiceRegistry, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID          uint   `json:"id" validate:"required"`
		Title       string `json:"title" validate:"required"`
		Slug        string `json:"slug" validate:"required"`
		Content     string `json:"content"`
		Order       *uint  `json:"order"`
		PageGroupId *uint  `json:"pageGroupId"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	token, err := GetTokenFromHeader(r)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_request"})
		return
	}

	user, err := services.AuthService.GetUserFromToken(token)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_request"})
		return
	}

	err = services.DocService.EditPage(user, req.ID, req.Title, req.Slug, req.Content, req.Order, req.PageGroupId)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "page_updated", "id": fmt.Sprint(req.ID)})
}

func DeletePage(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID uint `json:"id" validate:"required"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	err = service.DeletePage(req.ID)
	if err != nil {
		switch err.Error() {
		case "page_not_found":
			SendJSONResponse(http.StatusNotFound, w, map[string]string{"status": "error", "message": "Page not found"})
		default:
			logger.Error(err.Error())
			SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		}
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "page_deleted", "id": fmt.Sprint(req.ID)})
}

func GetPageGroups(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	pageGroups, err := service.GetPageGroups()
	if err != nil {
		logger.Error(err.Error())
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, pageGroups)
}

func GetPageGroup(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID uint `json:"id" validate:"required"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	pageGroup, err := service.GetPageGroup(req.ID)
	if err != nil {
		switch err.Error() {
		case "page_group_not_found":
			SendJSONResponse(http.StatusNotFound, w, map[string]string{"status": "error", "message": "Page group not found"})
		default:
			logger.Error(err.Error())
			SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		}
		return
	}

	SendJSONResponse(http.StatusOK, w, pageGroup)
}

func CreatePageGroup(services *services.ServiceRegistry, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Name            string `json:"name" validate:"required"`
		Label           string `json:"label" validate:"required"`
		DocumentationID uint   `json:"documentationId" validate:"required"`
		ParentID        *uint  `json:"parentId"`
		Order           *uint  `json:"order"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	token, err := GetTokenFromHeader(r)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_request"})
		return
	}

	user, err := services.AuthService.GetUserFromToken(token)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_request"})
		return
	}

	pageGroup := models.PageGroup{
		Name:            req.Name,
		Label:           req.Label,
		DocumentationID: req.DocumentationID,
		AuthorID:        user.ID,
		Author:          user,
		Editors:         []models.User{user},
		LastEditorID:    &user.ID,
	}

	if req.ParentID != nil {
		pageGroup.ParentID = req.ParentID
	}

	if req.Order != nil {
		pageGroup.Order = req.Order
	}

	_, err = services.DocService.CreatePageGroup(&pageGroup)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "page_group_created", "id": fmt.Sprint(pageGroup.ID)})
}

func EditPageGroup(services *services.ServiceRegistry, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID              uint   `json:"id" validate:"required"`
		Name            string `json:"name" validate:"required"`
		Label           string `json:"label" validate:"required"`
		DocumentationID uint   `json:"documentationId" validate:"required"`
		ParentID        *uint  `json:"parentId"`
		Order           *uint  `json:"order"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	token, err := GetTokenFromHeader(r)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_request"})
		return
	}

	user, err := services.AuthService.GetUserFromToken(token)
	if err != nil {
		SendJSONResponse(http.StatusUnauthorized, w, map[string]string{"status": "error", "message": "invalid_request"})
		return
	}

	err = services.DocService.EditPageGroup(user, req.ID, req.Name, req.Label, req.DocumentationID, req.ParentID, req.Order)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "page_group_updated", "id": fmt.Sprint(req.ID)})
}

func DeletePageGroup(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID uint `json:"id" validate:"required"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	err = service.DeletePageGroup(req.ID)
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "page_group_deleted", "id": fmt.Sprint(req.ID)})
}

func GetRsPress(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	// _, docPath, baseURL, _, _, err := service.GetRsPressFromURL(urlPath)
	docData, found, err := service.GetRsPressFromURL(urlPath)
	if err != nil || !found {
		http.Redirect(w, r, "/admin/", http.StatusTemporaryRedirect)
		return
	}

	fullPath := filepath.Join(docData.Path, strings.TrimPrefix(urlPath, docData.BaseURL))

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fullPath = filepath.Join(docData.Path, "index.html")
	}

	http.ServeFile(w, r, fullPath)
}

func BulkReorderPageOrPageGroup(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Order []struct {
			ID          uint  `json:"id" validate:"required"`
			Order       *uint `json:"order"`
			ParentID    *uint `json:"parentId"`
			PageGroupID *uint `json:"pageGroupId"`
			IsPageGroup bool  `json:"isPageGroup"`
		} `json:"order" validate:"required"`
	}

	req, err := ValidateRequest[Request](w, r)
	if err != nil {
		return
	}

	err = service.BulkReorderPageOrPageGroup(req.Order)
	if err != nil {
		logger.Error(err.Error())
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]string{"status": "success", "message": "pages_and_page_groups_reordered"})
}

func GetRootParentId(service *services.DocService, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "missing or invalid id"})
		return
	}

	documentationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendJSONResponse(http.StatusBadRequest, w, map[string]string{"status": "error", "message": "invalid id format"})
		return
	}

	rootParentID, err := service.GetRootParentID(uint(documentationID))
	if err != nil {
		SendJSONResponse(http.StatusInternalServerError, w, map[string]string{"status": "error", "message": err.Error()})
		return
	}

	SendJSONResponse(http.StatusOK, w, map[string]uint{"rootParentId": rootParentID})
}
