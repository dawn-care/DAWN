package scheduled_items

import (
	"time"

	model "server/internal/models/scheduled_items"
)

// DateLayout is how dates are written in requests and responses
const DateLayout = "2006-01-02"

// MaxCalendarDays caps how many days one calendar request can expand
const MaxCalendarDays = 62

// parseDate reads a "YYYY-MM-DD" date as midnight UTC, so day arithmetic never meets DST
func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation(DateLayout, s, time.UTC)
}

/*
OccursOn reports whether an item repeating from start occurs on date.
once: only the start date. daily: every day. weekly: the start date's weekday.
monthly: the start date's day of the month; months without that day are skipped,
so an item starting on the 31st doesn't occur in April.
The item never occurs before start or after end (if set).
*/
func OccursOn(start time.Time, repeat string, end *time.Time, date time.Time) bool {
	if date.Before(start) || (end != nil && date.After(*end)) {
		return false
	}
	switch repeat {
	case model.RepeatOnce:
		return date.Equal(start)
	case model.RepeatDaily:
		return true
	case model.RepeatWeekly:
		return date.Weekday() == start.Weekday()
	case model.RepeatMonthly:
		return date.Day() == start.Day()
	}
	return false
}

// OccurrenceDates lists the dates in [from, to] an item occurs on, leaving out skipped dates
func OccurrenceDates(start time.Time, repeat string, end *time.Time, skips []string, from, to time.Time) []string {
	skipped := map[string]bool{}
	for _, d := range skips {
		skipped[d] = true
	}
	dates := []string{}
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		day := d.Format(DateLayout)
		if !skipped[day] && OccursOn(start, repeat, end, d) {
			dates = append(dates, day)
		}
	}
	return dates
}
