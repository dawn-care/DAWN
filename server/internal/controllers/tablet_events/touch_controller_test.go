package tablet_events_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	controller "server/internal/controllers/tablet_events"
	model "server/internal/models/tablet_events"
	routes "server/internal/routes/tablet_events"
	service "server/internal/services/tablet_events"
)

const path = "/api/tablet-events/touches"

// fakeService records what the controller passed it and answers with err
type fakeService struct {
	called        bool
	carepartnerID int
	events        []model.TouchEvent
	err           error
}

func (f *fakeService) RecordTouches(_ context.Context, carepartnerID int, events []model.TouchEvent) error {
	f.called, f.carepartnerID, f.events = true, carepartnerID, events
	return f.err
}

func router(svc controller.TouchServiceInterface) http.Handler {
	r := chi.NewRouter()
	routes.RegisterTabletEventRoutes(r, controller.NewTouchController(svc))
	return r
}

func post(h http.Handler, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(response, request)
	return response
}

// The ticket's example event, as touchLog.ts sends it
const widgetEvent = `{"screen":"always_on","target_type":"widget","widget_type":"weather","widget_id":"12",
	"media_kind":null,"gesture":"tap","count":3,
	"first_at":"2026-10-18T14:00:00.000Z","last_at":"2026-10-18T14:00:02.800Z"}`

const nightEvent = `{"screen":"night","target_type":null,"widget_type":null,"widget_id":null,"media_kind":null,
	"gesture":"long_press","count":1,"first_at":"2026-10-18T22:00:00.000Z","last_at":"2026-10-18T22:00:00.000Z"}`

func batch(events ...string) string {
	return `{"carepartner_id":7,"events":[` + strings.Join(events, ",") + `]}`
}

func str(s string) *string { return &s }

func TestRecordTouches_204AndPassesTheBatch(t *testing.T) {
	svc := &fakeService{}
	response := post(router(svc), batch(widgetEvent, nightEvent))

	require.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
	require.Empty(t, response.Body.String())
	require.Equal(t, 7, svc.carepartnerID)
	first := time.Date(2026, 10, 18, 14, 0, 0, 0, time.UTC)
	require.Equal(t, []model.TouchEvent{
		{Screen: "always_on", TargetType: str("widget"), WidgetType: str("weather"), WidgetID: str("12"),
			Gesture: "tap", Count: 3, FirstAt: first, LastAt: first.Add(2800 * time.Millisecond)},
		{Screen: "night", Gesture: "long_press", Count: 1,
			FirstAt: first.Add(8 * time.Hour), LastAt: first.Add(8 * time.Hour)},
	}, svc.events)
}

