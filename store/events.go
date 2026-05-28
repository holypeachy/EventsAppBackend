package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/models"
)

func (s *Store) GetEvents(ctx context.Context, userId uuid.UUID) (*[]models.EventsRow, error) {
	return nil, nil
}

func (s *Store) GetEventById(ctx context.Context, eventId uuid.UUID) (*models.EventsRow, error) {
	return nil, nil
}

func (s *Store) GetEventParticipants(ctx context.Context, eventId uuid.UUID) (*[]models.UsersRow, error) {
	return nil, nil
}

func (s *Store) Rsvp(ctx context.Context, userId uuid.UUID, eventId uuid.UUID, response models.ParticipantStatus) error {
	return nil
}

func (s *Store) UpdateEventInfo(ctx context.Context, eventId uuid.UUID, dto models.EventModel) error {
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

func (s *Store) CreateEvent(ctx context.Context, groupId uuid.UUID, userId uuid.UUID, dto models.EventModel) (*models.EventsRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var event models.EventsRow
	row := tx.QueryRow(ctx, `
			INSERT INTO events(group_id, created_by, name, description, location, status,rsvp_deadline, starts_at, ends_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			RETURNING *
		`, groupId, userId, dto.Name, dto.Description, dto.Location, dto.Status, dto.RsvpDeadline, dto.StartsAt, dto.EndsAt)

	err = row.Scan(&event.Id, &event.GroupId, &event.CreatedBy, &event.Name, &event.Description, &event.Location, &event.Status, &event.CreatedAt, &event.RsvpDeadline, &event.StartsAt, &event.EndsAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
			INSERT INTO event_participants(event_id, user_id, status, role, responded_at)
			VALUES ($1, $2, $3, $4, $5)
		`, event.Id, userId, models.EventGoing, models.EventAdmin, time.Now())
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *Store) GetGroupEvents(ctx context.Context, groupId uuid.UUID) (*[]models.EventsRow, error) {
	// only if user is in event
	events := make([]models.EventsRow, 0)

	rows, err := s.pool.Query(ctx, `
			SELECT * FROM events
			WHERE group_id = $1
		`, groupId)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var event models.EventsRow

		err := rows.Scan(&event.Id, &event.GroupId, &event.CreatedBy, &event.Name, &event.Description, &event.Location, &event.Status, &event.CreatedAt, &event.RsvpDeadline, &event.StartsAt, &event.EndsAt)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return &events, nil
}
