package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/helpers"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRow struct {
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

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) RegisterUser(ctx context.Context, username, email, password string) (*UserRow, error) {
	row := s.pool.QueryRow(ctx, `
	INSERT INTO users (username, email, password_hash)
	VALUES ($1,$2,$3)
	RETURNING *`, username, email, password)

	var user UserRow
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*UserRow, error) {
	row := s.pool.QueryRow(ctx, "SELECT * FROM users WHERE email = $1", email)

	var user UserRow
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) GetUserById(ctx context.Context, id uuid.UUID) (*UserRow, error) {
	row := s.pool.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", id)

	var user UserRow
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) StoreRefreshToken(ctx context.Context, userId uuid.UUID, tokenHash string) error {
	now := time.Now()
	expire := now.Add(720 * time.Hour) //30 days

	tag, err := s.pool.Exec(ctx, `
			INSERT INTO refresh_tokens (user_id, token_hash, expires_at, last_used_at, created_at)
			VALUES ($1,$2,$3,$4,$5)
		`, userId, tokenHash, expire, now, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("insert failed")
	}

	return nil
}

func (s *Store) GetRefreshRowByHash(ctx context.Context, hashedRefreshToken string) (*RefreshTokenRow, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT * FROM refresh_tokens
		WHERE token_hash = $1
		`, hashedRefreshToken)

	var token RefreshTokenRow
	err := row.Scan(&token.Id, &token.UserId, &token.TokenHash, &token.ExpiresAt, &token.LastUsedAt, &token.CreatedAt)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	return &token, nil
}

func (s *Store) DeleteRefreshTokenById(ctx context.Context, tokenId uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM refresh_tokens
		WHERE id = $1
		`, tokenId)

	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("token not found")
	}

	return nil
}

func (s *Store) DeleteRefreshTokenByHash(ctx context.Context, hashedToken string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM refresh_tokens
		WHERE token_hash = $1
		`, hashedToken)

	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("token not found")
	}

	return nil
}

func (s *Store) CreateGroup(ctx context.Context, userId uuid.UUID, name, desc string) (uuid.UUID, error) {
	code, err := helpers.GenerateNewInviteCode(helpers.InviteCodeLength)
	if err != nil {
		log.Println("error: ", err)
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
		log.Println("error: ", err)
		return uuid.UUID{}, err
	}

	return groupId, nil
}

func (s *Store) assignGroupToUser(ctx context.Context, userId uuid.UUID, groupId uuid.UUID, role GroupRole) error {

	tags, err := s.pool.Exec(ctx, `
			INSERT INTO group_members(group_id, user_id, role)
			VALUES ($1,$2,$3)
		`, groupId, userId, role)

	if err != nil {
		return err
	}

	if tags.RowsAffected() == 0 {
		return errors.New("failed to insert")
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
	log.Println(groupId)

	return groupId, nil
}

func (s *Store) GetGroupsUserBelongsTo(ctx context.Context, userId uuid.UUID) (*[]GroupRow, error) {
	var groupIds []uuid.UUID
	rows, err := s.pool.Query(ctx, `
			SELECT group_id FROM group_members
			WHERE user_id = $1
		`, userId)

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
	rows.Close()

	rows, err = s.pool.Query(ctx, `
		SELECT *
		FROM groups
		WHERE id = ANY($1)
	`, groupIds)

	var groups []GroupRow

	for rows.Next() {
		var group GroupRow

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

	rows, err = s.pool.Query(ctx, `
		SELECT * FROM users
		WHERE id = ANY($1)
	`, userIds)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var user UserRow

		err := rows.Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash)
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
		log.Println("error:", err)
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
	if tag.RowsAffected() != 1 {
		return errors.New("no group updated")
	}
	return nil
}
