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

		c, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-1", Name: "Pat", Email: "Pat@Gmail.com", EmailVerified: true})
		require.NoError(t, err)
		assert.Equal(t, id, c.ID)
		email, _ := emailOf(t, db, id)
		assert.Equal(t, "pat@gmail.com", email.String)

		byEmail, err := svc.FindOrCreateFromEmail("PAT@gmail.com")
		require.NoError(t, err)
		assert.Equal(t, id, byEmail.ID)
	})

	t.Run("an email caregiver is linked, not duplicated, when they later use Google", func(t *testing.T) {
		svc, db := newService(t)
		first, err := svc.FindOrCreateFromEmail("jo@gmail.com")
		require.NoError(t, err)
		assert.Equal(t, "jo", first.Name)

		c, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-2", Name: "Jo Smith", Email: "jo@gmail.com", EmailVerified: true})
		require.NoError(t, err)
		assert.Equal(t, first.ID, c.ID)
		_, googleID := emailOf(t, db, first.ID)
		assert.Equal(t, "g-2", googleID.String)

		again, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-2", Name: "Jo Smith", Email: "jo@gmail.com", EmailVerified: true})
		require.NoError(t, err)
		assert.Equal(t, first.ID, again.ID, "the next Google sign-in finds the linked account by Google ID")

		var count int
		// Not counting Jane Smith, the demo caregiver migration 1 seeds
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM "Caregiver" WHERE name <> 'Jane Smith'`).Scan(&count))
		assert.Equal(t, 1, count)
	})

	t.Run("a different Google account with a linked caregiver's email does not take the account over", func(t *testing.T) {
		svc, db := newService(t)
		owner, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-old", Name: "Old Pat", Email: "pat@gmail.com", EmailVerified: true})
		require.NoError(t, err)

		newcomer, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-new", Name: "New Pat", Email: "pat@gmail.com", EmailVerified: true})
		require.NoError(t, err)
		assert.NotEqual(t, owner.ID, newcomer.ID)
		email, googleID := emailOf(t, db, newcomer.ID)
		assert.False(t, email.Valid, "the address still belongs to the first account")
		assert.Equal(t, "g-new", googleID.String)
		_, ownerGoogleID := emailOf(t, db, owner.ID)
		assert.Equal(t, "g-old", ownerGoogleID.String, "the first account keeps its Google ID")
	})

	t.Run("an email-only account made before the Google account had its email is merged into it", func(t *testing.T) {
		svc, db := newService(t)
		// Signed up with Google before emails were stored, then used email sign-in
		var googleAccount int
		require.NoError(t, db.QueryRow(`INSERT INTO "Caregiver" (name, role, google_id) VALUES ('Pat', 'caregiver', 'g-3') RETURNING id`).Scan(&googleAccount))
		emailAccount, err := svc.FindOrCreateFromEmail("shared@gmail.com")
		require.NoError(t, err)
		var carepartner int
		require.NoError(t, db.QueryRow(`INSERT INTO "Carepartner" (name, caregiver_id) VALUES ('Mom', $1) RETURNING id`, emailAccount.ID).Scan(&carepartner))
		_, err = db.Exec(`INSERT INTO "Session" (token_hash, user_id, expires_at) VALUES ('h', $1, now() + interval '1 day')`, emailAccount.ID)
		require.NoError(t, err)

		c, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-3", Name: "Pat", Email: "SHARED@gmail.com", EmailVerified: true})
		require.NoError(t, err)
		assert.Equal(t, googleAccount, c.ID, "the Google account is kept")
		email, _ := emailOf(t, db, googleAccount)
		assert.Equal(t, "shared@gmail.com", email.String, "and gets the email")

		var owner int
		require.NoError(t, db.QueryRow(`SELECT caregiver_id FROM "Carepartner" WHERE id = $1`, carepartner).Scan(&owner))
		assert.Equal(t, googleAccount, owner, "the carepartner moved")
		var left int
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM "Caregiver" WHERE id = $1`, emailAccount.ID).Scan(&left))
		assert.Zero(t, left, "the email account is gone")
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM "Session" WHERE user_id = $1`, emailAccount.ID).Scan(&left))
		assert.Zero(t, left, "with its sessions")

		byEmail, err := svc.FindOrCreateFromEmail("shared@gmail.com")
		require.NoError(t, err)
		assert.Equal(t, googleAccount, byEmail.ID, "email sign-in now reaches the Google account")
	})

	t.Run("an email held by an account with another Google ID is not saved or merged", func(t *testing.T) {
		svc, db := newService(t)
		owner, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-owner", Name: "Owner", Email: "shared@gmail.com", EmailVerified: true})
		require.NoError(t, err)
		var other int
		require.NoError(t, db.QueryRow(`INSERT INTO "Caregiver" (name, role, google_id) VALUES ('Other', 'caregiver', 'g-3') RETURNING id`).Scan(&other))

		c, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-3", Name: "Other", Email: "SHARED@gmail.com", EmailVerified: true})
		require.NoError(t, err, "signing in still works")
		assert.Equal(t, other, c.ID)
		email, _ := emailOf(t, db, other)
		assert.False(t, email.Valid)
		ownerEmail, _ := emailOf(t, db, owner.ID)
		assert.Equal(t, "shared@gmail.com", ownerEmail.String)
	})

	t.Run("a personal Google account with an outside address does not take over the email caregiver who has it now", func(t *testing.T) {
		svc, db := newService(t)
		carol, err := svc.FindOrCreateFromEmail("bob@corp.org")
		require.NoError(t, err)

		bob, err := svc.FindOrCreateFromGoogle(&user.GoogleUser{ID: "g-bob", Name: "Bob", Email: "bob@corp.org", EmailVerified: true})
		require.NoError(t, err)
		assert.NotEqual(t, carol.ID, bob.ID, "Bob gets his own account")
		_, carolGoogleID := emailOf(t, db, carol.ID)
		assert.False(t, carolGoogleID.Valid, "Carol's account is not linked to Bob's Google account")
		bobEmail, _ := emailOf(t, db, bob.ID)
		assert.False(t, bobEmail.Valid, "Bob's account does not get the address")
	})
}
