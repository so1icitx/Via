// Package repository implements PostgreSQL persistence.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides typed database access for auth and saved results.
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

type UserRecord struct {
	ID            string
	Name          string
	Email         string
	EmailVerified bool
	Image         string
}

func (s *Store) CreateUser(ctx context.Context, name, email string, verified bool) (*UserRecord, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO "user" (id, name, email, "emailVerified", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, $4, $5, $5)
	`, id, name, email, verified, now)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return &UserRecord{ID: id, Name: name, Email: email, EmailVerified: verified}, nil
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (*UserRecord, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, email, "emailVerified", COALESCE(image, '')
		FROM "user" WHERE email = $1
	`, email)

	var u UserRecord
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.EmailVerified, &u.Image); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

func (s *Store) FindUserByID(ctx context.Context, id string) (*UserRecord, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, email, "emailVerified", COALESCE(image, '')
		FROM "user" WHERE id = $1
	`, id)

	var u UserRecord
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.EmailVerified, &u.Image); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

func (s *Store) UpdateUserName(ctx context.Context, userID, name string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE "user" SET name = $1, "updatedAt" = $2 WHERE id = $3
	`, name, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("update user name: %w", err)
	}
	return nil
}

func (s *Store) CreateCredentialAccount(ctx context.Context, userID, passwordHash string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO account (id, "accountId", "providerId", "userId", password, "createdAt", "updatedAt")
		VALUES ($1, $2, 'credential', $3, $4, $5, $5)
	`, uuid.NewString(), userID, userID, passwordHash, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert credential account: %w", err)
	}
	return nil
}

func (s *Store) GetCredentialPasswordHash(ctx context.Context, userID string) (string, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT COALESCE(password, '') FROM account
		WHERE "userId" = $1 AND "providerId" = 'credential'
		LIMIT 1
	`, userID)

	var hash string
	if err := row.Scan(&hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("get credential password: %w", err)
	}
	return hash, nil
}

func (s *Store) FindGoogleAccount(ctx context.Context, googleAccountID string) (*UserRecord, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u."emailVerified", COALESCE(u.image, '')
		FROM account a
		JOIN "user" u ON u.id = a."userId"
		WHERE a."providerId" = 'google' AND a."accountId" = $1
		LIMIT 1
	`, googleAccountID)

	var u UserRecord
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.EmailVerified, &u.Image); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find google account: %w", err)
	}
	return &u, nil
}

func (s *Store) CreateGoogleAccount(ctx context.Context, userID, googleAccountID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO account (id, "accountId", "providerId", "userId", "createdAt", "updatedAt")
		VALUES ($1, $2, 'google', $3, $4, $4)
	`, uuid.NewString(), googleAccountID, userID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert google account: %w", err)
	}
	return nil
}

type SessionRecord struct {
	ID        string
	Token     string
	UserID    string
	ExpiresAt time.Time
}

func (s *Store) CreateSession(ctx context.Context, userID, token, ip, userAgent string, expiresAt time.Time) (*SessionRecord, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO session (id, "expiresAt", token, "createdAt", "updatedAt", "ipAddress", "userAgent", "userId")
		VALUES ($1, $2, $3, $4, $4, $5, $6, $7)
	`, id, expiresAt, token, now, ip, userAgent, userID)
	if err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}
	return &SessionRecord{ID: id, Token: token, UserID: userID, ExpiresAt: expiresAt}, nil
}

func (s *Store) FindSessionByToken(ctx context.Context, token string) (*SessionRecord, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, token, "userId", "expiresAt"
		FROM session WHERE token = $1
	`, token)

	var sess SessionRecord
	if err := row.Scan(&sess.ID, &sess.Token, &sess.UserID, &sess.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find session: %w", err)
	}
	return &sess, nil
}

func (s *Store) DeleteSessionByToken(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM session WHERE token = $1`, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *Store) TouchSession(ctx context.Context, sessionID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE session SET "updatedAt" = $1 WHERE id = $2
	`, time.Now().UTC(), sessionID)
	return err
}

func (s *Store) ListSavedResults(ctx context.Context, userID string) ([]model.SavedResultRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, query, data, "createdAt"
		FROM saved_result
		WHERE "userId" = $1
		ORDER BY "createdAt" DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list saved results: %w", err)
	}
	defer rows.Close()

	out := make([]model.SavedResultRow, 0)
	for rows.Next() {
		var row model.SavedResultRow
		var createdAt time.Time
		if err := rows.Scan(&row.ID, &row.Query, &row.Data, &createdAt); err != nil {
			return nil, fmt.Errorf("scan saved result: %w", err)
		}
		row.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *Store) InsertSavedResult(ctx context.Context, userID, query string, data model.ResultsPayload) (int, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("marshal saved result: %w", err)
	}

	var id int
	err = s.pool.QueryRow(ctx, `
		INSERT INTO saved_result ("userId", query, data, "createdAt")
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, query, raw, time.Now().UTC()).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert saved result: %w", err)
	}
	return id, nil
}

func (s *Store) DeleteSavedResult(ctx context.Context, userID string, id int) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM saved_result WHERE id = $1 AND "userId" = $2
	`, id, userID)
	if err != nil {
		return fmt.Errorf("delete saved result: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}