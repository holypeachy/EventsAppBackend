package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/helpers"
)

type GroupRow struct {
	Id          uuid.UUID
	Name        string
	Description string
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	InviteCode  string
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

func (s *Store) CreateGroup(ctx context.Context, userId uuid.UUID, name, desc string) (uuid.UUID, error) {
	code, err := helpers.GenerateNewInviteCode(helpers.InviteCodeLength)
	if err != nil {
		return uuid.UUID{}, err
	}

	row := s.pool.QueryRow(ctx, `
			INSERT INTO groups(name, description, created_by, invite_code)
			VALUES ($1,$2,$3,$4)
			RETURNING (id)
		`, name, desc, userId, code)

	var groupId uuid.UUID

	err = row.Scan(&groupId)
	if err != nil {
		return uuid.UUID{}, err
	}

	err = s.assignGroupToUser(ctx, userId, groupId, Owner)
	if err != nil {
		return uuid.UUID{}, err
	}

	return groupId, nil
}

func (s *Store) assignGroupToUser(ctx context.Context, userId uuid.UUID, groupId uuid.UUID, role GroupRole) error {

	_, err := s.pool.Exec(ctx, `
			INSERT INTO group_members(group_id, user_id, role)
			VALUES ($1,$2,$3)
		`, groupId, userId, role)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) JoinGroupByInviteCode(ctx context.Context, userId uuid.UUID, inviteCode string) error {
	groupId, err := s.getGroupIdByInviteCode(ctx, inviteCode)
	if err != nil {
		return err
	}

	err = s.assignGroupToUser(ctx, userId, groupId, Member)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) getGroupIdByInviteCode(ctx context.Context, inviteCode string) (uuid.UUID, error) {
	var groupId uuid.UUID
	row := s.pool.QueryRow(ctx, `
			SELECT id FROM groups
			WHERE invite_code = $1
		`, inviteCode)

	err := row.Scan(&groupId)
	if err != nil {
		return uuid.UUID{}, err
	}

	return groupId, nil
}

func (s *Store) GetGroupsUserBelongsTo(ctx context.Context, userId uuid.UUID) (*[]GroupRow, error) {
	var groupIds []uuid.UUID
	rows, err := s.pool.Query(ctx, `
			SELECT group_id FROM group_members
			WHERE user_id = $1
		`, userId)
	defer rows.Close()

	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id uuid.UUID

		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		groupIds = append(groupIds, id)
	}

	rows2, err := s.pool.Query(ctx, `
		SELECT *
		FROM groups
		WHERE id = ANY($1)
	`, groupIds)
	defer rows2.Close()

	var groups []GroupRow

	for rows2.Next() {
		var group GroupRow

		err := rows2.Scan(
			&group.Id,
			&group.Name,
			&group.Description,
			&group.CreatedBy,
			&group.CreatedAt,
			&group.InviteCode,
		)
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}
	return &groups, nil
}

func (s *Store) GetGroupById(ctx context.Context, groupId uuid.UUID) (*GroupRow, error) {
	var group GroupRow

	row := s.pool.QueryRow(ctx, `
			SELECT * FROM groups
			WHERE id = $1
		`, groupId)

	err := row.Scan(&group.Id, &group.Name, &group.Description, &group.CreatedBy, &group.CreatedAt, &group.InviteCode)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (s *Store) DoesUserBelongToGroup(ctx context.Context, userId uuid.UUID, groupId uuid.UUID) (bool, error) {
	var exists bool

	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM group_members
			WHERE user_id = $1 AND group_id = $2
		)`, userId, groupId).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Store) DoesUserExist(ctx context.Context, userId uuid.UUID) (bool, error) {
	var exists bool

	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE id = $1
		)`, userId).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Store) GetGroupMembers(ctx context.Context, groupId uuid.UUID) (*[]UserRow, error) {

	rows, err := s.pool.Query(ctx, `
		SELECT user_id FROM group_members
		WHERE group_id = $1
	`, groupId)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var userIds []uuid.UUID

	for rows.Next() {
		var id uuid.UUID

		err := rows.Scan(
			&id,
		)
		if err != nil {
			return nil, err
		}

		userIds = append(userIds, id)
	}

	var users []UserRow

	rows2, err := s.pool.Query(ctx, `
		SELECT * FROM users
		WHERE id = ANY($1)
	`, userIds)
	defer rows2.Close()

	if err != nil {
		return nil, err
	}

	for rows2.Next() {
		var user UserRow

		err := rows2.Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return &users, nil
}

func (s *Store) GetUserRoleInGroup(ctx context.Context, userId uuid.UUID, groupId uuid.UUID) (GroupRole, error) {
	var role GroupRole

	row := s.pool.QueryRow(ctx, `
		SELECT role FROM group_members
		WHERE group_id = $1 AND user_id = $2
		`, groupId, userId)

	err := row.Scan(&role)
	if err != nil {
		return "", err
	}

	return role, nil
}

func (s *Store) UpdateGroupInviteCode(ctx context.Context, groupId uuid.UUID, code string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE groups
		SET invite_code = $1
		WHERE id = $2
		`, code, groupId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helpers.ErrNoGroupUpdated
	}
	return nil
}

func (s *Store) UpdateGroupInfo(ctx context.Context, groupId uuid.UUID, name, desc string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE groups
		SET name = $1, description = $2
		WHERE id = $3
		`, name, desc, groupId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() < 1 {
		return helpers.ErrNoGroupUpdated
	}
	return nil
}

func (s *Store) UpdateMemberRole(ctx context.Context, groupId uuid.UUID, memberId uuid.UUID, role GroupRole) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE group_members
		SET role = $1
		WHERE group_id = $2 AND user_id = $3
		`, role, groupId, memberId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() < 1 {
		return helpers.ErrNoGroupUpdated
	}
	return nil
}

func (s *Store) RemoveMemberFromGroup(ctx context.Context, groupId uuid.UUID, memberId uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM group_members
		WHERE group_id = $1 AND user_id = $2
		`, groupId, memberId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return helpers.ErrGroupMemberNotRemoved
	}
	return nil
}

func (s *Store) DeleteGroup(ctx context.Context, groupId uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM groups
		WHERE id = $1
		`, groupId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helpers.ErrGroupNotDeleted
	}
	return nil
}
