package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rakib/job-portal-api/internal/config"
)

type JWTClaims struct {
	Email     string  `json:"email"`
	Role      string  `json:"role"`
	CompanyID *string `json:"companyId,omitempty"`
	Name      string  `json:"name"`
	FirstName string  `json:"firstName"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userId string, email string, role string, companyId *string, name string, firstName string) (string, error) {
	cfg := config.AppConfig
	claims := JWTClaims{
		Email:     email,
		Role:      role,
		CompanyID: companyId,
		Name:      name,
		FirstName: firstName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWTAccessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTAccessSecret))
}

func GenerateRefreshToken(userId string) (string, error) {
	cfg := config.AppConfig
	claims := jwt.RegisteredClaims{
		Subject:   userId,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWTRefreshTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTRefreshSecret))
}

func VerifyAccessToken(tokenStr string) (*JWTClaims, error) {
	cfg := config.AppConfig
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWTAccessSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid access token")
	}

	return claims, nil
}

func VerifyRefreshToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	cfg := config.AppConfig
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWTRefreshSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	return claims, nil
}
