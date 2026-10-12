package scheduled_items

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func day(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := parseDate(s)
	require.NoError(t, err)
	return d
}

func TestOccursOn(t *testing.T) {
	end := day(t, "2026-12-31")
	cases := []struct {
		name   string
		start  string
		repeat string
		end    *time.Time
		date   string
		want   bool
	}{
		{"once on its day", "2026-10-12", "once", nil, "2026-10-12", true},
		{"once the next day", "2026-10-12", "once", nil, "2026-10-13", false},
		{"daily", "2026-10-12", "daily", nil, "2027-03-01", true},
		{"nothing before the start", "2026-10-12", "daily", nil, "2026-10-11", false},
		{"daily on the end date", "2026-10-12", "daily", &end, "2026-12-31", true},
		{"daily after the end date", "2026-10-12", "daily", &end, "2027-01-01", false},
		{"weekly, same weekday (Monday)", "2026-10-12", "weekly", nil, "2026-10-26", true},
		{"weekly, other weekday", "2026-10-12", "weekly", nil, "2026-10-27", false},
		{"monthly, same day", "2026-01-15", "monthly", nil, "2026-06-15", true},
		{"monthly, other day", "2026-01-15", "monthly", nil, "2026-06-16", false},
		{"monthly from the 31st, a 31 day month", "2026-01-31", "monthly", nil, "2026-03-31", true},
		{"monthly from the 31st skips April", "2026-01-31", "monthly", nil, "2026-04-30", false},
		{"monthly from the 31st skips February", "2026-01-31", "monthly", nil, "2026-02-28", false},
		{"monthly from the 29th, leap year February", "2027-01-29", "monthly", nil, "2028-02-29", true},
		{"unknown repeat", "2026-10-12", "yearly", nil, "2026-10-12", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, OccursOn(day(t, tc.start), tc.repeat, tc.end, day(t, tc.date)))
		})
	}
}

func TestOccurrenceDates(t *testing.T) {
	from, to := day(t, "2026-10-01"), day(t, "2026-10-31")

	weekly := OccurrenceDates(day(t, "2026-10-05"), "weekly", nil, []string{"2026-10-19"}, from, to)
	assert.Equal(t, []string{"2026-10-05", "2026-10-12", "2026-10-26"}, weekly, "the skipped Monday is left out")

	ending := day(t, "2026-10-03")
	daily := OccurrenceDates(day(t, "2026-09-25"), "daily", &ending, nil, from, to)
	assert.Equal(t, []string{"2026-10-01", "2026-10-02", "2026-10-03"}, daily, "clipped to the range and the end date")

	once := OccurrenceDates(day(t, "2026-11-02"), "once", nil, nil, from, to)
	assert.Equal(t, []string{}, once, "outside the range")

	// Across a daylight saving change (Nov 1, 2026 in the US) days stay whole
	dst := OccurrenceDates(day(t, "2026-10-31"), "daily", nil, nil, day(t, "2026-10-31"), day(t, "2026-11-02"))
	assert.Equal(t, []string{"2026-10-31", "2026-11-01", "2026-11-02"}, dst)
}

func TestOutsideAwakeHours(t *testing.T) {
	cases := []struct {
		name     string
		start    string
		duration int
		want     bool
	}{
		{"mid-morning", "10:00", 600, false},
		{"exactly at wake", "07:00", 300, false},
		{"before wake", "06:45", 60, true},
		{"finishes exactly at bed", "20:50", 600, false},
		{"finishes after bed", "20:45", 901, true},
		{"starts at bed", "21:00", 1, true},
		{"no length, at bed", "21:00", 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, outsideAwakeHours(tc.start, tc.duration, "07:00", "21:00"))
		})
	}
}
