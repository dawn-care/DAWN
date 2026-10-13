package playback_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	controller "server/internal/controllers/playback"
	model "server/internal/models/playback"
	"server/internal/realtime"
	routes "server/internal/routes/playback"
	service "server/internal/services/playback"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// mockService implements PlaybackServiceInterface.
type mockService struct {
	requestStaffPlay func(ctx context.Context, carepartnerID, moduleID int) (*model.Playback, error)
	requestStop      func(ctx context.Context, id, carepartnerID int) (*model.Playback, service.StopOutcome, error)
}

func (m *mockService) CreatePlayback(context.Context, int, int, string, time.Time) (*model.Playback, error) {
	panic("not used by command endpoints")
}

func (m *mockService) UpdateStatus(context.Context, int, int, string) (*model.Playback, error) {
	panic("not used by command endpoints")
}

func (m *mockService) RequestStop(ctx context.Context, id, carepartnerID int) (*model.Playback, service.StopOutcome, error) {
	return m.requestStop(ctx, id, carepartnerID)
}

func (m *mockService) RequestStaffPlay(ctx context.Context, carepartnerID, moduleID int) (*model.Playback, error) {
	return m.requestStaffPlay(ctx, carepartnerID, moduleID)
}

// recordingNotifier stands in for the tablet hub and records what was pushed.
type recordingNotifier struct {
	online bool
	sent   []sentMessage
}

type sentMessage struct {
	carepartnerID int
	msg           realtime.Message
}

func (n *recordingNotifier) Send(carepartnerID int, msg realtime.Message) bool {
	n.sent = append(n.sent, sentMessage{carepartnerID, msg})
	return n.online
}

func commandRouter(svc *mockService, notifier *recordingNotifier) http.Handler {
	r := chi.NewRouter()
	routes.RegisterPlaybackRoutes(r, controller.NewPlaybackController(svc, notifier))
	return r
}

func post(r http.Handler, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) controller.CommandResponse {
	t.Helper()
	var resp controller.CommandResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp
}

func TestPlayPushesCommandToTablet(t *testing.T) {
	expires := time.Now().UTC().Add(30 * time.Second).Truncate(time.Second)
	svc := &mockService{requestStaffPlay: func(_ context.Context, cp, module int) (*model.Playback, error) {
		require.Equal(t, 2, cp)
		require.Equal(t, 3, module)
		return &model.Playback{ID: 9, CarepartnerID: cp, ModuleID: module, Source: "staff", Status: "requested", ExpiresAt: expires}, nil
	}}
	notifier := &recordingNotifier{online: true}

	rec := post(commandRouter(svc, notifier), "/api/playbacks/play", `{"carepartner_id":2,"module_id":3}`)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	resp := decode(t, rec)
	require.True(t, resp.Delivered)
	require.Equal(t, 9, resp.Playback.ID)

	require.Len(t, notifier.sent, 1)
	require.Equal(t, 2, notifier.sent[0].carepartnerID)
	require.Equal(t, realtime.PlayCommand(9, 3, expires), notifier.sent[0].msg)
}

func TestPlayWithTabletOfflineStillRecordsCommand(t *testing.T) {
	svc := &mockService{requestStaffPlay: func(_ context.Context, cp, module int) (*model.Playback, error) {
		return &model.Playback{ID: 9, CarepartnerID: cp, ModuleID: module, Status: "requested", ExpiresAt: time.Now().Add(30 * time.Second)}, nil
	}}
	notifier := &recordingNotifier{online: false}

	rec := post(commandRouter(svc, notifier), "/api/playbacks/play", `{"carepartner_id":2,"module_id":3}`)

	// The row exists, so the command is resent if the tablet reconnects in time.
	require.Equal(t, http.StatusCreated, rec.Code)
	require.False(t, decode(t, rec).Delivered)
}

func TestPlayRejectedByServiceSendsNothing(t *testing.T) {
	svc := &mockService{requestStaffPlay: func(context.Context, int, int) (*model.Playback, error) {
		return nil, service.ErrInvalidInput // e.g. the module belongs to another carepartner
	}}
	notifier := &recordingNotifier{online: true}

	rec := post(commandRouter(svc, notifier), "/api/playbacks/play", `{"carepartner_id":2,"module_id":99}`)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Empty(t, notifier.sent)
}

