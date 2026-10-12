package screen_settings

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	model "server/internal/models/screen_settings"
	svc "server/internal/services/screen_settings"
)

type mockScreenSettingsService struct {
	settings     *model.ScreenSettings
	err          error
	gotID        int
	gotUpdate    svc.UpdateInput
	updateCalled bool
}

func (m *mockScreenSettingsService) GetPresets() []svc.Preset { return svc.Presets }

func (m *mockScreenSettingsService) GetScreenSettings(carepartnerID int) (*model.ScreenSettings, error) {
	m.gotID = carepartnerID
	return m.settings, m.err
}

func (m *mockScreenSettingsService) UpdateScreenSettings(carepartnerID int, in svc.UpdateInput) (*model.ScreenSettings, error) {
	m.gotID, m.gotUpdate, m.updateCalled = carepartnerID, in, true
	return m.settings, m.err
}

func decodeError(t *testing.T, w *httptest.ResponseRecorder) ScreenSettingsError {
	var e ScreenSettingsError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &e))
	return e
}

var defaults = &model.ScreenSettings{CarepartnerID: 1, LayoutPreset: "balanced", WakeTime: "07:00", BedTime: "21:00"}

func TestGetScreenSettingsHandler(t *testing.T) {
	m := &mockScreenSettingsService{settings: defaults}
	w := httptest.NewRecorder()

	NewScreenSettingsController(m).GetScreenSettingsHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen-settings?carepartner_id=1", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, m.gotID)
	assert.JSONEq(t, `{"carepartner_id":1,"layout_preset":"balanced","wake_time":"07:00","bed_time":"21:00"}`, w.Body.String())
}

func TestGetScreenSettingsHandler_Errors(t *testing.T) {
	w := httptest.NewRecorder()
	NewScreenSettingsController(&mockScreenSettingsService{}).GetScreenSettingsHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen-settings", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, CodeInvalidRequest, decodeError(t, w).Code)

	w = httptest.NewRecorder()
	NewScreenSettingsController(&mockScreenSettingsService{err: svc.ErrNotFound}).GetScreenSettingsHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen-settings?carepartner_id=9", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, CodeNotFound, decodeError(t, w).Code)

	w = httptest.NewRecorder()
	NewScreenSettingsController(&mockScreenSettingsService{err: errors.New("db down")}).GetScreenSettingsHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen-settings?carepartner_id=1", nil))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateScreenSettingsHandler(t *testing.T) {
	m := &mockScreenSettingsService{settings: defaults}
	body := `{"layout_preset":"balanced","wake_time":"07:00","bed_time":"21:00"}`
	w := httptest.NewRecorder()

	NewScreenSettingsController(m).UpdateScreenSettingsHandler(w, httptest.NewRequest(http.MethodPut, "/api/screen-settings?carepartner_id=1", strings.NewReader(body)))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, m.gotID)
	assert.Equal(t, svc.UpdateInput{LayoutPreset: "balanced", WakeTime: "07:00", BedTime: "21:00"}, m.gotUpdate)
}

func TestUpdateScreenSettingsHandler_Errors(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"bad preset", svc.ErrInvalidPreset, http.StatusBadRequest, CodeInvalidPreset},
		{"bad wake", svc.ErrInvalidWakeTime, http.StatusBadRequest, CodeInvalidWakeTime},
		{"bad bed", svc.ErrInvalidBedTime, http.StatusBadRequest, CodeInvalidBedTime},
		{"missing", svc.ErrNotFound, http.StatusNotFound, CodeNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			NewScreenSettingsController(&mockScreenSettingsService{err: tc.err}).UpdateScreenSettingsHandler(w,
				httptest.NewRequest(http.MethodPut, "/api/screen-settings?carepartner_id=1", strings.NewReader(`{}`)))
			assert.Equal(t, tc.status, w.Code)
			assert.Equal(t, tc.code, decodeError(t, w).Code)
		})
	}

	m := &mockScreenSettingsService{}
	w := httptest.NewRecorder()
	NewScreenSettingsController(m).UpdateScreenSettingsHandler(w,
		httptest.NewRequest(http.MethodPut, "/api/screen-settings?carepartner_id=1", strings.NewReader(`not json`)))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, m.updateCalled)

	w = httptest.NewRecorder()
	NewScreenSettingsController(m).UpdateScreenSettingsHandler(w,
		httptest.NewRequest(http.MethodPut, "/api/screen-settings", strings.NewReader(`{}`)))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, m.updateCalled)
}

func TestGetPresetsHandler(t *testing.T) {
	w := httptest.NewRecorder()
	NewScreenSettingsController(&mockScreenSettingsService{}).GetPresetsHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen-settings/presets", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	var presets []svc.Preset
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &presets))
	require.Len(t, presets, 1)
	assert.Equal(t, "balanced", presets[0].Name)
	assert.Equal(t, []svc.Region{
		{Name: "photo_frame", Slots: 1, Accepts: []string{"photo"}},
		{Name: "tiles", Slots: 3, Accepts: []string{"reminder", "weather", "sports", "news"}},
		{Name: "plate", Slots: 1, Accepts: []string{"message"}},
	}, presets[0].Regions)
}
