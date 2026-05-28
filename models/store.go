package models

import (
	"time"

	"github.com/google/uuid"
)

// Auth
type UsersRow struct {
	Id           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type RefreshTokenRow struct {
	Id         uuid.UUID
	UserId     uuid.UUID
	TokenHash  string
	ExpiresAt  time.Time
	LastUsedAt time.Time
	CreatedAt  time.Time
}

// Groups
type GroupsRow struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	InviteCode  string    `json:"inviteCode"`
}

type GroupMemberRow struct {
	GroupId  uuid.UUID
	UserId   uuid.UUID
	Role     GroupRole
	JoinedAt time.Time
}

type GroupRole string

const (
	Member GroupRole = "member"
	Admin  GroupRole = "admin"
	Owner  GroupRole = "owner"
)

// Events
type EventStatus string

const (
	EventRsvpOpen   EventStatus = "rsvp_open"
	EventRsvpClosed EventStatus = "rsvp_closed"
	EventCancelled  EventStatus = "cancelled"
	EventCompleted  EventStatus = "completed"
)

type ParticipantStatus string

const (
	EventInvited  ParticipantStatus = "invited"
	EventGoing    ParticipantStatus = "going"
	EventMaybe    ParticipantStatus = "maybe"
	EventDeclined ParticipantStatus = "declined"
)

type ParticipantRole string

const (
	EventParticipant ParticipantRole = "participant"
	EventAdmin       ParticipantRole = "admin"
)

type EventsRow struct {
	Id           uuid.UUID   `json:"id"`
	GroupId      uuid.UUID   `json:"groupId"`
	CreatedBy    uuid.UUID   `json:"createdBy"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Location     string      `json:"location"`
	Status       EventStatus `json:"status"`
	CreatedAt    time.Time   `json:"createdAt"`
	RsvpDeadline time.Time   `json:"rsvpDeadline"`
	StartsAt     time.Time   `json:"startsAt"`
	EndsAt       time.Time   `json:"endsAt"`
}

type EventParticipantsRow struct {
	EventId     uuid.UUID
	UserId      uuid.UUID
	Status      ParticipantStatus
	Role        ParticipantRole
	CreatedAt   time.Time
	RespondedAt time.Time
}

type EventDto struct {
	GroupId      uuid.UUID
	CreatedBy    uuid.UUID
	Name         string
	Description  string
	Location     string
	Status       EventStatus
	RsvpDeadline time.Time
	StartsAt     time.Time
	EndsAt       time.Time
}

type EventModel struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Location     string      `json:"location"`
	Status       EventStatus `json:"status"`
	RsvpDeadline time.Time   `json:"rsvpDeadline"`
	StartsAt     time.Time   `json:"startsAt"`
	EndsAt       time.Time   `json:"endsAt"`
}
