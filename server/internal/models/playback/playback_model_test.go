package playback

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRatingClosesAt(t *testing.T) {
	started := time.Date(2026, 10, 28, 14, 0, 0, 0, time.UTC)
	ended := started.Add(time.Minute)

	require.Nil(t, Playback{}.RatingClosesAt(), "a play that never reached the screen can't be rated")
	require.Equal(t, ended.Add(5*time.Minute), *Playback{StartedAt: &started, EndedAt: &ended}.RatingClosesAt())
	require.Equal(t, started.Add(15*time.Minute), *Playback{StartedAt: &started}.RatingClosesAt(),
		"a play the tablet never reported ending: the longest clip plus the window")
}

func TestPlaybackJSONHasRatingClosesAt(t *testing.T) {
	started := time.Date(2026, 10, 28, 14, 0, 0, 0, time.UTC)
	ended := started.Add(time.Minute)
	rating := RatingHelped

	body, err := json.Marshal(Playback{ID: 4, Status: StatusFinished, StartedAt: &started, EndedAt: &ended, Rating: &rating})
	require.NoError(t, err)
	require.Contains(t, string(body), `"rating_closes_at":"2026-10-28T14:06:00Z"`)
	// The usual fields are still there
	require.Contains(t, string(body), `"id":4`)
	require.Contains(t, string(body), `"status":"finished"`)
	require.Contains(t, string(body), `"rating":1`)

	body, err = json.Marshal(Playback{ID: 5, Status: StatusExpired})
	require.NoError(t, err)
	require.Contains(t, string(body), `"rating_closes_at":null`)
}
