package screen_widgets

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	model "server/internal/models/screen_widgets"
	svc "server/internal/services/screen_widgets"
)

type mockWidgetService struct {
	layout *model.Layout
	err    error
	gotID  int
	saved  *model.LayoutInput
}

func (m *mockWidgetService) GetLayout(carepartnerID int) (*model.Layout, error) {
	m.gotID = carepartnerID
	return m.layout, m.err
}
func (m *mockWidgetService) SaveLayout(carepartnerID int, in model.LayoutInput) (*model.Layout, error) {
	m.gotID, m.saved = carepartnerID, &in
	return m.layout, m.err
}

func errorCode(t *testing.T, w *httptest.ResponseRecorder) string {
	var e ScreenWidgetError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &e))
	return e.Code
}

var weatherLayout = &model.Layout{Widgets: []model.ScreenWidget{{
	ID: 2, CarepartnerID: 1, Type: "weather",
	TopLeft: model.Cell{Col: 2, Row: 0}, BottomRight: model.Cell{Col: 3, Row: 0},
	Config: json.RawMessage(`{"zip":"14617"}`),
}}}

func TestGetLayoutHandler(t *testing.T) {
	m := &mockWidgetService{layout: weatherLayout}
	w := httptest.NewRecorder()
	NewScreenWidgetController(m).GetLayoutHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen-layout?carepartner_id=1", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, m.gotID)
	assert.JSONEq(t, `{"widgets":[{"id":2,"carepartner_id":1,"type":"weather",
		"top_left":{"col":2,"row":0},"bottom_right":{"col":3,"row":0},
		"config":{"zip":"14617"},"photo_media_id":null,"photo_url":null}]}`, w.Body.String())

	w = httptest.NewRecorder()
	NewScreenWidgetController(m).GetLayoutHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen-layout", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, CodeInvalidRequest, errorCode(t, w))

	w = httptest.NewRecorder()
	NewScreenWidgetController(&mockWidgetService{err: svc.ErrCarepartnerMissing}).GetLayoutHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen-layout?carepartner_id=9", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSaveLayoutHandler(t *testing.T) {
	m := &mockWidgetService{layout: weatherLayout}
	body := `{"widgets":[
		{"id":2,"type":"weather","top_left":{"col":2,"row":0},"bottom_right":{"col":3,"row":0},"config":{"zip":"14617"}},
		{"type":"photo","top_left":{"col":0,"row":0},"bottom_right":{"col":1,"row":1},"photo_media_id":4}
	]}`
	w := httptest.NewRecorder()
	NewScreenWidgetController(m).SaveLayoutHandler(w, httptest.NewRequest(http.MethodPut, "/api/screen-layout?carepartner_id=1", strings.NewReader(body)))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, m.gotID)
	require.NotNil(t, m.saved.Widgets)
	widgets := *m.saved.Widgets
	require.Len(t, widgets, 2)
	assert.Equal(t, 2, *widgets[0].ID)
	assert.Nil(t, widgets[1].ID, "a new widget has no id")
	assert.Equal(t, model.Cell{Col: 3, Row: 0}, *widgets[0].BottomRight)
	assert.Equal(t, 4, *widgets[1].PhotoMediaID)
	assert.Contains(t, w.Body.String(), `"widgets":[{"id":2`)
}

func TestSaveLayoutHandler_WidgetsKey(t *testing.T) {
	cases := []struct {
		body    string
		present bool
	}{
		{`{"widgets":[]}`, true},
		{`{}`, false},
		{`{"widgets":null}`, false},
		{`{"widget":[]}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.body, func(t *testing.T) {
			m := &mockWidgetService{layout: &model.Layout{Widgets: []model.ScreenWidget{}}}
			w := httptest.NewRecorder()
			NewScreenWidgetController(m).SaveLayoutHandler(w, httptest.NewRequest(http.MethodPut, "/api/screen-layout?carepartner_id=1", strings.NewReader(tc.body)))
			require.NotNil(t, m.saved)
			assert.Equal(t, tc.present, m.saved.Widgets != nil, "only an explicit list reaches the service as a list")
		})
	}
}

func TestSaveLayoutHandler_Errors(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{svc.ErrInvalidRequest, http.StatusBadRequest, CodeInvalidRequest},
		{svc.ErrInvalidCarepartner, http.StatusBadRequest, CodeInvalidRequest},
		{svc.ErrInvalidType, http.StatusBadRequest, CodeInvalidType},
		{svc.ErrInvalidPosition, http.StatusBadRequest, CodeInvalidPosition},
		{svc.ErrInvalidConfig, http.StatusBadRequest, CodeInvalidConfig},
		{svc.ErrInvalidPhoto, http.StatusBadRequest, CodeInvalidPhoto},
		{svc.ErrOverlap, http.StatusBadRequest, CodeOverlap},
		{svc.ErrNotFound, http.StatusNotFound, CodeNotFound},
		{svc.ErrCarepartnerMissing, http.StatusNotFound, CodeNotFound},
		{errors.New("db down"), http.StatusInternalServerError, CodeInternal},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			// Service errors arrive wrapped with the widget's index
			err := errors.Join(errors.New("widgets[1]"), tc.err)
			w := httptest.NewRecorder()
			NewScreenWidgetController(&mockWidgetService{err: err}).SaveLayoutHandler(w,
				httptest.NewRequest(http.MethodPut, "/api/screen-layout?carepartner_id=1", strings.NewReader(`{"widgets":[]}`)))
			assert.Equal(t, tc.status, w.Code)
			assert.Equal(t, tc.code, errorCode(t, w))
		})
	}

	m := &mockWidgetService{}
	w := httptest.NewRecorder()
	NewScreenWidgetController(m).SaveLayoutHandler(w, httptest.NewRequest(http.MethodPut, "/api/screen-layout?carepartner_id=1", strings.NewReader(`nope`)))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Nil(t, m.saved, "a body that isn't JSON never reaches the service")

	w = httptest.NewRecorder()
	NewScreenWidgetController(m).SaveLayoutHandler(w, httptest.NewRequest(http.MethodPut, "/api/screen-layout", strings.NewReader(`{"widgets":[]}`)))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Nil(t, m.saved)
}
