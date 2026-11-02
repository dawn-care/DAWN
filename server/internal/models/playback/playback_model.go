package playback

import (
	"encoding/json"
	"time"
)

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

// Ratings staff give a play on five steps, as stored in Rating (migration 8 allows
// only these)
const (
	RatingHelpedALot = 2
	RatingHelped     = 1
	RatingNoChange   = 0
	RatingWorse      = -1
	RatingMuchWorse  = -2
)

// RatingNames are the words the staff app sends for each rating
var RatingNames = map[string]int{
	"helped_a_lot": RatingHelpedALot,
	"helped":       RatingHelped,
	"no_change":    RatingNoChange,
	"worse":        RatingWorse,
	"much_worse":   RatingMuchWorse,
}

// NegativeWeight is how much more a "worse" rating counts than the matching
// "helped" one, so media that upset someone drops faster than helpful media rises
const NegativeWeight = 1.5

/*
RatingPoints are what each rating adds to its media's score, which orders the staff
app's list (see ListMedia in models/media): the step itself (a step off the middle is
1 point, the ends are 2), with "worse" steps times NegativeWeight.
*/
var RatingPoints = map[int]float64{
	RatingHelpedALot: 2,
	RatingHelped:     1,
	RatingNoChange:   0,
	RatingWorse:      -1 * NegativeWeight,
	RatingMuchWorse:  -2 * NegativeWeight,
}

// RatingHalfLife is how long it takes a rating to count half as much in the score,
// so what helped recently outranks what helped months ago
const RatingHalfLife = 30 * 24 * time.Hour

// RatingWindow is how long after a play ends staff can rate it; later ratings aren't
// accurate
const RatingWindow = 5 * time.Minute

// UnendedRatingLimit is how long after it started a play the tablet never reported
// ending can still be rated: the longest clip (10 minutes) plus RatingWindow
const UnendedRatingLimit = 15 * time.Minute

// Playback records one attempt to play a media item, not the media itself.
// ScheduledItemID is set for scheduled plays and nil for tap and staff plays.
// MediaID is nil once the media is deleted; MediaKind and MediaDisplayName are copied
// from the media when the playback is created, so they still say what played.
type Playback struct {
	ID               int        `json:"id"`
	CarepartnerID    int        `json:"carepartner_id"`
	MediaID          *int       `json:"media_id"`
	MediaKind        string     `json:"media_kind"`
	MediaDisplayName string     `json:"media_display_name"`
	ScheduledItemID  *int       `json:"scheduled_item_id"`
	Source           string     `json:"source"`
	Status           string     `json:"status"`
	RequestedAt      time.Time  `json:"requested_at"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at"`
	StartedAt        *time.Time `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	Rating           *int       `json:"rating"`
}

// RatingClosesAt is when staff can no longer rate this play, or nil if it never
// reached the screen and so can't be rated at all
func (p Playback) RatingClosesAt() *time.Time {
	if p.StartedAt == nil {
		return nil
	}
	closes := p.StartedAt.Add(UnendedRatingLimit)
	if p.EndedAt != nil {
		closes = p.EndedAt.Add(RatingWindow)
	}
	return &closes
}

// MarshalJSON adds rating_closes_at, so the staff app shows the rating buttons only
// while a rating would be accepted
func (p Playback) MarshalJSON() ([]byte, error) {
	type plain Playback
	return json.Marshal(struct {
		plain
		RatingClosesAt *time.Time `json:"rating_closes_at"`
	}{plain(p), p.RatingClosesAt()})
}
