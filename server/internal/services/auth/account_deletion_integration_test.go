package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"server/internal/services/user/caregiver"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"server/internal/migrations"
)

// Deletes accounts through the real caregiver service, transaction and schema, and checks that
// every login of that caregiver stops working. The sessions are removed by the Session table's
// ON DELETE CASCADE (migration 8), so this is also what fails if that foreign key goes missing.
// Set MIGRATION_TEST_DSN only for a disposable PostgreSQL database.
func TestAccountDeletionEndsSessionsIntegration(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires a disposable PostgreSQL database via MIGRATION_TEST_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer admin.Close()

	schema := fmt.Sprintf("account_deletion_test_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	defer admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, migrations.Up(context.Background(), db))

	ctx := context.Background()
	sessions := NewSessionService(db)
	caregivers := caregiver.NewCaregiverService(db)

	newCaregiver := func(t *testing.T) int {
		t.Helper()
		var id int
		require.NoError(t, db.QueryRow(`INSERT INTO "Caregiver" (name) VALUES ('Deletion test') RETURNING id`).Scan(&id))
		return id
	}
	// A carepartner paired to the caregiver, with a pairing code: what a real account has
	pairCarepartner := func(t *testing.T, caregiverID int, code string) int {
		t.Helper()
		var id int
		require.NoError(t, db.QueryRow(`INSERT INTO "Carepartner" (name, caregiver_id) VALUES ('Test carepartner', $1) RETURNING id`, caregiverID).Scan(&id))
		_, err := db.Exec(`INSERT INTO "DevicePairing" (carepartner_id, pairing_code, expires_at) VALUES ($1, $2, now() + interval '1 hour')`, id, code)
		require.NoError(t, err)
		return id
	}
	login := func(t *testing.T, caregiverID int) string {
		t.Helper()
		token, err := sessions.Create(ctx, caregiverID)
		require.NoError(t, err)
		return token
	}
	loggedIn := func(token string) bool {
		_, err := sessions.Validate(ctx, token)
		return err == nil
	}
	count := func(t *testing.T, query string, args ...any) (n int) {
		t.Helper()
		require.NoError(t, db.QueryRow(query, args...).Scan(&n))
		return n
	}

	t.Run("deleting the account ends every session of that caregiver", func(t *testing.T) {
		pat, sam := newCaregiver(t), newCaregiver(t)
		carepartner := pairCarepartner(t, pat, "PAT111")
		laptop, phone, samLaptop := login(t, pat), login(t, pat), login(t, sam)
		require.True(t, loggedIn(laptop) && loggedIn(phone) && loggedIn(samLaptop))

		require.NoError(t, caregivers.DeleteCaregiver(pat))

		assert.False(t, loggedIn(laptop))
		assert.False(t, loggedIn(phone))
		assert.True(t, loggedIn(samLaptop), "another caregiver is untouched")
		assert.Equal(t, 0, count(t, `SELECT count(*) FROM "Session" WHERE user_id = $1`, pat))
		assert.Equal(t, 0, count(t, `SELECT count(*) FROM "Caregiver" WHERE id = $1`, pat))
		assert.Equal(t, 1, count(t, `SELECT count(*) FROM "Carepartner" WHERE id = $1 AND caregiver_id IS NULL`, carepartner),
			"the carepartner is unpaired, not deleted")
		assert.Equal(t, 0, count(t, `SELECT count(*) FROM "DevicePairing" WHERE carepartner_id = $1`, carepartner))
	})

	t.Run("an account with no carepartner can be deleted too", func(t *testing.T) {
		id := newCaregiver(t)
		token := login(t, id)

		require.NoError(t, caregivers.DeleteCaregiver(id))

		assert.False(t, loggedIn(token))
	})

	t.Run("deleting an account that does not exist changes nobody's login", func(t *testing.T) {
		other := newCaregiver(t)
		token := login(t, other)

		assert.Error(t, caregivers.DeleteCaregiver(-5))
		assert.Error(t, caregivers.DeleteCaregiver(987654))

		assert.True(t, loggedIn(token))
	})

	t.Run("a deletion that fails changes nothing, so the caregiver is not left half logged out", func(t *testing.T) {
		pat := newCaregiver(t)
		carepartner := pairCarepartner(t, pat, "PAT222")
		token := login(t, pat)
		// Make the last step of the deletion, removing the caregiver, fail
		_, err := db.Exec(`CREATE FUNCTION block_caregiver_delete() RETURNS trigger LANGUAGE plpgsql
			AS $$ BEGIN RAISE EXCEPTION 'blocked for this test'; END $$`)
		require.NoError(t, err)
		_, err = db.Exec(`CREATE TRIGGER block_caregiver_delete BEFORE DELETE ON "Caregiver"
			FOR EACH ROW EXECUTE FUNCTION block_caregiver_delete()`)
		require.NoError(t, err)
		// Defers run last first, so the trigger is dropped before the function it uses
		defer db.Exec(`DROP FUNCTION block_caregiver_delete()`)
		defer db.Exec(`DROP TRIGGER block_caregiver_delete ON "Caregiver"`)

		assert.Error(t, caregivers.DeleteCaregiver(pat))

		assert.True(t, loggedIn(token), "the session survives a failed deletion")
		assert.Equal(t, 1, count(t, `SELECT count(*) FROM "Caregiver" WHERE id = $1`, pat))
		assert.Equal(t, 1, count(t, `SELECT count(*) FROM "Carepartner" WHERE id = $1 AND caregiver_id = $2`, carepartner, pat),
			"still paired")
		assert.Equal(t, 1, count(t, `SELECT count(*) FROM "DevicePairing" WHERE carepartner_id = $1`, carepartner),
			"pairing code kept")
	})
}
