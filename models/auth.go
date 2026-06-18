package models

import (
	"time"

	"github.com/google/uuid"
)

type UsersRow struct {
	Id           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type RefreshTokensRow struct {
	Id         uuid.UUID
	UserId     uuid.UUID
	TokenHash  string
	ExpiresAt  time.Time
	LastUsedAt time.Time
	CreatedAt  time.Time
}

type RegisterModel struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginModel struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshModel struct {
	RefreshToken string `json:"refreshToken"`
}

type LogoutModel struct {
	RefreshToken string `json:"refreshToken"`
}

type LoginResponse struct {
	AccessToken  string             `json:"accessToken"`
	RefreshToken string             `json:"refreshToken"`
	User         *LoginUserResponse `json:"user"`
}

type LoginUserResponse struct {
	Id       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}
