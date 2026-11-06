package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"server/internal/migrations"
	"server/internal/services/user/caregiver"
)

// clock is a time the test moves by hand
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

type codeEnv struct {
	svc    *loginCodeService
	db     *sql.DB
	mailer *codeMailer
	clock  *clock
	issued []string // codes handed out, oldest first
}

// newCodeEnv gives each test its own schema on a disposable PostgreSQL database, and codes it can predict
func newCodeEnv(t *testing.T) *codeEnv {
	t.Helper()
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires a disposable PostgreSQL database via MIGRATION_TEST_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { admin.Close() })

	schema := fmt.Sprintf("login_code_test_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() { admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE") })
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, migrations.Up(context.Background(), db))

	env := &codeEnv{db: db, mailer: &codeMailer{}, clock: &clock{t: time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)}}
	env.svc, err = NewLoginCodeService(db, env.mailer, []byte(testCodeSecret))
	require.NoError(t, err)
	env.svc.now = env.clock.Now
	var mu sync.Mutex
	next := 0
	env.svc.generate = func() (string, error) {
		mu.Lock()
		defer mu.Unlock()
		next++
		code := fmt.Sprintf("%06d", 100000+next)
		env.issued = append(env.issued, code)
		return code, nil
	}
	return env
}

func (e *codeEnv) count(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	require.NoError(t, e.db.QueryRow(query, args...).Scan(&n))
	return n
}

// lastCode is the most recent code the service made
func (e *codeEnv) lastCode() string { return e.issued[len(e.issued)-1] }

// Set MIGRATION_TEST_DSN only for a disposable PostgreSQL database.
func TestLoginCodeFlowIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("a right code works once", func(t *testing.T) {
		env := newCodeEnv(t)
		require.NoError(t, env.svc.RequestCode(ctx, "Pat@Example.com", testIP))
		require.Len(t, env.mailer.emails(), 1)
		assert.Equal(t, testAddress, env.mailer.emails()[0].to)

		assert.ErrorIs(t, env.svc.VerifyCode(ctx, testAddress, "000000"), ErrCodeRejected)
		assert.NoError(t, env.svc.VerifyCode(ctx, " PAT@example.com ", env.lastCode()))
		assert.ErrorIs(t, env.svc.VerifyCode(ctx, testAddress, env.lastCode()), ErrCodeRejected, "a used code does not work again")
	})

	t.Run("a code expires after 10 minutes", func(t *testing.T) {
		env := newCodeEnv(t)
		require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP))
		env.clock.Advance(10 * time.Minute)
		assert.ErrorIs(t, env.svc.VerifyCode(ctx, testAddress, env.lastCode()), ErrCodeRejected)
	})

	t.Run("a new code replaces the old one", func(t *testing.T) {
		env := newCodeEnv(t)
		require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP))
		first := env.lastCode()
		require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP))
		second := env.lastCode()
		require.NotEqual(t, first, second)

		assert.ErrorIs(t, env.svc.VerifyCode(ctx, testAddress, first), ErrCodeRejected, "the old code no longer works")
		assert.NoError(t, env.svc.VerifyCode(ctx, testAddress, second))
	})

	t.Run("guessing wrong 5 times burns the code, even for the right code afterwards", func(t *testing.T) {
		env := newCodeEnv(t)
		require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP))
		for i := 1; i <= 4; i++ {
			assert.ErrorIs(t, env.svc.VerifyCode(ctx, testAddress, "000000"), ErrCodeRejected, "wrong guess %d", i)
		}
		assert.ErrorIs(t, env.svc.VerifyCode(ctx, testAddress, "000000"), ErrTooManyAttempts, "wrong guess 5")
		assert.ErrorIs(t, env.svc.VerifyCode(ctx, testAddress, env.lastCode()), ErrTooManyAttempts, "the right code is refused once burned")
		assert.Equal(t, loginCodeMaxAttempts, env.count(t, `SELECT attempts FROM "LoginCode"`), "refused guesses are not counted further")

		require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP), "a new code starts fresh")
		assert.NoError(t, env.svc.VerifyCode(ctx, testAddress, env.lastCode()))
	})

	t.Run("a code never stored in plain text, and neither is the IP", func(t *testing.T) {
		env := newCodeEnv(t)
		require.NoError(t, env.svc.RequestCode(ctx, "Pat@Example.com", testIP))

		var email, hash string
		require.NoError(t, env.db.QueryRow(`SELECT email, code_hash FROM "LoginCode"`).Scan(&email, &hash))
		assert.Equal(t, testAddress, email, "stored lowercase")
		assert.NotContains(t, hash, env.lastCode())
		assert.Regexp(t, `^[0-9a-f]{64}$`, hash)

		var requestEmail, ipHash string
		require.NoError(t, env.db.QueryRow(`SELECT email, ip_hash FROM "LoginCodeRequest"`).Scan(&requestEmail, &ipHash))
		assert.Equal(t, testAddress, requestEmail)
		assert.NotContains(t, ipHash, testIP)
		assert.Regexp(t, `^[0-9a-f]{64}$`, ipHash)
	})
}

func TestLoginCodeRateLimitsIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("3 codes per email in 15 minutes, then again once the window has passed", func(t *testing.T) {
		env := newCodeEnv(t)
		for i := 0; i < 3; i++ {
			require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP), "request %d", i+1)
		}
		assert.ErrorIs(t, env.svc.RequestCode(ctx, "PAT@example.com", testIP), ErrTooManyRequests, "the 4th, in any letter case")
		assert.Len(t, env.mailer.emails(), 3, "no 4th email")

		env.clock.Advance(15*time.Minute + time.Second)
		assert.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP))
		assert.Len(t, env.mailer.emails(), 4)
	})

	t.Run("10 codes per email in a day", func(t *testing.T) {
		env := newCodeEnv(t)
		for i := 0; i < 10; i++ {
			require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP), "request %d", i+1)
			env.clock.Advance(16 * time.Minute) // never 3 inside 15 minutes
		}
		assert.ErrorIs(t, env.svc.RequestCode(ctx, testAddress, testIP), ErrTooManyRequests, "the 11th inside 24 hours")

		env.clock.Advance(24 * time.Hour)
		assert.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP))
	})

	t.Run("10 codes per IP in an hour, for any addresses", func(t *testing.T) {
		env := newCodeEnv(t)
		for i := 0; i < 10; i++ {
			require.NoError(t, env.svc.RequestCode(ctx, fmt.Sprintf("user%d@example.com", i), testIP), "request %d", i+1)
		}
		assert.ErrorIs(t, env.svc.RequestCode(ctx, "user10@example.com", testIP), ErrTooManyRequests)
		assert.NoError(t, env.svc.RequestCode(ctx, "user10@example.com", testIPSecondClient), "another IP is fine")

		env.clock.Advance(time.Hour + time.Second)
		assert.NoError(t, env.svc.RequestCode(ctx, "user11@example.com", testIP))
	})

	t.Run("a limited request does not replace the code the person already has", func(t *testing.T) {
		env := newCodeEnv(t)
		for i := 0; i < 3; i++ {
			require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP))
		}
		latest := env.lastCode()
		assert.ErrorIs(t, env.svc.RequestCode(ctx, testAddress, testIP), ErrTooManyRequests)
		assert.NoError(t, env.svc.VerifyCode(ctx, testAddress, latest), "the code from the last email still works")
	})

	t.Run("codes and request logs older than a day are deleted", func(t *testing.T) {
		env := newCodeEnv(t)
		require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP))
		assert.Equal(t, 1, env.count(t, `SELECT count(*) FROM "LoginCode"`))

		env.clock.Advance(24*time.Hour + time.Minute)
		require.NoError(t, env.svc.RequestCode(ctx, "sam@example.com", testIPSecondClient))

		assert.Equal(t, 1, env.count(t, `SELECT count(*) FROM "LoginCode"`), "only the new code is left")
		assert.Equal(t, 1, env.count(t, `SELECT count(*) FROM "LoginCodeRequest"`))
		assert.Equal(t, 0, env.count(t, `SELECT count(*) FROM "LoginCode" WHERE email = $1`, testAddress))
	})
}