func TestPlayRejectsClientChosenFields(t *testing.T) {
	notifier := &recordingNotifier{online: true}
	r := commandRouter(&mockService{}, notifier)

	// The deadline and source are the server's to set.
	for _, body := range []string{
		`{"carepartner_id":2,"module_id":3,"expires_at":"2099-01-01T00:00:00Z"}`,
		`{"carepartner_id":2,"module_id":3,"source":"tap"}`,
		`not json`,
	} {
		require.Equal(t, http.StatusBadRequest, post(r, "/api/playbacks/play", body).Code, body)
	}
	require.Empty(t, notifier.sent)
}

// stopWith builds a service whose RequestStop returns the given playback status
// and outcome, checking it was asked about playback 9 for carepartner 2.
func stopWith(t *testing.T, status string, outcome service.StopOutcome) *mockService {
	return &mockService{requestStop: func(_ context.Context, id, cp int) (*model.Playback, service.StopOutcome, error) {
		require.Equal(t, 9, id)
		require.Equal(t, 2, cp)
		return &model.Playback{ID: id, CarepartnerID: cp, Status: status}, outcome, nil
	}}
}

var stopMessage = sentMessage{2, realtime.Message{Type: realtime.TypeStop, PlaybackID: 9}}

func TestStopCancelsCommandTabletHasNotAccepted(t *testing.T) {
	notifier := &recordingNotifier{online: true}
	rec := post(commandRouter(stopWith(t, "stopped", service.StopCancelled), notifier), "/api/playbacks/9/stop", `{"carepartner_id":2}`)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	resp := decode(t, rec)
	require.Equal(t, "stopped", resp.Playback.Status)
	require.True(t, resp.Delivered)
	// Sent anyway, in case the command reached the tablet before its acknowledgement did.
	require.Equal(t, []sentMessage{stopMessage}, notifier.sent)
}

func TestStopOfPlayingClipWaitsForTablet(t *testing.T) {
	notifier := &recordingNotifier{online: true}
	rec := post(commandRouter(stopWith(t, "started", service.StopForwarded), notifier), "/api/playbacks/9/stop", `{"carepartner_id":2}`)

	// Accepted, not done: the record stays "started" until the tablet reports "stopped".
	require.Equal(t, http.StatusAccepted, rec.Code, rec.Body.String())
	resp := decode(t, rec)
	require.Equal(t, "started", resp.Playback.Status)
	require.True(t, resp.Delivered)
	require.Equal(t, []sentMessage{stopMessage}, notifier.sent)
}

func TestStopOfPlayingClipOnDisconnectedTabletChangesNothing(t *testing.T) {
	notifier := &recordingNotifier{online: false}
	rec := post(commandRouter(stopWith(t, "started", service.StopForwarded), notifier), "/api/playbacks/9/stop", `{"carepartner_id":2}`)

	// The clip may still be playing, so the record must not say stopped.
	require.Equal(t, http.StatusConflict, rec.Code)
	require.JSONEq(t, `{"error":"tablet is not connected"}`, rec.Body.String())
}

func TestRepeatedStopSendsNothing(t *testing.T) {
	notifier := &recordingNotifier{online: true}
	rec := post(commandRouter(stopWith(t, "stopped", service.StopAlreadyStopped), notifier), "/api/playbacks/9/stop", `{"carepartner_id":2}`)

	require.Equal(t, http.StatusOK, rec.Code)
	require.False(t, decode(t, rec).Delivered)
	require.Empty(t, notifier.sent)
}

func TestStopOfFinishedPlaybackSendsNothing(t *testing.T) {
	svc := &mockService{requestStop: func(context.Context, int, int) (*model.Playback, service.StopOutcome, error) {
		return nil, 0, service.ErrInvalidTransition
	}}
	notifier := &recordingNotifier{online: true}

	rec := post(commandRouter(svc, notifier), "/api/playbacks/9/stop", `{"carepartner_id":2}`)

	require.Equal(t, http.StatusConflict, rec.Code)
	require.Empty(t, notifier.sent)
}

func TestStopRejectsBadPlaybackID(t *testing.T) {
	r := commandRouter(&mockService{}, &recordingNotifier{})
	for _, id := range []string{"0", "-1", "abc"} {
		require.Equal(t, http.StatusBadRequest, post(r, "/api/playbacks/"+id+"/stop", `{"carepartner_id":2}`).Code, id)
	}
}
