package realtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/require"
)

// serveHub exposes a hub over a real HTTP server. The carepartner ID comes from
// the query string, and `initial` and `handle` are passed straight to Serve.
func serveHub(t *testing.T, h *Hub, initial []Message, handle Handler) *httptest.Server {
	t.Helper()
	if handle == nil {
		handle = func(context.Context, int, Message) []Message { return nil }
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
		require.NoError(t, err)
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		h.Serve(r.Context(), conn, id, initial, handle)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func dial(t *testing.T, srv *httptest.Server, carepartnerID int) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/?carepartner_id=" + strconv.Itoa(carepartnerID)
	conn, _, err := websocket.Dial(ctx, url, nil)
	require.NoError(t, err)
	t.Cleanup(func() { conn.CloseNow() })
	return conn
}

func read(t *testing.T, conn *websocket.Conn) Message {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var msg Message
	require.NoError(t, wsjson.Read(ctx, conn, &msg))
	return msg
}

// waitConnected polls because registration happens on the server's goroutine.
func waitConnected(t *testing.T, h *Hub, carepartnerID int, want bool) {
	t.Helper()
	require.Eventually(t, func() bool {
		connected, _ := h.Connected(carepartnerID)
		return connected == want
	}, 5*time.Second, 10*time.Millisecond)
}

func TestSendReachesConnectedTablet(t *testing.T) {
	h := NewHub()
	srv := serveHub(t, h, nil, nil)
	conn := dial(t, srv, 7)
	waitConnected(t, h, 7, true)

	expires := time.Now().UTC().Add(30 * time.Second).Truncate(time.Second)
	require.True(t, h.Send(7, PlayCommand(12, 3, expires)))

	msg := read(t, conn)
	require.Equal(t, TypePlay, msg.Type)
	require.Equal(t, 12, msg.PlaybackID)
	require.Equal(t, 3, msg.MediaID)
	require.True(t, expires.Equal(*msg.ExpiresAt))
}

func TestSendToDisconnectedTabletReportsUndelivered(t *testing.T) {
	h := NewHub()
	require.False(t, h.Send(99, Message{Type: TypeRefresh}))
	connected, _ := h.Connected(99)
	require.False(t, connected)
}

func TestInitialMessagesSentOnConnect(t *testing.T) {
	h := NewHub()
	srv := serveHub(t, h, []Message{{Type: TypeStop, PlaybackID: 5}}, nil)
	conn := dial(t, srv, 1)

	msg := read(t, conn)
	require.Equal(t, TypeStop, msg.Type)
	require.Equal(t, 5, msg.PlaybackID)
}

func TestHandlerRepliesGoBackToTablet(t *testing.T) {
	h := NewHub()
	var gotCarepartner int
	srv := serveHub(t, h, nil, func(_ context.Context, carepartnerID int, msg Message) []Message {
		gotCarepartner = carepartnerID
		return []Message{{Type: TypeStatusResult, PlaybackID: msg.PlaybackID, Status: msg.Status}}
	})
	conn := dial(t, srv, 4)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, wsjson.Write(ctx, conn, Message{Type: TypeStatus, PlaybackID: 8, Status: "acknowledged"}))

	reply := read(t, conn)
	require.Equal(t, TypeStatusResult, reply.Type)
	require.Equal(t, 8, reply.PlaybackID)
	require.Equal(t, "acknowledged", reply.Status)
	require.Equal(t, 4, gotCarepartner, "the handler is told which tablet sent the message")
}

func TestReconnectReplacesOldConnection(t *testing.T) {
	h := NewHub()
	srv := serveHub(t, h, nil, nil)
	first := dial(t, srv, 2)
	waitConnected(t, h, 2, true)
	second := dial(t, srv, 2)

	// The replaced connection is closed by the server.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := first.Read(ctx)
	require.Error(t, err)

	// Commands go to the newer connection.
	require.True(t, h.Send(2, Message{Type: TypeRefresh}))
	require.Equal(t, TypeRefresh, read(t, second).Type)
}

func TestDisconnectUnregisters(t *testing.T) {
	h := NewHub()
	srv := serveHub(t, h, nil, nil)
	conn := dial(t, srv, 3)
	waitConnected(t, h, 3, true)

	require.NoError(t, conn.Close(websocket.StatusNormalClosure, ""))
	waitConnected(t, h, 3, false)
	require.False(t, h.Send(3, Message{Type: TypeRefresh}))
}

func TestHeartbeatKeepsAnsweredConnectionOpen(t *testing.T) {
	h := NewHub()
	h.PingInterval = 20 * time.Millisecond
	srv := serveHub(t, h, nil, nil)
	conn := dial(t, srv, 6)
	waitConnected(t, h, 6, true)

	// A client answers pings only while it is reading, as a real tablet does.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			if _, _, err := conn.Read(ctx); err != nil {
				return
			}
		}
	}()

	// Several ping intervals later the connection is still registered.
	time.Sleep(10 * h.PingInterval)
	connected, _ := h.Connected(6)
	require.True(t, connected)
}

func TestUnansweredHeartbeatClosesConnection(t *testing.T) {
	h := NewHub()
	h.PingInterval = 20 * time.Millisecond
	h.PingTimeout = 50 * time.Millisecond
	srv := serveHub(t, h, nil, nil)
	dial(t, srv, 8) // never reads, so never answers a ping — like a tablet that lost power
	waitConnected(t, h, 8, true)

	waitConnected(t, h, 8, false)
}
