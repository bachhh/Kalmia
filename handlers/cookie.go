package handlers

import (
	"net/http"
	"time"

	"git.difuse.io/Difuse/kalmia/config"
	"git.difuse.io/Difuse/kalmia/consts"
	"git.difuse.io/Difuse/kalmia/utils"
)

const defaultMaxCookieExpireDays = 7

// This jwt cookie allow user to view any document, signed with a global jwt secret key
func SetDocAdminCookie(w http.ResponseWriter, cfg config.CookieConfig, jwtSecretKey string) error {
	// default ttl to 7 days
	jwtValue, _, err := utils.GenDocJWT(jwtSecretKey, 24*defaultMaxCookieExpireDays)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     consts.COOKIE_NAME_ADMIN_TOKEN,
		Domain:   cfg.Domain,
		Value:    jwtValue,
		Path:     cfg.Path,
		Expires:  time.Now().Add(24 * time.Hour * time.Duration(cfg.AgeDays)),
		Secure:   cfg.Secure,
		SameSite: http.SameSite(cfg.SameSite),
		HttpOnly: cfg.HttpOnly,
	}

	http.SetCookie(w, cookie)
	return nil
}

// Similar to doc admin cookie, but this cookie is limited to viewing per
// document, and is signed with the jwt secret key of the respective document.
func SetDocVisitorCookie(w http.ResponseWriter, cfg config.CookieConfig, jwtSecretKey string) error {
	// default ttl to 7 days
	jwtValue, _, err := utils.GenDocJWT(jwtSecretKey, 24*defaultMaxCookieExpireDays)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     consts.COOKIE_NAME_VISITOR_TOKEN,
		Domain:   cfg.Domain,
		Value:    jwtValue,
		Path:     cfg.Path,
		Expires:  time.Now().Add(24 * time.Hour * time.Duration(cfg.AgeDays)),
		Secure:   cfg.Secure,
		SameSite: http.SameSite(cfg.SameSite),
		HttpOnly: cfg.HttpOnly,
	}

	http.SetCookie(w, cookie)
	return nil
}
