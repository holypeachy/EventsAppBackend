package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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
