package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"time"
)

const (
	// A session ends this long after it last renewed. A request through RequireAuth renews it,
	// at most once every sessionTouchInterval.
	SessionIdleTTL = 90 * 24 * time.Hour
	// A session ends this long after it was created, however active it was
	SessionAbsoluteTTL = 365 * 24 * time.Hour
	// A session renews at most this often, so using the app is not a database write each time
	sessionTouchInterval = 24 * time.Hour
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

// ValidatedSession is a session that is still valid
type ValidatedSession struct {
	UserID    int
	ExpiresAt time.Time
	// Refreshed is true when this call moved ExpiresAt later, so the cookie needs sending again
	Refreshed bool
}

type SessionService interface {
	Create(ctx context.Context, userID int) (string, error)
	Validate(ctx context.Context, token string) (ValidatedSession, error)
	Find(ctx context.Context, token string) (ValidatedSession, error)
	DeleteAllForUser(ctx context.Context, userID int) error
}

type sessionService struct {
	db  *sql.DB
	now func() time.Time
}

func NewSessionService(db *sql.DB) *sessionService {
	return &sessionService{db: db, now: time.Now}
}

// Generate a random session token, which only the browser's cookie holds
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// The database stores only this hash, so a leaked database or backup holds no usable logins
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Create a session record in DB and return the token for the cookie
func (s *sessionService) Create(ctx context.Context, userID int) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	now := s.now()

	// Clean up this caregiver's dead sessions; a failure here must not block the login
	_, err = s.db.ExecContext(ctx, `
		DELETE FROM "Session"
		WHERE user_id = $1 AND (expires_at < $2 OR created_at < $3)
	`, userID, now, now.Add(-SessionAbsoluteTTL))
	if err != nil {
		log.Printf("session cleanup for caregiver %d failed: %v", userID, err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO "Session" (token_hash, user_id, created_at, expires_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $3)
	`, hashToken(token), userID, now, now.Add(SessionIdleTTL))
	if err != nil {
		return "", err
	}

	return token, nil
}

// storedSession is a session row that has not ended
type storedSession struct {
	userID                           int
	createdAt, expiresAt, lastSeenAt time.Time
}

// cutoff is when the session ends however active it is
func (r storedSession) cutoff() time.Time {
	return r.createdAt.Add(SessionAbsoluteTTL)
}

// find looks up a session by its token's hash. One that has ended is deleted and reported as expired.
func (s *sessionService) find(ctx context.Context, tokenHash string, now time.Time) (storedSession, error) {
	var row storedSession
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, created_at, expires_at, last_seen_at
		FROM "Session"
		WHERE token_hash = $1
	`, tokenHash).Scan(&row.userID, &row.createdAt, &row.expiresAt, &row.lastSeenAt)
	if errors.Is(err, sql.ErrNoRows) {
		return storedSession{}, ErrSessionNotFound
	}
	if err != nil {
		return storedSession{}, err
	}

	if now.After(row.expiresAt) || now.After(row.cutoff()) {
		if _, err := s.db.ExecContext(ctx, `DELETE FROM "Session" WHERE token_hash = $1`, tokenHash); err != nil {
			log.Printf("deleting expired session of caregiver %d failed: %v", row.userID, err)
		}
		return storedSession{}, ErrSessionExpired
	}
	return row, nil
}

// Find returns who a session belongs to without renewing it, for logout, which is about to delete it
func (s *sessionService) Find(ctx context.Context, token string) (ValidatedSession, error) {
	row, err := s.find(ctx, hashToken(token), s.now())
	if err != nil {
		return ValidatedSession{}, err
	}
	return ValidatedSession{UserID: row.userID, ExpiresAt: row.expiresAt}, nil
}

// Validate session and return who it belongs to, renewing it when it was last renewed over a day ago
func (s *sessionService) Validate(ctx context.Context, token string) (ValidatedSession, error) {
	tokenHash := hashToken(token)
	now := s.now()

	row, err := s.find(ctx, tokenHash, now)
	if err != nil {
		return ValidatedSession{}, err
	}

	if now.Sub(row.lastSeenAt) < sessionTouchInterval {
		return ValidatedSession{UserID: row.userID, ExpiresAt: row.expiresAt}, nil
	}

	// Renew: 90 days from now, but never past the cutoff
	renewed := now.Add(SessionIdleTTL)
	if renewed.After(row.cutoff()) {
		renewed = row.cutoff()
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE "Session" SET expires_at = $2, last_seen_at = $3
		WHERE token_hash = $1
	`, tokenHash, renewed, now)
	if err != nil {
		// The session is still valid, so a failed renewal must not log the caregiver out
		log.Printf("renewing session of caregiver %d failed: %v", row.userID, err)
		return ValidatedSession{UserID: row.userID, ExpiresAt: row.expiresAt}, nil
	}
	if n, err := result.RowsAffected(); err == nil && n == 0 {
		// It ended between the lookup and the renewal (a logout on another device), so it no
		// longer exists and must not get a new cookie
		return ValidatedSession{}, ErrSessionNotFound
	}

	return ValidatedSession{UserID: row.userID, ExpiresAt: renewed, Refreshed: renewed.After(row.expiresAt)}, nil
}

// DeleteAllForUser ends every session of one caregiver, on all their devices (logout). Deleting
// the caregiver ends them through the foreign key's ON DELETE CASCADE instead.
func (s *sessionService) DeleteAllForUser(ctx context.Context, userID int) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM "Session" WHERE user_id = $1`, userID)
	return err
}