func TestLoginCodeConcurrencyIntegration(t *testing.T) {
	ctx := context.Background()

	// run calls f n times at once and returns what each call returned
	run := func(n int, f func(i int) error) []error {
		results := make([]error, n)
		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				<-start
				results[i] = f(i)
			}(i)
		}
		close(start)
		wg.Wait()
		return results
	}
	countErrors := func(results []error, target error) int {
		n := 0
		for _, err := range results {
			if err == target {
				n++
			}
		}
		return n
	}

	t.Run("a right code signs in exactly one of many simultaneous tries", func(t *testing.T) {
		env := newCodeEnv(t)
		require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP))
		code := env.lastCode()

		results := run(12, func(int) error { return env.svc.VerifyCode(ctx, testAddress, code) })

		assert.Equal(t, 1, countErrors(results, nil), "one success")
		assert.Equal(t, 11, countErrors(results, ErrCodeRejected))
	})

	t.Run("simultaneous wrong guesses cannot get more than 5 tries", func(t *testing.T) {
		env := newCodeEnv(t)
		require.NoError(t, env.svc.RequestCode(ctx, testAddress, testIP))

		results := run(30, func(int) error { return env.svc.VerifyCode(ctx, testAddress, "000000") })

		assert.Equal(t, 4, countErrors(results, ErrCodeRejected))
		assert.Equal(t, 26, countErrors(results, ErrTooManyAttempts))
		assert.Equal(t, loginCodeMaxAttempts, env.count(t, `SELECT attempts FROM "LoginCode"`))
		assert.ErrorIs(t, env.svc.VerifyCode(ctx, testAddress, env.lastCode()), ErrTooManyAttempts)
	})

	t.Run("simultaneous requests cannot pass the per-email limit", func(t *testing.T) {
		env := newCodeEnv(t)

		results := run(10, func(int) error { return env.svc.RequestCode(ctx, testAddress, testIP) })

		assert.Equal(t, 3, countErrors(results, nil), "exactly the limit")
		assert.Equal(t, 7, countErrors(results, ErrTooManyRequests))
		assert.Len(t, env.mailer.emails(), 3)
		assert.Equal(t, 3, env.count(t, `SELECT count(*) FROM "LoginCodeRequest"`))
		assert.Equal(t, 1, env.count(t, `SELECT count(*) FROM "LoginCode" WHERE used_at IS NULL`), "only the newest code is active")
	})
}

// Signing in end to end: the real code service, caregivers and sessions
func TestLoginWithEmailCodeIntegration(t *testing.T) {
	ctx := context.Background()
	newAuth := func(t *testing.T) (*AuthService, *codeEnv, SessionService) {
		env := newCodeEnv(t)
		sessions := NewSessionService(env.db)
		auth := NewAuthService(nil, caregiver.NewCaregiverService(env.db), nil, sessions).WithLoginCodes(env.svc)
		return auth, env, sessions
	}

	t.Run("a first sign-in creates the caregiver, the next one finds them", func(t *testing.T) {
		auth, env, sessions := newAuth(t)
		require.NoError(t, auth.RequestEmailCode(ctx, "Jo.Smith@Example.com", testIP))
		token, err := auth.LoginWithEmailCode(ctx, "jo.smith@example.com", env.lastCode())
		require.NoError(t, err)
		first, err := sessions.Validate(ctx, token)
		require.NoError(t, err)

		var name, email string
		require.NoError(t, env.db.QueryRow(`SELECT name, email FROM "Caregiver" WHERE id = $1`, first.UserID).Scan(&name, &email))
		assert.Equal(t, "jo.smith", name)
		assert.Equal(t, "jo.smith@example.com", email)

		require.NoError(t, auth.RequestEmailCode(ctx, "jo.smith@example.com", testIP))
		token, err = auth.LoginWithEmailCode(ctx, "JO.SMITH@example.com", env.lastCode())
		require.NoError(t, err)
		second, err := sessions.Validate(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, first.UserID, second.UserID, "the same caregiver")
		assert.Equal(t, 1, env.count(t, `SELECT count(*) FROM "Caregiver" WHERE email = 'jo.smith@example.com'`))
	})

}
