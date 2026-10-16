package screen

import (
	"encoding/json"
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	scheduledModel "server/internal/models/scheduled_items"
	newsSvc "server/internal/services/widgets/news"
	"server/internal/services/widgets/weather"
	"server/internal/utils"
)

// ---- fakes ----

type fakeWeather struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (f *fakeWeather) GetWeatherForZip(zip string) (*weather.WeatherOverviewResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return &weather.WeatherOverviewResponse{Location: "Irondequoit", Temp: 48, Condition: "Clouds"}, nil
}

type fakeSports struct{}

func (fakeSports) GetTeamGameFor(league, team string) (*newsSvc.GameSummaryResponse, error) {
	return &newsSvc.GameSummaryResponse{UserTeamName: "New York Yankees", OpponentTeamName: "Boston Red Sox",
		UserScore: 5, OpponentScore: 3, Result: newsSvc.ResultWon, Message: "won"}, nil
}

type fakeNews struct{}

func (fakeNews) GetGoodNews() (*newsSvc.NewsAPIResponse, error) {
	r := &newsSvc.NewsAPIResponse{}
	for _, title := range []string{"a", "b", "c", "d"} {
		r.Articles = append(r.Articles, struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			URLToImage  string `json:"urlToImage"`
		}{Title: title})
	}
	return r, nil
}

type fakeCalendar struct{ from, to string }

func (f *fakeCalendar) Calendar(carepartnerID int, from, to string) ([]scheduledModel.Occurrence, error) {
	f.from, f.to = from, to
	return []scheduledModel.Occurrence{{ScheduledItemID: 7, Date: from, Time: "10:00", Shape: "video"}}, nil
}

// ---- helpers ----

var (
	qCarepartner = regexp.QuoteMeta(`FROM "Carepartner"`)
	qSettings    = regexp.QuoteMeta(`FROM "ScreenSettings"`)
	qWidgets     = regexp.QuoteMeta(`FROM "ScreenWidget"`)
	qMedia       = regexp.QuoteMeta(`FROM "Media" m`)
	widgetCols   = []string{"id", "carepartner_id", "type", "spot", "config", "photo_media_id"}
	mediaCols    = []string{"id", "carepartner_id", "kind", "display_name", "object_key", "mime_type",
		"size_bytes", "duration_seconds", "created_at", "on_screen", "audio_background", "scheduled"}
)

// 2026-10-12 14:30 in New York
var afternoon = time.Date(2026, 10, 12, 18, 30, 0, 0, time.UTC)

func newService(t *testing.T, now time.Time) (*ScreenService, sqlmock.Sqlmock, *fakeWeather, *fakeCalendar) {
	db, mock := utils.SetupMockDB(t)
	t.Cleanup(func() { db.Close() })
	w, cal := &fakeWeather{}, &fakeCalendar{}
	s := NewScreenService(db, w, fakeSports{}, fakeNews{}, cal)
	s.Now = func() time.Time { return now }
	s.cache = newLookupCache(s.Now)
	return s, mock, w, cal
}

func expectBasicsWith(mock sqlmock.Sqlmock, preset string, rows *sqlmock.Rows) {
	mock.ExpectQuery(qCarepartner).WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "caregiver_id", "timezone"}).AddRow(1, "John", 1, "America/New_York"))
	mock.ExpectQuery(qSettings).WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "layout_preset", "wake_time", "bed_time"}).AddRow(1, preset, "07:00", "21:00"))
	mock.ExpectQuery(qWidgets).WithArgs(1).WillReturnRows(rows)
}

// Balanced has 4 spots, 0 to 3
func expectBasics(mock sqlmock.Sqlmock, rows *sqlmock.Rows) {
	expectBasicsWith(mock, "balanced", rows)
}

func mockupWidgets() *sqlmock.Rows {
	return sqlmock.NewRows(widgetCols).
		AddRow(1, 1, "photo", 0, []byte(`{}`), 4).
		AddRow(2, 1, "message", 1, []byte(`{"text":"Ana","label":"Nurse"}`), nil).
		AddRow(5, 1, "weather", 2, []byte(`{"zip":"14617"}`), nil).
		// spot 3 is empty
		AddRow(3, 1, "news", 4, []byte(`{}`), nil) // hidden: Balanced has no spot 4
}

