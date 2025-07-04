package services

import (
	"auth-service/config"
	"auth-service/models"
	"auth-service/models/consts"
	"auth-service/repositories"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type TokenService struct {
	repo     repositories.TokenRepository
	secret   string
	issuer   string
	duration time.Duration
}

func NewTokenService(repo repositories.TokenRepository, c config.Config) *TokenService {
	return &TokenService{
		repo:     repo,
		secret:   c.Jwt.SecretKey,
		issuer:   c.Jwt.Issuer,
		duration: consts.TokenRefreshLifeTime,
	}
}

func (s *TokenService) DeleteTokenByID(id uint) error {
	return s.repo.DeleteByID(id)
}

func (s *TokenService) FindTokenByUserGUID(guid string) (*models.Token, error) {
	return s.repo.FindByUserGUID(guid)
}

func (s *TokenService) GenerateTokens(guid, userAgent, ip string) (string, string, error) {
	refreshToken, err := s.createRefreshToken()
	if err != nil {
		return "", "", err
	}

	accessToken, err := s.createAccessToken(guid, refreshToken)
	if err != nil {
		return "", "", err
	}

	hashedRefresh, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)

	token := &models.Token{
		UserGuid:     guid,
		RefreshToken: string(hashedRefresh),
		UserAgent:    userAgent,
		IpAddress:    ip,
		ExpiresAt:    time.Now().Add(s.duration),
	}

	err = s.repo.Create(token)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *TokenService) ValidateRefreshToken(hashedToken, inputToken string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(inputToken))
	return err == nil
}

func (s *TokenService) ValidateTokenPair(accessToken models.TokenClaims, refreshToken string) bool {
	hash := sha256.Sum256([]byte(refreshToken))
	currentRefreshSig := hex.EncodeToString(hash[:])[:8]
	return accessToken.RefreshSig == currentRefreshSig
}

func (s *TokenService) createRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *TokenService) createAccessToken(guid, refreshToken string) (string, error) {
	hash := sha256.Sum256([]byte(refreshToken))
	sig := hex.EncodeToString(hash[:])[:8]

	claims := jwt.MapClaims{
		"exp":         time.Now().Add(s.duration).Unix(),
		"sub":         guid,
		"iat":         time.Now().Unix(),
		"iss":         s.issuer,
		"refresh_sig": sig,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(s.secret))
}
