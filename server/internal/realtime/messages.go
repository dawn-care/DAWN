package realtime

import "time"

// Message types sent from the server to a tablet.
const (
	// TypePlay tells the tablet to play a media item. It carries the playback ID the
	// tablet must report against, and the deadline after which it must not start.
	TypePlay = "play"
	// TypeStop tells the tablet to stop a playback and return to its idle screen.
	TypeStop = "stop"
	// TypeRefresh tells the tablet to re-fetch its schedule and content.
	TypeRefresh = "refresh"
	// TypeStatusResult answers a TypeStatus report with the status actually stored,
	// which can differ from the one reported (an acknowledgement after the deadline
	// is stored as "expired").
	TypeStatusResult = "status_result"
	// TypeError answers a message the server could not understand.
	TypeError = "error"
)

// Message types sent from a tablet to the server.
const (
	// TypeStatus reports a playback status change: acknowledged, started,
	// finished, stopped, failed or expired.
	TypeStatus = "status"
)

// Message is the single JSON envelope used in both directions. Fields that do not
// apply to a given type are omitted.
//
// Server to tablet:
//
//	{"type":"play","playback_id":12,"media_id":3,"expires_at":"2026-10-12T20:00:30Z"}
//	{"type":"stop","playback_id":12}
//	{"type":"refresh"}
//	{"type":"status_result","playback_id":12,"status":"acknowledged"}
//	{"type":"status_result","playback_id":12,"error":"invalid playback status transition"}
//
// Tablet to server:
//
//	{"type":"status","playback_id":12,"status":"acknowledged"}
type Message struct {
	Type       string     `json:"type"`
	PlaybackID int        `json:"playback_id,omitempty"`
	MediaID    int        `json:"media_id,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Status     string     `json:"status,omitempty"`
	Error      string     `json:"error,omitempty"`
}

// PlayCommand builds the play command for a playback. Both the play endpoint and
// the reconnect resend use it, so a tablet sees the same message either way.
func PlayCommand(playbackID, mediaID int, expiresAt time.Time) Message {
	return Message{Type: TypePlay, PlaybackID: playbackID, MediaID: mediaID, ExpiresAt: &expiresAt}
}
