package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTData struct {
	jwt.RegisteredClaims
	CustomClaims map[string]string `json:"custom_claims"`
	UserId       string            `json:"userId"`
	Username     string            `json:"username"`
	Email        string            `json:"email"`
	Photo        string            `json:"photo"`
	IsAdmin      bool              `json:"admin"`
	Permissions  []string          `json:"permissions"`
}

func GenerateJWTAccessToken(
	dbUserId uint,
	userId string,
	email string,
	photo string,
	isAdmin bool,
	permissions string,
	secretKey string,
) (string, int64, error) {
	permissionsSlice := []string{}

	if permissions != "" {
		err := json.Unmarshal([]byte(permissions), &permissionsSlice)
		if err != nil {
			return "", 0, err
		}
	}

	if len(permissionsSlice) == 0 {
		permissionsSlice = append(permissionsSlice, "read")
	}

	claims := JWTData{
		RegisteredClaims: jwt.RegisteredClaims{
			// TODO: make configurable
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
		CustomClaims: map[string]string{
			"user_id": userId,
			"email":   email,
		},
		UserId:      fmt.Sprintf("%d", dbUserId),
		Username:    userId,
		Email:       email,
		Photo:       photo,
		IsAdmin:     isAdmin,
		Permissions: permissionsSlice,
	}

	tokenString := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenString.SignedString([]byte(secretKey))
	if err != nil {
		return "", 0, err
	}

	expiry := claims.ExpiresAt.Time.Unix()

	return token, expiry, nil
}

func GetJWTExpirationTime(token string, secretKey string) (int64, error) {
	claims := &JWTData{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return 0, err
	}

	return claims.ExpiresAt.Time.Unix(), nil
}

func ValidateJWT(token string, secretKey string) (*JWTData, error) {
	claims := &JWTData{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func GetJWTUserId(token string, secretKey string) (string, error) {
	claims, err := ValidateJWT(token, secretKey)
	if err != nil {
		return "", err
	}

	return claims.UserId, nil
}

// special jwt for document auth, this is not related to user jwt auth
type DocJWTData struct {
	jwt.RegisteredClaims
}

func GenDocJWT(secretKey string, ttlHours int64) (string, int64, error) {
	claims := DocJWTData{
		RegisteredClaims: jwt.RegisteredClaims{
			// TODO: make configurable
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	tokenString := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenString.SignedString([]byte(secretKey))
	if err != nil {
		return "", 0, err
	}

	expiry := claims.ExpiresAt.Time.Unix()

	return token, expiry, nil
}

// validate and parse JWT claims for viewing document
func ValidateDocJWT(token string, secretKey string) (*DocJWTData, error) {
	claims := &DocJWTData{}
	_, err := jwt.ParseWithClaims(
		token,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
