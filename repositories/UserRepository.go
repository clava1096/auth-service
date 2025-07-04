package repositories

import (
	"auth-service/connections"
	"auth-service/models"
)

type UserRepository interface {
	Create(u *models.User) error
	IsExist(guid string) bool
	GetUsers() ([]models.User, error)
}

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) Create(user *models.User) error {
	return connections.DB.Create(user).Error
}

func (r *userRepository) IsExist(guid string) bool {
	var exists bool
	connections.DB.Model(&models.User{}).
		Select("count(*) > 0").
		Where("guid = ?", guid).
		Find(&exists)
	return exists
}

func (r *userRepository) GetUsers() ([]models.User, error) {
	var users []models.User
	err := connections.DB.Find(&users).Error
	return users, err
}
