package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/helpers"
	"github.com/holypeachy/EventsAppBackend/models"
)

func (s *Store) GetEvents(ctx context.Context, userId uuid.UUID) ([]models.EventsRow, error) {
	rows, err := s.pool.Query(ctx, `
			SELECT e.id, e.group_id, e.created_by, e.name, e.description, e.location, e.status, e.created_at, e.rsvp_deadline, e.starts_at, e.ends_at
			FROM events e
			JOIN event_participants ep
				ON ep.event_id = e.id
			WHERE ep.user_id = $1
		`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]models.EventsRow, 0)

	for rows.Next() {
		var event models.EventsRow

		err := rows.Scan(
			&event.Id,
			&event.GroupId,
			&event.CreatedBy,
			&event.Name,
			&event.Description,
			&event.Location,
			&event.Status,
			&event.CreatedAt,
			&event.RsvpDeadline,
			&event.StartsAt,
			&event.EndsAt,
		)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

func (s *Store) GetEventById(ctx context.Context, eventId uuid.UUID) (*models.EventsRow, error) {
	row := s.pool.QueryRow(ctx, `
			SELECT * FROM events
			WHERE id = $1
		`, eventId)
	var event models.EventsRow
	err := row.Scan(
		&event.Id,
		&event.GroupId,
		&event.CreatedBy,
		&event.Name,
		&event.Description,
		&event.Location,
		&event.Status,
		&event.CreatedAt,
		&event.RsvpDeadline,
		&event.StartsAt,
		&event.EndsAt,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *Store) GetEventParticipants(ctx context.Context, eventId uuid.UUID) ([]models.ParticipantUserInfoRow, error) {
	rows, err := s.pool.Query(ctx, `
			SELECT ep.event_id, ep.user_id, ep.status, ep.role, ep.created_at, ep.responded_at, u.username
			FROM event_participants ep
			JOIN users u
				ON ep.user_id = u.id
			WHERE event_id = $1
		`, eventId)
	if err != nil {
		return nil, err
	}
	parts := make([]models.ParticipantUserInfoRow, 0)

	for rows.Next() {
		var part models.ParticipantUserInfoRow

		err := rows.Scan(
			&part.EventId,
			&part.UserId,
			&part.Status,
			&part.Role,
			&part.CreatedAt,
			&part.RespondedAt,
			&part.Username,
		)
		if err != nil {
			return nil, err
		}

		parts = append(parts, part)
	}

	return parts, nil
}

func (s *Store) Rsvp(ctx context.Context, userId uuid.UUID, eventId uuid.UUID, response models.ParticipantStatus) (*models.ParticipantUserInfoRow, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE event_participants ep
		SET status = $1,
		    responded_at = $2
		FROM users u
		WHERE ep.user_id = u.id
		  AND ep.event_id = $3
		  AND ep.user_id = $4
		RETURNING ep.event_id, ep.user_id, ep.status, ep.role, ep.created_at, ep.responded_at, u.username
	`, response, time.Now(), eventId, userId)

	var part models.ParticipantUserInfoRow
	err := row.Scan(
		&part.EventId,
		&part.UserId,
		&part.Status,
		&part.Role,
		&part.CreatedAt,
		&part.RespondedAt,
		&part.Username,
	)
	if err != nil {
		return nil, helpers.ErrRsvpFailed
	}

	return &part, nil
}

func (s *Store) IsUserPartOfEvent(ctx context.Context, userId uuid.UUID, eventId uuid.UUID) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `			
		SELECT EXISTS(
			SELECT 1
			FROM event_participants
			WHERE user_id = $1 AND event_id = $2
		)
		`, userId, eventId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) GetEventUserRole(ctx context.Context, userId uuid.UUID, eventId uuid.UUID) (models.ParticipantRole, error) {
	var role models.ParticipantRole
	row := s.pool.QueryRow(ctx, `			
		SELECT role FROM event_participants
		WHERE user_id = $1 AND event_id = $2
		`, userId, eventId)
	err := row.Scan(&role)
	if err != nil {
		return models.EventParticipant, err
	}
	return role, nil
}

func (s *Store) UpdateEventInfo(ctx context.Context, eventId uuid.UUID, dto models.UpdateEventModel) (*models.EventsRow, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE events
		SET
			name = $1,
			description = $2,
			location = $3,
			status = $4,
			rsvp_deadline = $5,
			starts_at = $6,
			ends_at = $7
		WHERE id = $8
		RETURNING *
	`, dto.Name, dto.Description, dto.Location, dto.Status, dto.RsvpDeadline, dto.StartsAt, dto.EndsAt, eventId)

	var event models.EventsRow
	err := row.Scan(&event.Id, &event.GroupId, &event.CreatedBy, &event.Name, &event.Description, &event.Location, &event.Status, &event.CreatedAt, &event.RsvpDeadline, &event.StartsAt, &event.EndsAt)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *Store) DeleteEvent(ctx context.Context, eventId uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM events
		WHERE id = $1
		`, eventId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helpers.ErrEventNotDeleted
	}
	return nil
}

func (s *Store) RemoveParticipant(ctx context.Context, eventId uuid.UUID, userId uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM event_participants
		WHERE user_id = $1 AND event_id = $2
		`, userId, eventId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helpers.ErrEventParticipantNotDeleted
	}
	return nil
}

