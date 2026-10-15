// Package realtime keeps one WebSocket connection open per paired tablet so the
// server can push commands (play, stop, refresh) to it.
//
// Connections live in this process's memory. That is fine for the single server
// container this project runs; running several server instances would need the
// connections shared through something like Redis pub/sub.
package realtime

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	// DefaultPingInterval keeps idle connections open behind load balancers and
	// proxies, most of which drop connections idle for 60 seconds.
	DefaultPingInterval = 25 * time.Second
	// DefaultPingTimeout is how long a tablet has to answer a ping.
	DefaultPingTimeout = 10 * time.Second

	writeTimeout = 10 * time.Second
	sendBuffer   = 16
)

// Handler processes one message received from a tablet and returns any replies
// to send back to it.
type Handler func(ctx context.Context, carepartnerID int, msg Message) []Message

// Hub tracks the open connection for each carepartner's tablet.
type Hub struct {
	// PingInterval is how often each connection is pinged. A connection that
	// fails to answer within PingTimeout is closed, so a tablet that lost power
	// or network does not appear connected.
	PingInterval time.Duration
	PingTimeout  time.Duration

	mu      sync.Mutex
	clients map[int]*client
}

type client struct {
	send  chan Message
	since time.Time
	// cancel ends this connection, for example when the tablet reconnects and a
	// newer connection replaces it.
	cancel context.CancelFunc
}

func NewHub() *Hub {
	return &Hub{PingInterval: DefaultPingInterval, PingTimeout: DefaultPingTimeout, clients: make(map[int]*client)}
}

// Serve runs one tablet connection until it closes, blocking until then.
// `initial` is sent as soon as the connection is registered — used to resend a
// command the tablet missed while it was disconnected.
func (h *Hub) Serve(ctx context.Context, conn *websocket.Conn, carepartnerID int, initial []Message, handle Handler) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := &client{send: make(chan Message, sendBuffer), since: time.Now().UTC(), cancel: cancel}
	h.register(carepartnerID, c)
	defer h.unregister(carepartnerID, c)
	log.Printf("realtime: tablet connected (carepartner %d)", carepartnerID)

	for _, msg := range initial {
		c.send <- msg
	}
	go writeLoop(ctx, conn, c, cancel)
	go h.pingLoop(ctx, conn, cancel)

	for {
		var msg Message
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			break
		}
		for _, reply := range handle(ctx, carepartnerID, msg) {
			enqueue(c, reply)
		}
	}
	conn.Close(websocket.StatusNormalClosure, "")
	log.Printf("realtime: tablet disconnected (carepartner %d)", carepartnerID)
}

// Send queues a message for a carepartner's tablet. It returns false when the
// tablet is not connected (or its queue is full), so callers can tell whether a
// command was delivered; undelivered commands are resent on reconnect.
func (h *Hub) Send(carepartnerID int, msg Message) bool {
	h.mu.Lock()
	c := h.clients[carepartnerID]
	h.mu.Unlock()
	if c == nil {
		return false
	}
	return enqueue(c, msg)
}

// Connected reports whether a carepartner's tablet has an open connection, and
// since when.
func (h *Hub) Connected(carepartnerID int) (bool, time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c := h.clients[carepartnerID]; c != nil {
		return true, c.since
	}
	return false, time.Time{}
}

// register replaces any existing connection for the carepartner. A tablet that
// reconnects after a network drop often has a stale connection the server has
// not noticed yet; the newer one wins.
func (h *Hub) register(carepartnerID int, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old := h.clients[carepartnerID]; old != nil {
		old.cancel()
	}
	h.clients[carepartnerID] = c
}

// unregister removes the connection only if it is still the current one, so a
// replaced connection closing late does not remove its replacement.
func (h *Hub) unregister(carepartnerID int, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[carepartnerID] == c {
		delete(h.clients, carepartnerID)
	}
}

func enqueue(c *client, msg Message) bool {
	select {
	case c.send <- msg:
		return true
	default:
		return false
	}
}

// writeLoop is the only goroutine that writes to the connection.
func writeLoop(ctx context.Context, conn *websocket.Conn, c *client, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.send:
			writeCtx, done := context.WithTimeout(ctx, writeTimeout)
			err := wsjson.Write(writeCtx, conn, msg)
			done()
			if err != nil {
				cancel()
				return
			}
		}
	}
}

func (h *Hub) pingLoop(ctx context.Context, conn *websocket.Conn, cancel context.CancelFunc) {
	ticker := time.NewTicker(h.PingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, done := context.WithTimeout(ctx, h.PingTimeout)
			err := conn.Ping(pingCtx)
			done()
			if err != nil {
				cancel()
				return
			}
		}
	}
}
