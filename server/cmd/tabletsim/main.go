// Command tabletsim pretends to be a carepartner's tablet, for testing the
// WebSocket hub by hand before the real tablet app supports it.
//
// It connects as one carepartner, prints every message it receives, and plays
// staff commands the way the tablet is expected to: acknowledge, start, then
// finish after a pretend clip length. It reconnects with backoff when the
// connection drops, and ignores commands it has already handled or that have
// expired.
//
//	go run ./cmd/tabletsim -carepartner 1
//	go run ./cmd/tabletsim -carepartner 1 -clip 20s -ack-delay 35s   # acknowledge too late
package main

import (
	"context"
	"flag"
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"

	"server/internal/realtime"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func main() {
	server := flag.String("server", "http://localhost:8080", "API server base URL")
	carepartnerID := flag.Int("carepartner", 1, "carepartner ID to connect as")
	clip := flag.Duration("clip", 10*time.Second, "how long a pretend clip plays before finishing")
	ackDelay := flag.Duration("ack-delay", 0, "wait this long before acknowledging (to test expiry)")
	flag.Parse()

	u, err := url.Parse(*server)
	if err != nil {
		log.Fatalf("bad -server: %v", err)
	}
	u.Scheme = map[string]string{"https": "wss"}[u.Scheme]
	if u.Scheme == "" {
		u.Scheme = "ws"
	}
	u.Path = "/api/realtime/tablet"
	u.RawQuery = "carepartner_id=" + strconv.Itoa(*carepartnerID)

	sim := &tablet{clip: *clip, ackDelay: *ackDelay, handled: map[int]bool{}}
	backoff := time.Second
	for {
		connectedAt := time.Now()
		err := sim.run(u.String())
		if time.Since(connectedAt) > 30*time.Second {
			backoff = time.Second // it was a healthy connection; retry promptly
		}
		log.Printf("disconnected: %v — reconnecting in %s", err, backoff)
		time.Sleep(backoff)
		backoff = min(backoff*2, 30*time.Second)
	}
}

type tablet struct {
	clip, ackDelay time.Duration

	mu      sync.Mutex
	conn    *websocket.Conn
	handled map[int]bool // playback IDs already acted on, kept across reconnects
	playing int          // playback ID currently playing, 0 when idle
	cancel  context.CancelFunc
}

func (t *tablet) run(addr string) error {
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, addr, nil)
	if err != nil {
		return err
	}
	defer conn.CloseNow()
	t.mu.Lock()
	t.conn = conn
	t.mu.Unlock()
	log.Printf("connected to %s", addr)

	for {
		var msg realtime.Message
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			return err
		}
		log.Printf("<- %s", describe(msg))
		switch msg.Type {
		case realtime.TypePlay:
			t.play(msg)
		case realtime.TypeStop:
			t.stop(msg.PlaybackID)
		case realtime.TypeRefresh:
			log.Printf("   (would re-fetch schedule and content now)")
		}
	}
}

func (t *tablet) play(msg realtime.Message) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.handled[msg.PlaybackID] {
		log.Printf("   ignoring playback %d: already handled (resent after reconnect)", msg.PlaybackID)
		return
	}
	t.handled[msg.PlaybackID] = true
	if msg.ExpiresAt != nil && time.Now().After(*msg.ExpiresAt) {
		log.Printf("   ignoring playback %d: expired at %s", msg.PlaybackID, msg.ExpiresAt.Local().Format(time.TimeOnly))
		return
	}
	// A staff play pre-empts whatever is playing. The interrupted playback is
	// reported "stopped", since the server waits for the tablet to say so.
	if t.cancel != nil {
		t.cancel()
		if interrupted := t.playing; interrupted != 0 {
			log.Printf("   ■ interrupting playback %d", interrupted)
			go t.report(interrupted, "stopped") // report locks t.mu, which play holds
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel, t.playing = cancel, msg.PlaybackID

	go func() {
		id := msg.PlaybackID
		if !sleep(ctx, t.ackDelay) {
			return
		}
		t.report(id, "acknowledged")
		t.report(id, "started")
		log.Printf("   ▶ playing media %d for %s", msg.MediaID, t.clip)
		if !sleep(ctx, t.clip) {
			return
		}
		t.report(id, "finished")
		t.mu.Lock()
		if t.playing == id {
			t.playing = 0
		}
		t.mu.Unlock()
	}()
}

// stop ends a playback and reports "stopped". The server records a stop only
// once the tablet reports it, so the record never says stopped while a clip is
// still playing.
func (t *tablet) stop(playbackID int) {
	t.mu.Lock()
	wasPlaying := t.playing == playbackID && t.cancel != nil
	if wasPlaying {
		t.cancel()
		t.playing = 0
	}
	t.mu.Unlock()

	if wasPlaying {
		log.Printf("   ■ stopped, back to the idle screen")
		t.report(playbackID, "stopped")
	}
}

// report sends a status change. The server answers with a status_result, which
// the read loop prints.
func (t *tablet) report(playbackID int, status string) {
	t.mu.Lock()
	conn := t.conn
	t.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	msg := realtime.Message{Type: realtime.TypeStatus, PlaybackID: playbackID, Status: status}
	if err := wsjson.Write(ctx, conn, msg); err != nil {
		log.Printf("   could not report %s: %v", status, err)
		return
	}
	log.Printf("-> %s", describe(msg))
}

func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

func describe(m realtime.Message) string {
	s := m.Type
	if m.PlaybackID != 0 {
		s += " playback=" + strconv.Itoa(m.PlaybackID)
	}
	if m.MediaID != 0 {
		s += " media=" + strconv.Itoa(m.MediaID)
	}
	if m.Status != "" {
		s += " status=" + m.Status
	}
	if m.ExpiresAt != nil {
		s += " expires=" + m.ExpiresAt.Local().Format(time.TimeOnly)
	}
	if m.Error != "" {
		s += " error=" + strconv.Quote(m.Error)
	}
	return s
}
