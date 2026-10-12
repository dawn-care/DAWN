package playback

import "time"

const (
	SourceScheduled = "scheduled"
	SourceStaff     = "staff"
	SourceTap       = "tap"

	StatusRequested    = "requested"
	StatusAcknowledged = "acknowledged"
	StatusStarted      = "started"
	StatusFinished     = "finished"
	StatusStopped      = "stopped"
	StatusFailed       = "failed"
	StatusExpired      = "expired"
)

// Playback records one attempt to play a media item, not the media itself.
// ScheduledItemID is set for scheduled plays and nil for tap and staff plays.
type Playback struct {
	ID              int        `json:"id"`
	CarepartnerID   int        `json:"carepartner_id"`
	MediaID         int        `json:"media_id"`
	ScheduledItemID *int       `json:"scheduled_item_id"`
	Source          string     `json:"source"`
	Status          string     `json:"status"`
	RequestedAt     time.Time  `json:"requested_at"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at"`
	StartedAt       *time.Time `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	Rating          *int       `json:"rating"`
}
