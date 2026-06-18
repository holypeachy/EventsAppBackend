package models

import (
	"time"

	"github.com/google/uuid"
)

type EventStatus string

const (
	EventRsvpOpen   EventStatus = "rsvp_open"
	EventRsvpClosed EventStatus = "rsvp_closed"
	EventCancelled  EventStatus = "cancelled"
	EventCompleted  EventStatus = "completed"
)

type ParticipantStatus string

const (
	PartInvited  ParticipantStatus = "invited"
	PartGoing    ParticipantStatus = "going"
	PartMaybe    ParticipantStatus = "maybe"
	PartDeclined ParticipantStatus = "declined"
)

type ParticipantRole string

const (
	EventOwner       ParticipantRole = "owner"
	EventManager     ParticipantRole = "manager"
	EventParticipant ParticipantRole = "participant"
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
	EventId     uuid.UUID         `json:"eventId"`
	UserId      uuid.UUID         `json:"userId"`
	Status      ParticipantStatus `json:"status"`
	Role        ParticipantRole   `json:"role"`
	CreatedAt   time.Time         `json:"createdAt"`
	RespondedAt time.Time         `json:"respondedAt"`
}

type ParticipantUserInfoRow struct {
	EventId     uuid.UUID         `json:"eventId"`
	UserId      uuid.UUID         `json:"userId"`
	Status      ParticipantStatus `json:"status"`
	Role        ParticipantRole   `json:"role"`
	CreatedAt   time.Time         `json:"createdAt"`
	RespondedAt time.Time         `json:"respondedAt"`

	Username string `json:"username"`
}

type EventModelDto struct {
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Location       string      `json:"location"`
	Status         EventStatus `json:"status"`
	RsvpDeadline   time.Time   `json:"rsvpDeadline"`
	StartsAt       time.Time   `json:"startsAt"`
	EndsAt         time.Time   `json:"endsAt"`
	ParticipantIds []uuid.UUID `json:"participantIds"`
}

type RsvpModel struct {
	Status string `json:"status"`
}

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
