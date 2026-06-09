package models

import (
	"time"

	"github.com/google/uuid"
)

// Auth
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
	User         *LoginResponseUser `json:"user"`
}

type LoginResponseUser struct {
	Id       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// Groups
type CreateGroupModel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GroupResponse struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	InviteCode  string    `json:"inviteCode"`
}

type JoinGroupModel struct {
	InviteCode string `json:"inviteCode"`
}

type UserResponse struct {
	Id        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type PatchGroupModel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateMemberRoleModel struct {
	Role string `json:"role"`
}

// Events

type UpdateEventModel struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	Status       string    `json:"status"`
	RsvpDeadline time.Time `json:"rsvpDeadline"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
}

type AddParticipantsModel struct {
	ParticipantIds []uuid.UUID `json:"participantIds"`
}
