package screen

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	svc "server/internal/services/screen"
)

type mockScreenService struct {
	screen *svc.Screen
	err    error
	gotID  int
}

func (m *mockScreenService) GetScreen(carepartnerID int) (*svc.Screen, error) {
	m.gotID = carepartnerID
	return m.screen, m.err
}

func TestGetScreenHandler(t *testing.T) {
	m := &mockScreenService{screen: &svc.Screen{CarepartnerID: 1, Today: "2026-10-16"}}
	w := httptest.NewRecorder()
	NewScreenController(m).GetScreenHandler(w, httptest.NewRequest(http.MethodGet, "/api/screen?carepartner_id=1", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, m.gotID)
	assert.Contains(t, w.Body.String(), `"today":"2026-10-16"`)
}

func TestGetScreenHandler_Errors(t *testing.T) {
	cases := []struct {
		url    string
		err    error
		status int
	}{
		{"/api/screen", nil, http.StatusBadRequest},
		{"/api/screen?carepartner_id=abc", nil, http.StatusBadRequest},
		{"/api/screen?carepartner_id=9", svc.ErrCarepartnerMissing, http.StatusNotFound},
		{"/api/screen?carepartner_id=1", errors.New("db down"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		NewScreenController(&mockScreenService{err: tc.err}).GetScreenHandler(w, httptest.NewRequest(http.MethodGet, tc.url, nil))
		assert.Equal(t, tc.status, w.Code, tc.url)
	}
}
