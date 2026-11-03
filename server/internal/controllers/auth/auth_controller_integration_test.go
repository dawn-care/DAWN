package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	mw "server/internal/middleware"
	"server/internal/migrations"
	authsvc "server/internal/services/auth"
	"server/internal/services/user/caregiver"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Logs in and out through the real controller, middleware and SQL.
// Set MIGRATION_TEST_DSN only for a disposable PostgreSQL database.
func TestLogoutEndsEverySessionIntegration(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires a disposable PostgreSQL database via MIGRATION_TEST_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer admin.Close()

	schema := fmt.Sprintf("logout_test_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	defer admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, migrations.Up(context.Background(), db))

	sessions := authsvc.NewSessionService(db)
	caregivers := caregiver.NewCaregiverService(db)
	service := authsvc.NewAuthService(nil, caregivers, nil, sessions)
	controller := &AuthController{AuthService: service}
	protected := mw.RequireAuth(sessions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	newCaregiver := func() int {
		var id int
		require.NoError(t, db.QueryRow(`INSERT INTO "Caregiver" (name) VALUES ('Logout test') RETURNING id`).Scan(&id))
		return id
	}
	statusWith := func(token string) int {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		w := httptest.NewRecorder()
		protected.ServeHTTP(w, req)
		return w.Code
	}
	ctx := context.Background()

	pat, sam := newCaregiver(), newCaregiver()
	laptop, err := service.StartSession(ctx, pat)
	require.NoError(t, err)
	phone, err := service.StartSession(ctx, pat)
	require.NoError(t, err)
	samLaptop, err := service.StartSession(ctx, sam)
	require.NoError(t, err)
	for _, token := range []string{laptop, phone, samLaptop} {
		require.Equal(t, http.StatusOK, statusWith(token), "logged in before logout")
	}

	w := logout(controller, sessionCookie(laptop))

	assert.Equal(t, http.StatusOK, w.Code)
	assertCleared(t, w)
	assert.Equal(t, http.StatusUnauthorized, statusWith(laptop), "the device that logged out")
	assert.Equal(t, http.StatusUnauthorized, statusWith(phone), "the caregiver's other device")
	assert.Equal(t, http.StatusOK, statusWith(samLaptop), "another caregiver is untouched")

	// Logging out again with the dead cookie is harmless
	w = logout(controller, sessionCookie(laptop))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCleared(t, w)
	assert.Equal(t, http.StatusOK, statusWith(samLaptop))
}
