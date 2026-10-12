package screen_widgets

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func normalize(t *testing.T, widgetType, config string) (string, error) {
	t.Helper()
	out, err := NormalizeConfig(widgetType, json.RawMessage(config))
	return string(out), err
}

func TestNormalizeConfig_Valid(t *testing.T) {
	cases := []struct {
		name, widgetType, in, want string
	}{
		{"photo, no config", TypePhoto, ``, `{}`},
		{"photo caption trimmed", TypePhoto, `{"caption":"  Sarah "}`, `{"caption":"Sarah"}`},
		{"message", TypeMessage, `{"text":"Ana","label":"Nurse"}`, `{"text":"Ana","label":"Nurse"}`},
		{"location reassurance", TypeMessage,
			`{"text":"You are safe at home","location":"Rochester, NY","about":"carepartner"}`,
			`{"text":"You are safe at home","location":"Rochester, NY","about":"carepartner"}`},
		{"reminders sorted, padded, default icon", TypeReminder,
			`{"items":[{"time":"17:30","text":"Dinner","icon":"meal"},{"time":"9:00","text":"Pills"}]}`,
			`{"items":[{"time":"09:00","text":"Pills","icon":"other"},{"time":"17:30","text":"Dinner","icon":"meal"}]}`},
		{"weather default zip", TypeWeather, `{}`, `{"zip":"14617"}`},
		{"weather null config", TypeWeather, `null`, `{"zip":"14617"}`},
		{"weather zip", TypeWeather, `{"zip":" 10001 "}`, `{"zip":"10001"}`},
		{"sports normalized case", TypeSports, `{"league":"MLB","team":"nyy"}`, `{"league":"mlb","team":"NYY"}`},
		{"news", TypeNews, `{}`, `{}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalize(t, tc.widgetType, tc.in)
			require.NoError(t, err)
			assert.JSONEq(t, tc.want, got)
		})
	}
}

func TestNormalizeConfig_Invalid(t *testing.T) {
	tooMany := `{"items":[` + strings.TrimSuffix(strings.Repeat(`{"time":"10:00","text":"x"},`, MaxReminderItems+1), ",") + `]}`
	cases := []struct {
		name, widgetType, in string
	}{
		{"unknown field (typo)", TypePhoto, `{"captoin":"Sarah"}`},
		{"not an object", TypeWeather, `["14617"]`},
		{"caption too long", TypePhoto, `{"caption":"` + strings.Repeat("a", 61) + `"}`},
		{"message without text", TypeMessage, `{"label":"Nurse"}`},
		{"message blank text", TypeMessage, `{"text":"   "}`},
		{"message bad about", TypeMessage, `{"text":"hi","about":"nurse"}`},
		{"label too long", TypeMessage, `{"text":"hi","label":"` + strings.Repeat("a", 41) + `"}`},
		{"no reminders", TypeReminder, `{"items":[]}`},
		{"too many reminders", TypeReminder, tooMany},
		{"bad reminder time", TypeReminder, `{"items":[{"time":"noon","text":"Lunch"}]}`},
		{"reminder without text", TypeReminder, `{"items":[{"time":"12:00","text":""}]}`},
		{"unknown icon", TypeReminder, `{"items":[{"time":"12:00","text":"Lunch","icon":"pizza"}]}`},
		{"bad zip", TypeWeather, `{"zip":"1461"}`},
		{"zip+4", TypeWeather, `{"zip":"14617-1234"}`},
		{"sports missing team", TypeSports, `{"league":"mlb"}`},
		{"unknown team", TypeSports, `{"league":"mlb","team":"XYZ"}`},
		{"team in the wrong league", TypeSports, `{"league":"nfl","team":"NYY"}`},
		{"unknown league", TypeSports, `{"league":"mls","team":"NYC"}`},
		{"news has no settings", TypeNews, `{"category":"sports"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := normalize(t, tc.widgetType, tc.in)
			assert.ErrorIs(t, err, ErrInvalidConfig)
		})
	}

	_, err := normalize(t, "clock", `{}`)
	assert.ErrorIs(t, err, ErrInvalidType)
}
