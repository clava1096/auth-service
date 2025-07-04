package models

import (
	"gorm.io/gorm"
	"time"
)

type TokenRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Token struct {
	gorm.Model
	UserGuid     string
	UserAgent    string
	IpAddress    string
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_in"`
}

func NewTokenResponse(access, refresh string) TokenResponse {
	return TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}
}