func expectPhoto(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(qMedia).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(mediaCols).AddRow(4, 1, "photo", "Sarah", "k", "image/png", 1, nil, time.Now(), true, false, false))
}

// ---- tests ----

func TestGetScreen_Mockup(t *testing.T) {
	s, mock, _, cal := newService(t, afternoon)
	expectBasics(mock, mockupWidgets())
	expectPhoto(mock)

	screen, err := s.GetScreen(1)

	require.NoError(t, err)
	assert.Equal(t, "2026-10-12T14:30:00-04:00", screen.GeneratedAt)
	assert.Equal(t, "2026-10-12", screen.Today)
	assert.Equal(t, AwakeHours{Wake: "07:00", Bed: "21:00", IsAwakeNow: true}, screen.AwakeHours)
	assert.Equal(t, "2026-10-12", cal.from)
	assert.Equal(t, "2026-10-13", cal.to, "schedule covers today and tomorrow")
	require.Len(t, screen.Schedule, 1)

	assert.Equal(t, "balanced", screen.LayoutPreset)
	require.Len(t, screen.Spots, 4, "one entry per spot of the preset")
	assert.Equal(t, PhotoData{Caption: "Sarah"}, screen.Spots[0].Data, "caption falls back to the photo's name")
	assert.Equal(t, "/media/4/file", screen.Spots[0].Photo.URL)
	assert.Equal(t, "Nurse", jsonField(t, screen.Spots[1].Config, "label"))
	assert.Nil(t, screen.Spots[1].Data, "messages are all config")
	assert.Equal(t, 48, screen.Spots[2].Data.(*weather.WeatherOverviewResponse).Temp)
	assert.Nil(t, screen.Spots[3], "empty spot")

	// The hidden news widget isn't anywhere
	for _, w := range screen.Spots {
		if w != nil {
			assert.NotEqual(t, 3, w.ID)
		}
	}
	assert.NoError(t, mock.ExpectationsWereMet())

	// The JSON a tablet receives
	raw, err := json.Marshal(screen)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"spots":[{"id":1,"type":"photo"`)
	assert.Contains(t, string(raw), `null]`, "empty spot is null")
	assert.NotContains(t, string(raw), `"regions"`)
}

func jsonField(t *testing.T, raw json.RawMessage, key string) any {
	m := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &m))
	return m[key]
}

func TestGetScreen_SportsAndNews(t *testing.T) {
	s, mock, _, _ := newService(t, afternoon)
	expectBasics(mock, sqlmock.NewRows(widgetCols).
		AddRow(3, 1, "photo", 0, []byte(`{"caption":"Grandkids"}`), 4).
		AddRow(1, 1, "sports", 1, []byte(`{"league":"mlb","team":"NYY"}`), nil).
		AddRow(2, 1, "news", 2, []byte(`{}`), nil).
		AddRow(4, 1, "reminder", 3, []byte(`{"items":[{"time":"12:00","text":"Lunch","icon":"meal"}]}`), nil))
	expectPhoto(mock)

	screen, err := s.GetScreen(1)

	require.NoError(t, err)
	assert.Equal(t, PhotoData{Caption: "Grandkids"}, screen.Spots[0].Data, "caption wins over the photo's name")
	assert.Equal(t, SportsData{League: "mlb", TeamName: "New York Yankees", Result: "won",
		OpponentTeamName: "Boston Red Sox", TeamScore: 5, OpponentScore: 3, Message: "won"}, screen.Spots[1].Data)
	assert.Len(t, screen.Spots[2].Data.(NewsData).Articles, MaxNewsArticles)
	assert.Equal(t, "reminder", screen.Spots[3].Type)
	assert.Nil(t, screen.Spots[3].Data, "reminders are all config")
}

func TestGetScreen_OtherPresetsKeepSpots(t *testing.T) {
	// The same stored widgets on Simple (2 spots) and Busier (6 spots)
	s, mock, _, _ := newService(t, afternoon)
	expectBasicsWith(mock, "simple", mockupWidgets())
	expectPhoto(mock)

	screen, err := s.GetScreen(1)

	require.NoError(t, err)
	assert.Equal(t, "simple", screen.LayoutPreset)
	require.Len(t, screen.Spots, 2)
	assert.Equal(t, 1, screen.Spots[0].ID)
	assert.Equal(t, 2, screen.Spots[1].ID)

	expectBasicsWith(mock, "busier", mockupWidgets())
	expectPhoto(mock)

	screen, err = s.GetScreen(1)

	require.NoError(t, err)
	require.Len(t, screen.Spots, 6)
	assert.Equal(t, 5, screen.Spots[2].ID, "widgets stay in their spot")
	assert.Nil(t, screen.Spots[3])
	assert.Equal(t, 3, screen.Spots[4].ID, "hidden on Balanced, shown on Busier")
	assert.Nil(t, screen.Spots[5])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetScreen_UpstreamFailureOnlyAffectsThatWidget(t *testing.T) {
	s, mock, w, _ := newService(t, afternoon)
	w.err = errors.New("OpenWeather is down")
	expectBasics(mock, mockupWidgets())
	expectPhoto(mock)

	screen, err := s.GetScreen(1)

	require.NoError(t, err)
	weatherTile := screen.Spots[2]
	assert.Nil(t, weatherTile.Data)
	require.NotNil(t, weatherTile.Error)
	assert.Equal(t, ErrorUpstreamUnavailable, *weatherTile.Error)
	assert.NotNil(t, screen.Spots[0].Data, "other widgets still filled in")
}

func TestGetScreen_CachesLookups(t *testing.T) {
	now := afternoon
	s, mock, w, _ := newService(t, now)
	s.Now = func() time.Time { return now }
	s.cache = newLookupCache(func() time.Time { return now })

	for i := 0; i < 2; i++ {
		expectBasics(mock, mockupWidgets())
		expectPhoto(mock)
		_, err := s.GetScreen(1)
		require.NoError(t, err)
	}
	assert.Equal(t, 1, w.calls, "second screen within 10 minutes reuses the weather")

	now = now.Add(LookupCacheTTL + time.Second)
	expectBasics(mock, mockupWidgets())
	expectPhoto(mock)
	_, err := s.GetScreen(1)
	require.NoError(t, err)
	assert.Equal(t, 2, w.calls, "fetched again once the cache expires")
}

func TestGetScreen_FailuresAreNotCached(t *testing.T) {
	s, mock, w, _ := newService(t, afternoon)
	w.err = errors.New("down")
	expectBasics(mock, mockupWidgets())
	expectPhoto(mock)
	_, _ = s.GetScreen(1)

	w.err = nil
	expectBasics(mock, mockupWidgets())
	expectPhoto(mock)
	screen, err := s.GetScreen(1)
	require.NoError(t, err)
	assert.NotNil(t, screen.Spots[2].Data, "recovers on the next request")
	assert.Equal(t, 2, w.calls)
}

func TestGetScreen_UsesCarepartnerTimezone(t *testing.T) {
	// 02:30 UTC on the 13th is 22:30 on the 12th in New York: still the 12th, after bed time
	s, mock, _, cal := newService(t, time.Date(2026, 10, 13, 2, 30, 0, 0, time.UTC))
	expectBasics(mock, sqlmock.NewRows(widgetCols))

	screen, err := s.GetScreen(1)

	require.NoError(t, err)
	assert.Equal(t, "2026-10-12", screen.Today)
	assert.Equal(t, "2026-10-12", cal.from)
	assert.False(t, screen.AwakeHours.IsAwakeNow)
	assert.Equal(t, "2026-10-12T22:30:00-04:00", screen.GeneratedAt)
	require.Len(t, screen.Spots, 4)
	for _, w := range screen.Spots {
		assert.Nil(t, w, "a new carepartner has an empty screen")
	}
}

func TestGetScreen_Errors(t *testing.T) {
	s, mock, _, _ := newService(t, afternoon)
	_, err := s.GetScreen(0)
	assert.ErrorIs(t, err, ErrInvalidCarepartner)

	mock.ExpectQuery(qCarepartner).WithArgs(9).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "caregiver_id", "timezone"}))
	_, err = s.GetScreen(9)
	assert.ErrorIs(t, err, ErrCarepartnerMissing)
}
