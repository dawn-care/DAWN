package family

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	model "server/internal/models/family"
	svc "server/internal/services/family"
)

// mockFamilyService implements FamilyServiceInterface and records what it was given
type mockFamilyService struct {
	err     error
	members []model.FamilyMember
	listed  int
	created *svc.CreateInput
}

func (m *mockFamilyService) ListFamilyMembers(carepartnerID int) ([]model.FamilyMember, error) {
	m.listed = carepartnerID
	return m.members, m.err
}

func (m *mockFamilyService) CreateFamilyMember(in svc.CreateInput) (*model.FamilyMember, error) {
	m.created = &in
	if m.err != nil {
		return nil, m.err
	}
	return &model.FamilyMember{ID: 9, CarepartnerID: in.CarepartnerID, Name: in.Name}, nil
}

func decodeError(t *testing.T, w *httptest.ResponseRecorder) FamilyError {
	var e FamilyError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &e))
	return e
}

func TestListFamilyMembersHandler(t *testing.T) {
	service := &mockFamilyService{members: []model.FamilyMember{{ID: 1, Name: "Elena"}}}
	w := httptest.NewRecorder()
	NewFamilyController(service).ListFamilyMembersHandler(w, httptest.NewRequest(http.MethodGet, "/api/family-members?carepartner_id=1", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, service.listed)
	assert.Contains(t, w.Body.String(), `"name":"Elena"`)
	assert.Contains(t, w.Body.String(), `"guidance_note":null`)
}

func TestListFamilyMembersHandler_Errors(t *testing.T) {
	for _, query := range []string{"", "?carepartner_id=0", "?carepartner_id=John"} {
		w := httptest.NewRecorder()
		NewFamilyController(&mockFamilyService{}).ListFamilyMembersHandler(w, httptest.NewRequest(http.MethodGet, "/api/family-members"+query, nil))
		assert.Equal(t, http.StatusBadRequest, w.Code, query)
		assert.Equal(t, CodeInvalidRequest, decodeError(t, w).Code)
	}

	w := httptest.NewRecorder()
	NewFamilyController(&mockFamilyService{err: errors.New("sensitive database detail")}).
		ListFamilyMembersHandler(w, httptest.NewRequest(http.MethodGet, "/api/family-members?carepartner_id=1", nil))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), "sensitive")
}

func TestCreateFamilyMemberHandler(t *testing.T) {
	service := &mockFamilyService{}
	w := httptest.NewRecorder()
	NewFamilyController(service).CreateFamilyMemberHandler(w, httptest.NewRequest(http.MethodPost, "/api/family-members",
		strings.NewReader(`{"carepartner_id":1,"name":"Elena","relationship":"Daughter"}`)))

	assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	assert.Equal(t, svc.CreateInput{CarepartnerID: 1, Name: "Elena", Relationship: "Daughter"}, *service.created)
}

func TestCreateFamilyMemberHandler_Errors(t *testing.T) {
	for _, body := range []string{`not json`, `{"carepartner_id":"one","name":"Elena"}`, `{"carepartner_id":1,"name":"Elena","age":40}`} {
		service := &mockFamilyService{}
		w := httptest.NewRecorder()
		NewFamilyController(service).CreateFamilyMemberHandler(w, httptest.NewRequest(http.MethodPost, "/api/family-members", strings.NewReader(body)))
		assert.Equal(t, http.StatusBadRequest, w.Code, body)
		assert.Nil(t, service.created, body)
	}

	cases := []struct {
		err  error
		code string
	}{
		{svc.ErrInvalidCarepartner, CodeInvalidRequest},
		{svc.ErrInvalidName, CodeInvalidName},
		{svc.ErrInvalidRelationship, CodeInvalidRelationship},
		{svc.ErrInvalidGuidance, CodeInvalidGuidance},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		NewFamilyController(&mockFamilyService{err: tc.err}).CreateFamilyMemberHandler(w,
			httptest.NewRequest(http.MethodPost, "/api/family-members", strings.NewReader(`{"carepartner_id":1,"name":"x"}`)))
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, tc.code, decodeError(t, w).Code)
	}
}
