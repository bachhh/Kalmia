package handlers

import (
	"net/http"
	"time"

	"git.difuse.io/Difuse/kalmia/consts"
	"git.difuse.io/Difuse/kalmia/utils"
)

const defaultMaxCookieExpireDays = 7

// This jwt cookie allow user to view any document, signed with a global jwt secret key
func SetDocAdminCookie(w http.ResponseWriter, jwtSecretKey string) error {
	// default ttl to 7 days
	jwtValue, _, err := utils.GenDocJWT(jwtSecretKey, 24*defaultMaxCookieExpireDays)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     consts.COOKIE_NAME_ADMIN_TOKEN,
		Domain:   ".localhost", // TODO: configurable
		Value:    jwtValue,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour * defaultMaxCookieExpireDays),
		Secure:   true,                  // TODO: configurable
		SameSite: http.SameSiteNoneMode, // TODO: configurable
		HttpOnly: true,                  // TODO: configurable

	}

	// TODO: add cookie to database
	http.SetCookie(w, cookie)

	return nil
}

// Similar to doc admin cookie, but this cookie is limited to viewing per
// document, and is signed with the jwt secret key of the respective document.
func SetDocVisitorCookie(w http.ResponseWriter, jwtSecretKey string) error {
	// default ttl to 7 days
	jwtValue, _, err := utils.GenDocJWT(jwtSecretKey, 24*defaultMaxCookieExpireDays)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     consts.COOKIE_NAME_VISITOR_TOKEN,
		Domain:   ".localhost", // TODO: configurable
		Value:    jwtValue,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour * defaultMaxCookieExpireDays),
		Secure:   true,                  // TODO: configurable
		SameSite: http.SameSiteNoneMode, // TODO: configurable
		HttpOnly: true,                  // TODO: configurable

	}

	// TODO: add cookie to database
	http.SetCookie(w, cookie)

	return nil
}