func (s *Store) AddParticipants(ctx context.Context, eventId uuid.UUID, participantIds []uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, currentId := range participantIds {
		_, err := tx.Exec(ctx, `
			INSERT INTO event_participants(event_id, user_id, status, role)
			SELECT e.id, gm.user_id, $3, $4
			FROM events e
			JOIN group_members gm
				ON gm.group_id = e.group_id
			WHERE e.id = $1 AND gm.user_id = $2
			ON CONFLICT (event_id, user_id) DO NOTHING
		`, eventId, currentId, models.EventInvited, models.EventParticipant)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateEvent(ctx context.Context, groupId uuid.UUID, userId uuid.UUID, dto models.EventModelDto) (*models.EventsRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var event models.EventsRow

	err = tx.QueryRow(ctx, `
		INSERT INTO events(group_id, created_by, name, description, location, status, rsvp_deadline, starts_at, ends_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, group_id, created_by, name, description, location, status, created_at, rsvp_deadline, starts_at, ends_at
	`, groupId, userId, dto.Name, dto.Description, dto.Location, dto.Status, dto.RsvpDeadline, dto.StartsAt, dto.EndsAt).Scan(
		&event.Id,
		&event.GroupId,
		&event.CreatedBy,
		&event.Name,
		&event.Description,
		&event.Location,
		&event.Status,
		&event.CreatedAt,
		&event.RsvpDeadline,
		&event.StartsAt,
		&event.EndsAt,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO event_participants(event_id, user_id, status, role, responded_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, event.Id, userId, models.EventGoing, models.EventOwner)
	if err != nil {
		return nil, err
	}

	for _, participantId := range dto.ParticipantIds {
		if participantId == userId {
			continue
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO event_participants(event_id, user_id, status, role)
			SELECT e.id, gm.user_id, $3, $4
			FROM events e
			JOIN group_members gm
				ON gm.group_id = e.group_id
			WHERE e.id = $1 AND gm.user_id = $2
			ON CONFLICT (event_id, user_id) DO NOTHING
		`, event.Id, participantId, models.EventInvited, models.EventParticipant)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *Store) GetGroupEvents(ctx context.Context, userId, groupId uuid.UUID) ([]models.EventsRow, error) {
	// only if user is in event
	events := make([]models.EventsRow, 0)

	rows, err := s.pool.Query(ctx, `
			SELECT e.id, e.group_id, e.created_by, e.name, e.description, e.location, e.status, e.created_at, e.rsvp_deadline, e.starts_at, e.ends_at
			FROM events e
			JOIN event_participants ep
				ON ep.event_id = e.id
			WHERE ep.user_id = $1 AND e.group_id = $2
		`, userId, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event models.EventsRow

		err := rows.Scan(&event.Id, &event.GroupId, &event.CreatedBy, &event.Name, &event.Description, &event.Location, &event.Status, &event.CreatedAt, &event.RsvpDeadline, &event.StartsAt, &event.EndsAt)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}
