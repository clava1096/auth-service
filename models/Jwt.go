package models

import (
	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	Exp        int64  `json:"exp"`
	Sub        string `json:"sub"`
	Iat        int64  `json:"iat"`
	Iss        string `json:"iss"`
	RefreshSig string `json:"refresh_sig"`
}

func GetClaims(accessToken, secretKey string) *TokenClaims {
	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil
	}
	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}
	result := &TokenClaims{
		Exp:        int64(payload["exp"].(float64)),
		Sub:        payload["sub"].(string),
		Iat:        int64(payload["iat"].(float64)),
		Iss:        payload["iss"].(string),
		RefreshSig: payload["refresh_sig"].(string),
	}
	return result
}
