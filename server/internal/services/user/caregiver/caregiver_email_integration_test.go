package caregiver

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"server/internal/migrations"
	"server/internal/services/user"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Set MIGRATION_TEST_DSN only for a disposable PostgreSQL database.
// This runs the real SQL of email and Google linking, which the sqlmock tests cannot check.
func TestEmailAndGoogleLinkingIntegration(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires a disposable PostgreSQL database via MIGRATION_TEST_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer admin.Close()

	newService := func(t *testing.T) (*CaregiverService, *sql.DB) {
		schema := fmt.Sprintf("linking_test_%d", time.Now().UnixNano())
		_, err := admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
		require.NoError(t, err)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE") })
		db, err := sql.Open("postgres", dsn+" search_path="+schema)
		require.NoError(t, err)
		t.Cleanup(func() { db.Close() })
		require.NoError(t, migrations.Up(context.Background(), db))
		return NewCaregiverService(db), db
	}
	emailOf := func(t *testing.T, db *sql.DB, id int) (email, googleID sql.NullString) {
		require.NoError(t, db.QueryRow(`SELECT email, google_id FROM "Caregiver" WHERE id = $1`, id).Scan(&email, &googleID))
		return
	}

	t.Run("a Google caregiver gets their email on the next sign-in and then signs in by email as the same account", func(t *testing.T) {
		svc, db := newService(t)
		// Signed up before emails were stored
		var id int
		require.NoError(t, db.QueryRow(`INSERT INTO "Caregiver" (name, role, google_id) VALUES ('Pat', 'caregiver', 'g-1') RETURNING id`).Scan(&id))

		c, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-1", Name: "Pat", Email: "Pat@Example.com", EmailVerified: true})
		require.NoError(t, err)
		assert.Equal(t, id, c.ID)
		email, _ := emailOf(t, db, id)
		assert.Equal(t, "pat@example.com", email.String)

		byEmail, err := svc.FindOrCreateFromEmail("PAT@example.com")
		require.NoError(t, err)
		assert.Equal(t, id, byEmail.ID)
	})

	t.Run("an email caregiver is linked, not duplicated, when they later use Google", func(t *testing.T) {
		svc, db := newService(t)
		first, err := svc.FindOrCreateFromEmail("jo@example.com")
		require.NoError(t, err)
		assert.Equal(t, "jo", first.Name)

		c, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-2", Name: "Jo Smith", Email: "jo@example.com", EmailVerified: true})
		require.NoError(t, err)
		assert.Equal(t, first.ID, c.ID)
		_, googleID := emailOf(t, db, first.ID)
		assert.Equal(t, "g-2", googleID.String)

		again, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-2", Name: "Jo Smith", Email: "jo@example.com", EmailVerified: true})
		require.NoError(t, err)
		assert.Equal(t, first.ID, again.ID, "the next Google sign-in finds the linked account by Google ID")

		var count int
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM "Caregiver" WHERE name <> 'Jane Smith'`).Scan(&count))
		assert.Equal(t, 1, count)
	})

	t.Run("a different Google account with a linked caregiver's email does not take the account over", func(t *testing.T) {
		svc, db := newService(t)
		owner, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-old", Name: "Old Pat", Email: "pat@example.com", EmailVerified: true})
		require.NoError(t, err)

		newcomer, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-new", Name: "New Pat", Email: "pat@example.com", EmailVerified: true})
		require.NoError(t, err)
		assert.NotEqual(t, owner.ID, newcomer.ID)
		email, googleID := emailOf(t, db, newcomer.ID)
		assert.False(t, email.Valid, "the address still belongs to the first account")
		assert.Equal(t, "g-new", googleID.String)
		_, ownerGoogleID := emailOf(t, db, owner.ID)
		assert.Equal(t, "g-old", ownerGoogleID.String, "the first account keeps its Google ID")
	})

	t.Run("an unverified Google email is never matched to an account", func(t *testing.T) {
		svc, _ := newService(t)
		victim, err := svc.FindOrCreateFromEmail("victim@example.com")
		require.NoError(t, err)

		attacker, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-evil", Name: "Evil", Email: "victim@example.com", EmailVerified: false})
		require.NoError(t, err)
		assert.NotEqual(t, victim.ID, attacker.ID)
	})

	t.Run("saving an email never takes one another caregiver has, in any letter case", func(t *testing.T) {
		svc, db := newService(t)
		owner, err := svc.FindOrCreateFromEmail("shared@example.com")
		require.NoError(t, err)
		var other int
		require.NoError(t, db.QueryRow(`INSERT INTO "Caregiver" (name, role, google_id) VALUES ('Other', 'caregiver', 'g-3') RETURNING id`).Scan(&other))

		c, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-3", Name: "Other", Email: "SHARED@example.com", EmailVerified: true})
		require.NoError(t, err, "signing in still works")
		assert.Equal(t, other, c.ID, "linked by Google ID, not merged into the email account")
		email, _ := emailOf(t, db, other)
		assert.False(t, email.Valid)
		ownerEmail, _ := emailOf(t, db, owner.ID)
		assert.Equal(t, "shared@example.com", ownerEmail.String)
	})
}
