package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"
)

type SessionService interface {
	Create(ctx context.Context, userID int) (string, error)
	Validate(ctx context.Context, sessionID string) (int, error)
}

type sessionService struct {
	db *sql.DB
}

func NewSessionService(db *sql.DB) *sessionService {
	return &sessionService{db: db}
}

const sessionTTL = time.Hour * 24 * 7 // token lasts a week (can update however we'd prefer)

// Generate a random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Create a session record in DB
func (s *sessionService) Create(ctx context.Context, userID int) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	expiry := time.Now().Add(sessionTTL)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO "Session" (session_id, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, sessionID, userID, expiry)

	if err != nil {
		return "", err
	}

	return sessionID, nil
}

// Validate session and return userID
func (s *sessionService) Validate(ctx context.Context, sessionID string) (int, error) {
	var userID int
	var expiresAt time.Time

	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, expires_at
		FROM "Session"
		WHERE session_id = $1
	`, sessionID).Scan(&userID, &expiresAt)

	if err != nil {
		return 0, err
	}

	if time.Now().After(expiresAt) {
		return 0, errors.New("session expired")
	}

	return userID, nil
}
