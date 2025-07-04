package services

import (
	"auth-service/models"
	"auth-service/repositories"
	"github.com/google/uuid"
)

type UserService struct {
	repo repositories.UserRepository
}

func NewUserService(r repositories.UserRepository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) NewUsers(count int) ([]models.User, error) {
	users := make([]models.User, count)
	for i := 0; i < count; i++ {
		users[i].Guid = uuid.New().String()
		err := s.repo.Create(&users[i])
		if err != nil {
			return users, err
		}
	}
	return users, nil
}

func (s *UserService) IsExist(guid string) bool {
	return s.repo.IsExist(guid)
}

func (s *UserService) GetUsers() ([]models.UserResponse, error) {
	users, err := s.repo.GetUsers()
	if err != nil {
		return nil, err
	}

	result := make([]models.UserResponse, len(users))
	for i, u := range users {
		result[i] = models.NewUserResponse(u.Guid)
	}
	return result, nil
}
