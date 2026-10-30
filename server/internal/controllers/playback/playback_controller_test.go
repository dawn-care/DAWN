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
	return sqlmock.NewRows([]string{"id", "carepartner_id", "media_id", "media_kind", "media_display_name", "scheduled_item_id", "source", "status", "requested_at", "acknowledged_at", "started_at", "ended_at", "expires_at", "rating"}).
		AddRow(1, 2, 3, "video", "Clip", nil, "tap", status, time.Now().UTC(), nil, nil, nil, time.Now().Add(time.Hour), nil)
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
	require.Contains(t, response.Body.String(), `"media_kind":"video"`)
	require.Contains(t, response.Body.String(), `"media_display_name":"Clip"`)
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

func TestListPlaybacksEndpoint(t *testing.T) {
	r, mock := setup(t)
	mock.ExpectQuery(`SELECT .* FROM "Playback"\s+WHERE carepartner_id = \$1 AND started_at IS NOT NULL`).
		WithArgs(2, 1).WillReturnRows(rows("finished"))
	response := call(r, "GET", "/api/playbacks?carepartner_id=2&limit=1", "")
	require.Equal(t, 200, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), `"media_display_name":"Clip"`)
	require.Contains(t, response.Body.String(), `"status":"finished"`)
}

func TestListPlaybacksDefaultsAndEmpty(t *testing.T) {
	r, mock := setup(t)
	// Without a limit, up to 50 are returned; none is [] rather than null
	mock.ExpectQuery(`SELECT .* FROM "Playback"`).WithArgs(2, 50).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	response := call(r, "GET", "/api/playbacks?carepartner_id=2", "")
	require.Equal(t, 200, response.Code, response.Body.String())
	require.Equal(t, "[]\n", response.Body.String())
}

func TestListPlaybacksRejectsBadQueries(t *testing.T) {
	for _, query := range []string{"", "?carepartner_id=0", "?carepartner_id=two", "?carepartner_id=2&limit=0", "?carepartner_id=2&limit=101", "?carepartner_id=2&limit=all"} {
		r, _ := setup(t)
		require.Equal(t, 400, call(r, "GET", "/api/playbacks"+query, "").Code, query)
	}
}

func TestListPlaybacksDatabaseError(t *testing.T) {
	r, mock := setup(t)
	mock.ExpectQuery(`SELECT .* FROM "Playback"`).WillReturnError(errors.New("sensitive database detail"))
	response := call(r, "GET", "/api/playbacks?carepartner_id=2", "")
	require.Equal(t, 500, response.Code)
	require.NotContains(t, response.Body.String(), "sensitive")
}

func TestRatePlaybackEndpoint(t *testing.T) {
	r, mock := setup(t)
	ended := time.Now().UTC().Add(-time.Minute)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FOR UPDATE`).WithArgs(1, 2).WillReturnRows(
		sqlmock.NewRows([]string{"id", "carepartner_id", "media_id", "media_kind", "media_display_name", "scheduled_item_id", "source", "status", "requested_at", "acknowledged_at", "started_at", "ended_at", "expires_at", "rating"}).
			AddRow(1, 2, 3, "video", "Clip", nil, "staff", "finished", ended, ended, ended, ended, ended, nil))
	mock.ExpectExec(`UPDATE "Playback" SET rating`).WithArgs(-1, 1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	response := call(r, "POST", "/api/playbacks/1/rating", `{"carepartner_id":2,"rating":"worse"}`)
	require.Equal(t, 200, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), `"rating":-1`)
}

func TestRatePlaybackRejectsBadRequests(t *testing.T) {
	for _, body := range []string{"", "{}", `{"carepartner_id":2}`, `{"carepartner_id":2,"rating":"great"}`, `{"carepartner_id":2,"rating":"helped","note":"x"}`, `{"carepartner_id":2,"rating":1}`} {
		r, _ := setup(t)
		require.Equal(t, 400, call(r, "POST", "/api/playbacks/1/rating", body).Code, body)
	}
	r, _ := setup(t)
	require.Equal(t, 400, call(r, "POST", "/api/playbacks/zero/rating", `{"carepartner_id":2,"rating":"helped"}`).Code)
}

func TestRatePlaybackConflicts(t *testing.T) {
	long := time.Now().UTC().Add(-time.Hour)
	for name, test := range map[string]struct {
		started, ended any
		want           string
	}{
		"never reached the screen": {nil, nil, "only a play that reached the screen can be rated"},
		"window closed":            {long, long, "the rating window has closed"},
	} {
		t.Run(name, func(t *testing.T) {
			r, mock := setup(t)
			mock.ExpectBegin()
			mock.ExpectQuery(`SELECT .* FOR UPDATE`).WithArgs(1, 2).WillReturnRows(
				sqlmock.NewRows([]string{"id", "carepartner_id", "media_id", "media_kind", "media_display_name", "scheduled_item_id", "source", "status", "requested_at", "acknowledged_at", "started_at", "ended_at", "expires_at", "rating"}).
					AddRow(1, 2, 3, "video", "Clip", nil, "staff", "stopped", long, nil, test.started, test.ended, long, nil))
			mock.ExpectRollback()

			response := call(r, "POST", "/api/playbacks/1/rating", `{"carepartner_id":2,"rating":"helped"}`)
			require.Equal(t, 409, response.Code)
			require.Contains(t, response.Body.String(), test.want)
		})
	}
}

func TestGetPlaybackEndpoint(t *testing.T) {
	r, mock := setup(t)
	ended := time.Date(2026, 10, 28, 14, 1, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT .* FROM "Playback" WHERE id = \$1 AND carepartner_id = \$2`).WithArgs(1, 2).WillReturnRows(
		sqlmock.NewRows([]string{"id", "carepartner_id", "media_id", "media_kind", "media_display_name", "scheduled_item_id", "source", "status", "requested_at", "acknowledged_at", "started_at", "ended_at", "expires_at", "rating"}).
			AddRow(1, 2, 3, "video", "Clip", nil, "staff", "finished", ended, ended, ended.Add(-time.Minute), ended, ended, nil))

	response := call(r, "GET", "/api/playbacks/1?carepartner_id=2", "")
	require.Equal(t, 200, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), `"status":"finished"`)
	require.Contains(t, response.Body.String(), `"rating_closes_at":"2026-10-28T14:06:00Z"`)
}

func TestGetPlaybackEndpointErrors(t *testing.T) {
	for _, path := range []string{"/api/playbacks/zero?carepartner_id=2", "/api/playbacks/1", "/api/playbacks/1?carepartner_id=two"} {
		r, _ := setup(t)
		require.Equal(t, 400, call(r, "GET", path, "").Code, path)
	}
	r, mock := setup(t)
	mock.ExpectQuery(`FROM "Playback"`).WithArgs(1, 2).WillReturnError(sql.ErrNoRows)
	require.Equal(t, 404, call(r, "GET", "/api/playbacks/1?carepartner_id=2", "").Code)
}
