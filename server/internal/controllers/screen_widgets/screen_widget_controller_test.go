package screen_widgets

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

	model "server/internal/models/screen_widgets"
	svc "server/internal/services/screen_widgets"
)

type mockWidgetService struct {
	widget  *model.ScreenWidget
	list    []model.ScreenWidget
	err     error
	created *svc.CreateInput
	updated *svc.UpdateInput
	moved   *svc.MoveInput
	gotID   int
	deleted int
}

func (m *mockWidgetService) CreateWidget(in svc.CreateInput) (*model.ScreenWidget, error) {
	m.created = &in
	return m.widget, m.err
}
func (m *mockWidgetService) ListWidgets(carepartnerID int) ([]model.ScreenWidget, error) {
	m.gotID = carepartnerID
	return m.list, m.err
}
func (m *mockWidgetService) GetWidget(id int) (*model.ScreenWidget, error) {
	m.gotID = id
	return m.widget, m.err
}
func (m *mockWidgetService) UpdateWidget(id int, in svc.UpdateInput) (*model.ScreenWidget, error) {
	m.gotID, m.updated = id, &in
	return m.widget, m.err
}
func (m *mockWidgetService) MoveWidget(id int, in svc.MoveInput) (*model.ScreenWidget, error) {
	m.gotID, m.moved = id, &in
	return m.widget, m.err
}
func (m *mockWidgetService) DeleteWidget(id int) error {
	m.deleted = id
	return m.err
}

func withID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func errorCode(t *testing.T, w *httptest.ResponseRecorder) string {
	var e ScreenWidgetError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &e))
	return e.Code
}

func TestCreateWidgetHandler(t *testing.T) {
	m := &mockWidgetService{widget: &model.ScreenWidget{ID: 9, Type: "weather", Spot: 2, Config: json.RawMessage(`{"zip":"14617"}`)}}
	body := `{"carepartner_id":1,"type":"weather","spot":2,"config":{"zip":"14617"}}`
	w := httptest.NewRecorder()

	NewScreenWidgetController(m).CreateWidgetHandler(w, httptest.NewRequest(http.MethodPost, "/api/screen-widgets", strings.NewReader(body)))

	assert.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, m.created)
	assert.Equal(t, 1, m.created.CarepartnerID)
	assert.Equal(t, 2, *m.created.Spot)
	assert.JSONEq(t, `{"zip":"14617"}`, string(m.created.Config))
	assert.JSONEq(t, `{"id":9,"carepartner_id":0,"type":"weather","spot":2,"config":{"zip":"14617"},"photo_media_id":null,"photo_url":null,"hidden":false}`, w.Body.String())
}

func TestCreateWidgetHandler_Errors(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{svc.ErrInvalidType, http.StatusBadRequest, CodeInvalidType},
		{svc.ErrInvalidConfig, http.StatusBadRequest, CodeInvalidConfig},
		{svc.ErrInvalidPhoto, http.StatusBadRequest, CodeInvalidPhoto},
		{svc.ErrUnknownSpot, http.StatusBadRequest, CodeUnknownSpot},
		{svc.ErrSpotTaken, http.StatusConflict, CodeSpotTaken},
		{svc.ErrCarepartnerMissing, http.StatusNotFound, CodeNotFound},
		{svc.ErrInvalidCarepartner, http.StatusBadRequest, CodeInvalidRequest},
		{errors.New("db down"), http.StatusInternalServerError, CodeInternal},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			w := httptest.NewRecorder()
			NewScreenWidgetController(&mockWidgetService{err: tc.err}).CreateWidgetHandler(w,
				httptest.NewRequest(http.MethodPost, "/api/screen-widgets", strings.NewReader(`{}`)))
			assert.Equal(t, tc.status, w.Code)
			assert.Equal(t, tc.code, errorCode(t, w))
		})
	}

	m := &mockWidgetService{}
	w := httptest.NewRecorder()
	NewScreenWidgetController(m).CreateWidgetHandler(w, httptest.NewRequest(http.MethodPost, "/api/screen-widgets", strings.NewReader(`nope`)))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Nil(t, m.created)
}

func TestListWidgetsHandler(t *testing.T) {
	m := &mockWidgetService{list: []model.ScreenWidget{{ID: 1, Hidden: true}}}
	w := httptest.NewRecorder()
	NewScreenWidgetController(m).ListWidgetsHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen-widgets?carepartner_id=3", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 3, m.gotID)
	assert.Contains(t, w.Body.String(), `"hidden":true`)

	w = httptest.NewRecorder()
	NewScreenWidgetController(m).ListWidgetsHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen-widgets", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetWidgetHandler(t *testing.T) {
	m := &mockWidgetService{widget: &model.ScreenWidget{ID: 4}}
	w := httptest.NewRecorder()
	NewScreenWidgetController(m).GetWidgetHandler(w, withID(httptest.NewRequest(http.MethodGet, "/api/screen-widgets/4", nil), "4"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 4, m.gotID)

	w = httptest.NewRecorder()
	NewScreenWidgetController(&mockWidgetService{err: svc.ErrNotFound}).GetWidgetHandler(w, withID(httptest.NewRequest(http.MethodGet, "/api/screen-widgets/4", nil), "4"))
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	NewScreenWidgetController(m).GetWidgetHandler(w, withID(httptest.NewRequest(http.MethodGet, "/api/screen-widgets/x", nil), "x"))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateWidgetHandler(t *testing.T) {
	m := &mockWidgetService{widget: &model.ScreenWidget{ID: 4}}
	w := httptest.NewRecorder()
	NewScreenWidgetController(m).UpdateWidgetHandler(w, withID(httptest.NewRequest(http.MethodPut, "/api/screen-widgets/4",
		strings.NewReader(`{"config":{"text":"Ana","label":"Nurse"},"photo_media_id":7}`)), "4"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 4, m.gotID)
	assert.JSONEq(t, `{"text":"Ana","label":"Nurse"}`, string(m.updated.Config))
	assert.Equal(t, 7, *m.updated.PhotoMediaID)
}

func TestMoveWidgetHandler(t *testing.T) {
	m := &mockWidgetService{widget: &model.ScreenWidget{ID: 4, Spot: 1}}
	w := httptest.NewRecorder()
	NewScreenWidgetController(m).MoveWidgetHandler(w, withID(httptest.NewRequest(http.MethodPost, "/api/screen-widgets/4/move",
		strings.NewReader(`{"spot":1}`)), "4"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, *m.moved.Spot)

	w = httptest.NewRecorder()
	NewScreenWidgetController(&mockWidgetService{err: svc.ErrSpotTaken}).MoveWidgetHandler(w, withID(httptest.NewRequest(http.MethodPost, "/api/screen-widgets/4/move",
		strings.NewReader(`{"spot":1}`)), "4"))
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, CodeSpotTaken, errorCode(t, w))
}

func TestDeleteWidgetHandler(t *testing.T) {
	m := &mockWidgetService{}
	w := httptest.NewRecorder()
	NewScreenWidgetController(m).DeleteWidgetHandler(w, withID(httptest.NewRequest(http.MethodDelete, "/api/screen-widgets/4", nil), "4"))
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, 4, m.deleted)

	w = httptest.NewRecorder()
	NewScreenWidgetController(&mockWidgetService{err: svc.ErrNotFound}).DeleteWidgetHandler(w, withID(httptest.NewRequest(http.MethodDelete, "/api/screen-widgets/4", nil), "4"))
	assert.Equal(t, http.StatusNotFound, w.Code)
}
