package realtime_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctrl "server/internal/controllers/realtime"
	model "server/internal/models/playback"
	"server/internal/models/user/carepartner"
	hub "server/internal/realtime"
	routes "server/internal/routes/realtime"
	service "server/internal/services/playback"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// mockPlaybacks implements PlaybackServiceInterface.
type mockPlaybacks struct {
	live         *model.Playback
	updateStatus func(ctx context.Context, id, carepartnerID int, status string) (*model.Playback, error)
}

func (m *mockPlaybacks) UpdateStatus(ctx context.Context, id, carepartnerID int, status string) (*model.Playback, error) {
	return m.updateStatus(ctx, id, carepartnerID, status)
}

func (m *mockPlaybacks) LiveCommand(context.Context, int) (*model.Playback, error) {
	return m.live, nil
}

// mockCarepartners implements CarepartnerServiceInterface; only listed IDs exist.
type mockCarepartners map[int]bool

func (m mockCarepartners) GetCarepartnerByID(id int) (*carepartner.Carepartner, error) {
	if m[id] {
		return &carepartner.Carepartner{ID: id}, nil
	}
	return nil, nil
}

// fakeHub implements HubInterface for the plain HTTP endpoints.
type fakeHub struct {
	connected bool
	since     time.Time
	sent      []hub.Message
}

func (f *fakeHub) Serve(context.Context, *websocket.Conn, int, []hub.Message, hub.Handler) {}
func (f *fakeHub) Send(_ int, msg hub.Message) bool {
	f.sent = append(f.sent, msg)
	return f.connected
}
func (f *fakeHub) Connected(int) (bool, time.Time) { return f.connected, f.since }

func router(playbacks *mockPlaybacks, carepartners mockCarepartners, h ctrl.HubInterface) http.Handler {
	r := chi.NewRouter()
	routes.RegisterRealtimeRoutes(r, ctrl.NewRealtimeController(playbacks, carepartners, h, nil))
	return r
}

func do(r http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestConnectionStatus(t *testing.T) {
	since := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	r := router(&mockPlaybacks{}, nil, &fakeHub{connected: true, since: since})

	rec := do(r, http.MethodGet, "/api/realtime/carepartners/3/connection")
	require.Equal(t, http.StatusOK, rec.Code)
	var resp ctrl.ConnectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 3, resp.CarepartnerID)
	require.True(t, resp.Connected)
	require.True(t, since.Equal(*resp.ConnectedSince))

	offline := router(&mockPlaybacks{}, nil, &fakeHub{})
	require.JSONEq(t, `{"carepartner_id":3,"connected":false}`,
		do(offline, http.MethodGet, "/api/realtime/carepartners/3/connection").Body.String())
}

func TestRefreshPushesRefresh(t *testing.T) {
	h := &fakeHub{connected: true}
	rec := do(router(&mockPlaybacks{}, nil, h), http.MethodPost, "/api/realtime/carepartners/3/refresh")

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"delivered":true}`, rec.Body.String())
	require.Equal(t, []hub.Message{{Type: hub.TypeRefresh}}, h.sent)
}

func TestBadCarepartnerIDs(t *testing.T) {
	r := router(&mockPlaybacks{}, mockCarepartners{}, &fakeHub{})
	for _, path := range []string{
		"/api/realtime/carepartners/0/connection",
		"/api/realtime/carepartners/abc/connection",
		"/api/realtime/tablet",
		"/api/realtime/tablet?carepartner_id=-4",
	} {
		require.Equal(t, http.StatusBadRequest, do(r, http.MethodGet, path).Code, path)
	}
}

func TestConnectUnknownCarepartner(t *testing.T) {
	r := router(&mockPlaybacks{}, mockCarepartners{}, hub.NewHub())
	require.Equal(t, http.StatusNotFound, do(r, http.MethodGet, "/api/realtime/tablet?carepartner_id=5").Code)
}

// TestTabletRoundTrip drives the real hub through the real connect endpoint:
// the tablet connects, receives the command it missed, and acknowledges it.
func TestTabletRoundTrip(t *testing.T) {
	expires := time.Now().UTC().Add(20 * time.Second).Truncate(time.Second)
	var reported []string
	playbacks := &mockPlaybacks{
		live: &model.Playback{ID: 11, CarepartnerID: 5, ModuleID: 4, Status: "requested", ExpiresAt: expires},
		updateStatus: func(_ context.Context, id, cp int, status string) (*model.Playback, error) {
			require.Equal(t, 11, id)
			require.Equal(t, 5, cp, "status reports are scoped to the connected tablet")
			reported = append(reported, status)
			if status == "finished" {
				return nil, service.ErrInvalidTransition
			}
			return &model.Playback{ID: id, CarepartnerID: cp, Status: status}, nil
		},
	}
	tabletHub := hub.NewHub()
	srv := httptest.NewServer(router(playbacks, mockCarepartners{5: true}, tabletHub))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/api/realtime/tablet?carepartner_id=5", nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	// Each read decodes into a fresh value: decoding into a reused struct would
	// keep fields from the previous message that the next one omits.
	next := func() hub.Message {
		var msg hub.Message
		require.NoError(t, wsjson.Read(ctx, conn, &msg))
		return msg
	}

	// 1. The missed command arrives first.
	require.Equal(t, hub.PlayCommand(11, 4, expires), next())

	// 2. Acknowledging it goes through the service and is confirmed.
	require.NoError(t, wsjson.Write(ctx, conn, hub.Message{Type: hub.TypeStatus, PlaybackID: 11, Status: "acknowledged"}))
	require.Equal(t, hub.Message{Type: hub.TypeStatusResult, PlaybackID: 11, Status: "acknowledged"}, next())

	// 3. A rejected report comes back as an error without internal detail.
	require.NoError(t, wsjson.Write(ctx, conn, hub.Message{Type: hub.TypeStatus, PlaybackID: 11, Status: "finished"}))
	require.Equal(t, hub.Message{Type: hub.TypeStatusResult, PlaybackID: 11, Error: "invalid playback status transition"}, next())

	// 4. Unknown message types are answered rather than ignored.
	require.NoError(t, wsjson.Write(ctx, conn, hub.Message{Type: "dance"}))
	require.Equal(t, hub.TypeError, next().Type)

	// 5. The connection is visible to the rest of the server.
	connected, _ := tabletHub.Connected(5)
	require.True(t, connected)
	require.Equal(t, []string{"acknowledged", "finished"}, reported)
}
