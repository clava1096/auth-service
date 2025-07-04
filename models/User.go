package models

import (
	"time"
)

type UserResponse struct {
	Guid string `json:"guid"`
}

type Logout struct {
	Msg string `json:"msg"`
}

type User struct {
	Guid      string `gorm:"unique;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `gorm:"index"`
}

func NewUserResponse(guid string) UserResponse {
	return UserResponse{
		Guid: guid,
	}
}
