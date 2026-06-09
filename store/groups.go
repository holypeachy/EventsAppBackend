package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/helpers"
	"github.com/holypeachy/EventsAppBackend/models"
)

func (s *Store) CreateGroup(ctx context.Context, userId uuid.UUID, name, desc string) (*models.GroupsRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	code, err := helpers.GenerateNewInviteCode(helpers.InviteCodeLength)
	if err != nil {
		return nil, err
	}

	var group models.GroupsRow
	row := tx.QueryRow(ctx, `
			INSERT INTO groups(name, description, created_by, invite_code)
			VALUES ($1, $2, $3, $4)
			RETURNING *
		`, name, desc, userId, code)

	err = row.Scan(&group.Id, &group.Name, &group.Description, &group.CreatedBy, &group.CreatedAt, &group.InviteCode)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
			INSERT INTO group_members(group_id, user_id, role)
			VALUES ($1, $2, $3)
		`, group.Id, userId, models.Owner)

	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (s *Store) JoinGroupByInviteCode(ctx context.Context, userId uuid.UUID, inviteCode string) (*models.GroupsRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id, role)
		SELECT id, $1, $2
		FROM groups
		WHERE invite_code = $3
	`, userId, models.Member, inviteCode)
	if err != nil {
		return nil, err
	}

	var group models.GroupsRow
	row := tx.QueryRow(ctx, `
		SELECT * FROM groups
		WHERE invite_code = $1
		`, inviteCode)
	err = row.Scan(&group.Id, &group.Name, &group.Description, &group.CreatedBy, &group.CreatedAt, &group.InviteCode)
	if err != nil {
		return nil, err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (s *Store) GetGroupsUserBelongsTo(ctx context.Context, userId uuid.UUID) (*[]models.GroupsRow, error) {
	rows, err := s.pool.Query(ctx, `
			SELECT g.id, g.name, g.description, g.created_by, g.created_at, g.invite_code
			FROM groups g
			JOIN group_members gm
				ON gm.group_id = g.id
			WHERE gm.user_id = $1
		`, userId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]models.GroupsRow, 0)

	for rows.Next() {
		var group models.GroupsRow

		err := rows.Scan(
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

func (s *Store) GetGroupById(ctx context.Context, groupId uuid.UUID) (*models.GroupsRow, error) {
	var group models.GroupsRow

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
	// return user?
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

func (s *Store) GetGroupMembers(ctx context.Context, groupId uuid.UUID) (*[]models.UsersRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id, u.username, u.email, u.password_hash, u.created_at
		FROM users u
		JOIN group_members gm
			ON gm.user_id = u.id
		WHERE gm.group_id = $1
	`, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.UsersRow, 0)

	for rows.Next() {
		var user models.UsersRow

		err := rows.Scan(
			&user.Id,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &users, nil
}

func (s *Store) GetUserRoleInGroup(ctx context.Context, userId uuid.UUID, groupId uuid.UUID) (models.GroupRole, error) {
	var role models.GroupRole

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

func (s *Store) UpdateGroupInviteCode(ctx context.Context, groupId uuid.UUID, code string) (*models.GroupsRow, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE groups
		SET invite_code = $1
		WHERE id = $2
		RETURNING *
		`, code, groupId)

	var group models.GroupsRow
	err := row.Scan(&group.Id, &group.Name, &group.Description, &group.CreatedBy, &group.CreatedAt, &group.InviteCode)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (s *Store) UpdateGroupInfo(ctx context.Context, groupId uuid.UUID, name, desc string) (*models.GroupsRow, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE groups
		SET name = $1, description = $2
		WHERE id = $3
		RETURNING *
		`, name, desc, groupId)

	var group models.GroupsRow
	err := row.Scan(&group.Id, &group.Name, &group.Description, &group.CreatedBy, &group.CreatedAt, &group.InviteCode)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (s *Store) UpdateMemberRole(ctx context.Context, groupId uuid.UUID, memberId uuid.UUID, role models.GroupRole) (*models.GroupMemberRow, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE group_members
		SET role = $1
		WHERE group_id = $2 AND user_id = $3
		RETURNING *
		`, role, groupId, memberId)
	var member models.GroupMemberRow
	err := row.Scan(&member.GroupId, &member.UserId, &member.Role, &member.JoinedAt)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (s *Store) RemoveMemberFromGroup(ctx context.Context, groupId uuid.UUID, memberId uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM group_members
		WHERE group_id = $1 AND user_id = $2
		`, groupId, memberId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
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