// The body never reaches the service when it isn't exactly the declared shape
func TestRecordTouches_MalformedBodyIs400(t *testing.T) {
	withField := func(event, field string) string { return strings.Replace(event, `{`, `{`+field+`,`, 1) }
	cases := map[string]string{
		"not JSON":                     `{"carepartner_id":7,`,
		"empty body":                   ``,
		"array instead of object":      `[` + widgetEvent + `]`,
		"unknown top-level field":      `{"carepartner_id":7,"device":"tablet-1","events":[` + widgetEvent + `]}`,
		"position in an event":         batch(withField(widgetEvent, `"x":120`)),
		"text in an event":             batch(withField(widgetEvent, `"label":"Weather"`)),
		"count not an integer":         batch(strings.Replace(widgetEvent, `"count":3`, `"count":1.5`, 1)),
		"carepartner_id as a string":   `{"carepartner_id":"7","events":[` + widgetEvent + `]}`,
		"widget_id as a number":        batch(strings.Replace(widgetEvent, `"widget_id":"12"`, `"widget_id":12`, 1)),
		"time not RFC 3339":            batch(strings.Replace(widgetEvent, `"2026-10-18T14:00:00.000Z"`, `"18/10/2026 14:00"`, 1)),
		"trailing data":                batch(widgetEvent) + `{}`,
		"too big (over 64 KB of JSON)": `{"carepartner_id":7,"events":[` + widgetEvent + strings.Repeat(" ", 64<<10) + `]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			svc := &fakeService{}
			response := post(router(svc), body)
			require.Equal(t, http.StatusBadRequest, response.Code)
			require.False(t, svc.called)
		})
	}
}

// A full batch of 50 fits under the body limit with room to spare
func TestRecordTouches_FullBatchFitsTheBodyLimit(t *testing.T) {
	events := make([]string, service.MaxBatch)
	for i := range events {
		events[i] = widgetEvent
	}
	svc := &fakeService{}
	require.Equal(t, http.StatusNoContent, post(router(svc), batch(events...)).Code)
	require.Len(t, svc.events, service.MaxBatch)
}

func TestRecordTouches_ServiceErrors(t *testing.T) {
	t.Run("invalid input is 400 with the reason", func(t *testing.T) {
		svc := &fakeService{err: fmt.Errorf("event 2: %w: unknown gesture", service.ErrInvalidInput)}
		response := post(router(svc), batch(widgetEvent))
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.JSONEq(t, `{"error":"event 2: invalid touch batch: unknown gesture"}`, response.Body.String())
	})
	t.Run("anything else is 500 without the database's words", func(t *testing.T) {
		svc := &fakeService{err: errors.New(`record touches: pq: relation "TabletTouch" does not exist`)}
		response := post(router(svc), batch(widgetEvent))
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.JSONEq(t, `{"error":"could not store touches"}`, response.Body.String())
	})
}

// Through the real service: the status codes the tablet's report queue acts on
func TestRecordTouches_ThroughTheService(t *testing.T) {
	qInsert := regexp.QuoteMeta(`INSERT INTO "TabletTouch"`)
	setup := func(t *testing.T) (http.Handler, sqlmock.Sqlmock) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, mock.ExpectationsWereMet()); db.Close() })
		return router(service.NewTouchService(db)), mock
	}

	t.Run("stored: 204", func(t *testing.T) {
		h, mock := setup(t)
		mock.ExpectBegin()
		mock.ExpectExec(qInsert).WithArgs(7, "always_on", "widget", "weather", "12", nil, "tap", 3, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(qInsert).WithArgs(7, "night", nil, nil, nil, nil, "long_press", 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectCommit()
		require.Equal(t, http.StatusNoContent, post(h, batch(widgetEvent, nightEvent)).Code)
	})
	t.Run("bad event: 400, nothing queried", func(t *testing.T) {
		h, _ := setup(t)
		response := post(h, batch(widgetEvent, strings.Replace(nightEvent, `"count":1`, `"count":0`, 1)))
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "event 1")
	})
	t.Run("missing carepartner_id or events: 400", func(t *testing.T) {
		h, _ := setup(t)
		require.Equal(t, http.StatusBadRequest, post(h, `{"events":[`+widgetEvent+`]}`).Code)
		require.Equal(t, http.StatusBadRequest, post(h, `{"carepartner_id":7}`).Code)
		require.Equal(t, http.StatusBadRequest, post(h, `{"carepartner_id":7,"events":[]}`).Code)
	})
	t.Run("unknown carepartner: 400, rolled back", func(t *testing.T) {
		h, mock := setup(t)
		mock.ExpectBegin()
		mock.ExpectExec(qInsert).WillReturnError(&pq.Error{Code: "23503"})
		mock.ExpectRollback()
		require.Equal(t, http.StatusBadRequest, post(h, batch(widgetEvent)).Code)
	})
	t.Run("database down: 500, so the tablet retries", func(t *testing.T) {
		h, mock := setup(t)
		mock.ExpectBegin().WillReturnError(errors.New("connection refused"))
		require.Equal(t, http.StatusInternalServerError, post(h, batch(widgetEvent)).Code)
	})
}
