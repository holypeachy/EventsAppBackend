package models

import (
	"time"

	"github.com/google/uuid"
)

type GroupsRow struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	InviteCode  string    `json:"inviteCode"`
}

type GroupMembersRow struct {
	GroupId  uuid.UUID `json:"groupId"`
	UserId   uuid.UUID `json:"userId"`
	Role     GroupRole `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
}

type GroupRole string

const (
	GroupMember GroupRole = "member"
	GroupAdmin  GroupRole = "admin"
	GroupOwner  GroupRole = "owner"
)

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
