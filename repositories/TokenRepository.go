package repositories

import (
	"auth-service/connections"
	"auth-service/models"
)

type TokenRepository interface {
	Create(t *models.Token) error
	DeleteByID(id uint) error
	FindByUserGUID(guid string) (*models.Token, error)
}

type tokenRepository struct{}

func NewTokenRepository() TokenRepository {
	return &tokenRepository{}
}

func (r *tokenRepository) Create(token *models.Token) error {
	return connections.DB.Create(token).Error
}

func (r *tokenRepository) DeleteByID(id uint) error {
	return connections.DB.Delete(&models.Token{}, id).Error
}

func (r *tokenRepository) FindByUserGUID(guid string) (*models.Token, error) {
	var token models.Token
	err := connections.DB.Where("user_guid = ?", guid).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}
