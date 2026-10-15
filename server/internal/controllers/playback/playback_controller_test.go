package playback_test

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	controller "server/internal/controllers/playback"
	"server/internal/realtime"
	routes "server/internal/routes/playback"
	service "server/internal/services/playback"
)

// noopNotifier stands in for the tablet hub in tests that do not push commands.
type noopNotifier struct{}

func (noopNotifier) Send(int, realtime.Message) bool { return false }

func setup(t *testing.T) (http.Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, mock.ExpectationsWereMet()); db.Close() })
	r := chi.NewRouter()
	routes.RegisterPlaybackRoutes(r, controller.NewPlaybackController(service.NewPlaybackService(db), noopNotifier{}))
	return r, mock
}

func call(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(response, request)
	return response
}

func rows(status string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "carepartner_id", "media_id", "scheduled_item_id", "source", "status", "requested_at", "acknowledged_at", "started_at", "ended_at", "expires_at", "rating"}).
		AddRow(1, 2, 3, nil, "tap", status, time.Now().UTC(), nil, nil, nil, time.Now().Add(time.Hour), nil)
}

func TestCreatePlaybackEndpoint(t *testing.T) {
	r, mock := setup(t)
	expiry := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	mock.ExpectQuery(`INSERT INTO "Playback"`).WithArgs(2, 3, nil, "tap", sqlmock.AnyArg(), expiry).WillReturnRows(rows("requested"))
	response := call(r, "POST", "/api/playbacks", `{"carepartner_id":2,"media_id":3,"source":"tap","expires_at":"`+expiry.Format(time.RFC3339)+`"}`)
	require.Equal(t, 201, response.Code, response.Body.String())
	require.Equal(t, "application/json", response.Header().Get("Content-Type"))
	require.Contains(t, response.Body.String(), `"status":"requested"`)
	require.Contains(t, response.Body.String(), `"rating":null`)
	require.Contains(t, response.Body.String(), `"media_id":3`)
}

func TestCreateScheduledPlaybackEndpoint(t *testing.T) {
	r, mock := setup(t)
	expiry := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	mock.ExpectQuery(`INSERT INTO "Playback"`).WithArgs(2, 3, 5, "scheduled", sqlmock.AnyArg(), expiry).WillReturnRows(rows("requested"))
	response := call(r, "POST", "/api/playbacks", `{"carepartner_id":2,"media_id":3,"scheduled_item_id":5,"source":"scheduled","expires_at":"`+expiry.Format(time.RFC3339)+`"}`)
	require.Equal(t, 201, response.Code, response.Body.String())
}

// Clients built against the old model send module_id; it's rejected before any query
func TestCreateRejectsModuleID(t *testing.T) {
	r, _ := setup(t)
	expiry := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	response := call(r, "POST", "/api/playbacks", `{"carepartner_id":2,"module_id":3,"source":"tap","expires_at":"`+expiry+`"}`)
	require.Equal(t, 400, response.Code)
}

func TestStatusEndpoint(t *testing.T) {
	r, mock := setup(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FOR UPDATE`).WithArgs(1, 2).WillReturnRows(rows("requested"))
	mock.ExpectExec(`UPDATE "Playback"`).WithArgs("acknowledged", sqlmock.AnyArg(), nil, nil, 1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	response := call(r, "PATCH", "/api/playbacks/1/status", `{"carepartner_id":2,"status":"acknowledged"}`)
	require.Equal(t, 200, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), `"status":"acknowledged"`)
}

func TestRejectInvalidBodiesWithoutDatabaseAccess(t *testing.T) {
	for _, body := range []string{"", "null", "[]", "{}", "{", `{} {}`, `{"notes":"private"}`, `{"rating":3}`, `{"requested_at":"2026-01-01T00:00:00Z"}`, `{"carepartner_id":"two"}`, `{"expires_at":"yesterday"}`, `{"source":"` + strings.Repeat("x", 5000) + `"}`} {
		t.Run(body[:min(len(body), 35)], func(t *testing.T) {
			r, _ := setup(t)
			require.Equal(t, 400, call(r, "POST", "/api/playbacks", body).Code)
		})
	}
	for _, body := range []string{"null", "{}", `{"carepartner_id":2,"status":"unknown"}`, `{"carepartner_id":2,"status":"started","reason":"private"}`, `{"carepartner_id":2,"status":"started"} true`} {
		r, _ := setup(t)
		require.Equal(t, 400, call(r, "PATCH", "/api/playbacks/1/status", body).Code)
	}
	for _, id := range []string{"zero", "0", "-1"} {
		r, _ := setup(t)
		require.Equal(t, 400, call(r, "PATCH", "/api/playbacks/"+id+"/status", `{"carepartner_id":2,"status":"started"}`).Code)
	}
}

func TestStatusErrorResponses(t *testing.T) {
	for _, test := range []struct {
		name string
		code int
	}{{"missing", 404}, {"conflict", 409}, {"database", 500}} {
		t.Run(test.name, func(t *testing.T) {
			r, mock := setup(t)
			mock.ExpectBegin()
			query := mock.ExpectQuery(`SELECT .*carepartner_id = \$2 FOR UPDATE`).WithArgs(1, 2)
			switch test.name {
			case "missing":
				query.WillReturnError(sql.ErrNoRows)
			case "database":
				query.WillReturnError(errors.New("sensitive database detail"))
			default:
				query.WillReturnRows(rows("finished"))
			}
			mock.ExpectRollback()
			response := call(r, "PATCH", "/api/playbacks/1/status", `{"carepartner_id":2,"status":"started"}`)
			require.Equal(t, test.code, response.Code)
			require.NotContains(t, response.Body.String(), "sensitive")
		})
	}
}

func TestCreateErrorResponses(t *testing.T) {
	for _, test := range []struct {
		cause error
		code  int
	}{{sql.ErrNoRows, 400}, {errors.New("sensitive database detail"), 500}} {
		r, mock := setup(t)
		mock.ExpectQuery(`INSERT INTO "Playback"`).WillReturnError(test.cause)
		response := call(r, "POST", "/api/playbacks", `{"carepartner_id":2,"media_id":3,"scheduled_item_id":5,"source":"scheduled","expires_at":"`+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+`"}`)
		require.Equal(t, test.code, response.Code)
		require.NotContains(t, response.Body.String(), "sensitive")
	}
}
