package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/helpers"
	"github.com/holypeachy/EventsAppBackend/models"
)

func (s *Store) RegisterUser(ctx context.Context, username, email, password string) (*models.UsersRow, error) {
	row := s.pool.QueryRow(ctx, `
	INSERT INTO users (username, email, password_hash)
	VALUES ($1,$2,$3)
	RETURNING *`, username, email, password)

	var user models.UsersRow
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.UsersRow, error) {
	row := s.pool.QueryRow(ctx, "SELECT * FROM users WHERE email = $1", email)

	var user models.UsersRow
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) GetUserById(ctx context.Context, id uuid.UUID) (*models.UsersRow, error) {
	row := s.pool.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", id)

	var user models.UsersRow
	err := row.Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) StoreRefreshToken(ctx context.Context, userId uuid.UUID, tokenHash string) error {
	now := time.Now()
	expire := now.Add(720 * time.Hour) //30 days

	_, err := s.pool.Exec(ctx, `
			INSERT INTO refresh_tokens (user_id, token_hash, expires_at, last_used_at, created_at)
			VALUES ($1,$2,$3,$4,$5)
		`, userId, tokenHash, expire, now, now)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetRefreshRowByHash(ctx context.Context, hashedRefreshToken string) (*models.RefreshTokenRow, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT * FROM refresh_tokens
		WHERE token_hash = $1
		`, hashedRefreshToken)

	var token models.RefreshTokenRow
	err := row.Scan(&token.Id, &token.UserId, &token.TokenHash, &token.ExpiresAt, &token.LastUsedAt, &token.CreatedAt)
	if err != nil {
		return nil, err
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
		return helpers.ErrTokenNotFound
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
		return helpers.ErrTokenNotFound
	}

	return nil
}
