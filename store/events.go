package store

import (
	"context"
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
	Id           uuid.UUID
	GroupId      uuid.UUID
	CreatedBy    uuid.UUID
	Name         string
	Description  string
	Location     string
	Status       EventStatus
	CreatedAt    time.Time
	RsvpDeadline time.Time
	StartsAt     time.Time
	EndsAt       time.Time
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
	CreatedBy    uuid.UUID
	Name         string
	Description  string
	Location     string
	Status       EventStatus
	RsvpDeadline time.Time
	StartsAt     time.Time
	EndsAt       time.Time
}

func (s *Store) GetEvents(ctx context.Context, userId uuid.UUID) (*[]EventsRow, error) {
	return nil, nil
}

func (s *Store) GetEventById(ctx context.Context, eventId uuid.UUID) (*EventsRow, error) {
	return nil, nil
}

func (s *Store) GetEventParticipants(ctx context.Context, eventId uuid.UUID) (*[]UserRow, error) {
	return nil, nil
}

func (s *Store) Rsvp(ctx context.Context, userId uuid.UUID, eventId uuid.UUID, response ParticipantStatus) error {
	return nil
}

func (s *Store) UpdateEventInfo(ctx context.Context, eventId uuid.UUID, dto EventDto) error {
	return nil
}

func (s *Store) DeleteEvent(ctx context.Context, eventId uuid.UUID) error {
	return nil
}

func (s *Store) RemoveParticipant(ctx context.Context, eventId uuid.UUID, userId uuid.UUID) error {
	return nil
}

func (s *Store) AddParticipant(ctx context.Context, eventId uuid.UUID, userId uuid.UUID) error {
	return nil
}

func (s *Store) CreateEvent(ctx context.Context, dto EventDto) error {
	return nil
}

func (s *Store) GetGroupEvents(ctx context.Context, groupId uuid.UUID) (*[]EventsRow, error) {
	return nil, nil
}
