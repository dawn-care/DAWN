package scheduled_items

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	model "server/internal/models/scheduled_items"
	svc "server/internal/services/scheduled_items"
)

type mockItemService struct {
	item        *model.ScheduledItem
	items       []model.ScheduledItem
	occurrences []svc.Occurrence
	err         error
	input       *svc.ItemInput
	gotID       int
	scope, date string
	from, to    string
}

func (m *mockItemService) CreateItem(in svc.ItemInput) (*model.ScheduledItem, error) {
	m.input = &in
	return m.item, m.err
}
func (m *mockItemService) GetItem(id int) (*model.ScheduledItem, error) {
	m.gotID = id
	return m.item, m.err
}
func (m *mockItemService) ListItems(carepartnerID int) ([]model.ScheduledItem, error) {
	m.gotID = carepartnerID
	return m.items, m.err
}
func (m *mockItemService) Calendar(carepartnerID int, from, to string) ([]svc.Occurrence, error) {
	m.gotID, m.from, m.to = carepartnerID, from, to
	return m.occurrences, m.err
}
func (m *mockItemService) UpdateItem(id int, in svc.ItemInput) (*model.ScheduledItem, error) {
	m.gotID, m.input = id, &in
	return m.item, m.err
}
func (m *mockItemService) DeleteItem(id int, scope, date string) error {
	m.gotID, m.scope, m.date = id, scope, date
	return m.err
}

func withID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func errorCode(t *testing.T, w *httptest.ResponseRecorder) string {
	var e ScheduledItemError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &e))
	return e.Code
}

func TestCreateItemHandler(t *testing.T) {
	m := &mockItemService{item: &model.ScheduledItem{ID: 5, Shape: "video"}}
	body := `{"carepartner_id":1,"media_id":2,"start_date":"2026-10-12","start_time":"10:00","repeat":"daily","end_date":null}`
	w := httptest.NewRecorder()

	NewScheduledItemController(m).CreateItemHandler(w, httptest.NewRequest(http.MethodPost, "/api/scheduled-items", strings.NewReader(body)))

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, svc.ItemInput{CarepartnerID: 1, MediaID: 2, StartDate: "2026-10-12", StartTime: "10:00", Repeat: "daily"}, *m.input)
}

func TestCreateItemHandler_ErrorCodes(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{svc.ErrNotOn15MinStep, 400, CodeNotOn15MinStep},
		{svc.ErrOutsideAwakeHours, 400, CodeOutsideAwakeHours},
		{svc.ErrInvalidShape, 400, CodeInvalidShape},
		{svc.ErrInvalidMedia, 400, CodeInvalidMedia},
		{svc.ErrInvalidRepeat, 400, CodeInvalidRepeat},
		{svc.ErrInvalidDate, 400, CodeInvalidDate},
		{svc.ErrInvalidTime, 400, CodeInvalidTime},
		{svc.ErrCarepartnerMissing, 404, CodeNotFound},
		{errors.New("db down"), 500, CodeInternal},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			w := httptest.NewRecorder()
			NewScheduledItemController(&mockItemService{err: tc.err}).CreateItemHandler(w,
				httptest.NewRequest(http.MethodPost, "/api/scheduled-items", strings.NewReader(`{}`)))
			assert.Equal(t, tc.status, w.Code)
			assert.Equal(t, tc.code, errorCode(t, w))
		})
	}

	w := httptest.NewRecorder()
	NewScheduledItemController(&mockItemService{}).CreateItemHandler(w,
		httptest.NewRequest(http.MethodPost, "/api/scheduled-items", strings.NewReader(`nope`)))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListItemsHandler(t *testing.T) {
	m := &mockItemService{items: []model.ScheduledItem{{ID: 5}}}
	w := httptest.NewRecorder()
	NewScheduledItemController(m).ListItemsHandler(w, httptest.NewRequest(http.MethodGet, "/api/scheduled-items?carepartner_id=1", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, m.gotID)

	w = httptest.NewRecorder()
	NewScheduledItemController(m).ListItemsHandler(w, httptest.NewRequest(http.MethodGet, "/api/scheduled-items", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAndUpdateItemHandler(t *testing.T) {
	m := &mockItemService{item: &model.ScheduledItem{ID: 5}}
	w := httptest.NewRecorder()
	NewScheduledItemController(m).GetItemHandler(w, withID(httptest.NewRequest(http.MethodGet, "/api/scheduled-items/5", nil), "5"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 5, m.gotID)

	w = httptest.NewRecorder()
	NewScheduledItemController(m).UpdateItemHandler(w, withID(httptest.NewRequest(http.MethodPut, "/api/scheduled-items/5",
		strings.NewReader(`{"media_id":2,"start_date":"2026-10-12","start_time":"11:30","repeat":"daily"}`)), "5"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "11:30", m.input.StartTime)

	w = httptest.NewRecorder()
	NewScheduledItemController(&mockItemService{err: svc.ErrNotFound}).GetItemHandler(w, withID(httptest.NewRequest(http.MethodGet, "/api/scheduled-items/5", nil), "5"))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteItemHandler(t *testing.T) {
	m := &mockItemService{}
	w := httptest.NewRecorder()
	NewScheduledItemController(m).DeleteItemHandler(w, withID(httptest.NewRequest(http.MethodDelete,
		"/api/scheduled-items/5?scope=occurrence&date=2026-10-19", nil), "5"))

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "occurrence", m.scope)
	assert.Equal(t, "2026-10-19", m.date)

	w = httptest.NewRecorder()
	NewScheduledItemController(&mockItemService{err: svc.ErrNotAnOccurrence}).DeleteItemHandler(w, withID(httptest.NewRequest(http.MethodDelete,
		"/api/scheduled-items/5?scope=occurrence&date=2026-10-20", nil), "5"))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, CodeNotAnOccurrence, errorCode(t, w))

	w = httptest.NewRecorder()
	NewScheduledItemController(&mockItemService{err: svc.ErrInvalidScope}).DeleteItemHandler(w, withID(httptest.NewRequest(http.MethodDelete,
		"/api/scheduled-items/5", nil), "5"))
	assert.Equal(t, CodeInvalidScope, errorCode(t, w))
}

func TestCalendarHandler(t *testing.T) {
	m := &mockItemService{occurrences: []svc.Occurrence{{ScheduledItemID: 5, Date: "2026-10-12", Time: "10:00"}}}
	w := httptest.NewRecorder()
	NewScheduledItemController(m).CalendarHandler(w, httptest.NewRequest(http.MethodGet, "/api/calendar?carepartner_id=1&from=2026-10-12&to=2026-10-18", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "2026-10-12", m.from)
	assert.Equal(t, "2026-10-18", m.to)
	assert.Contains(t, w.Body.String(), `"scheduled_item_id":5`)

	w = httptest.NewRecorder()
	NewScheduledItemController(&mockItemService{err: svc.ErrInvalidRange}).CalendarHandler(w, httptest.NewRequest(http.MethodGet, "/api/calendar?carepartner_id=1&from=x&to=y", nil))
	assert.Equal(t, CodeInvalidRange, errorCode(t, w))
}
