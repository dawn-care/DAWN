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
	widgetModel "server/internal/models/screen_widgets"
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
	widgetCols   = []string{"id", "carepartner_id", "type", "col_start", "row_start", "col_end", "row_end", "config", "photo_media_id"}
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

func expectBasics(mock sqlmock.Sqlmock, rows *sqlmock.Rows) {
	mock.ExpectQuery(qCarepartner).WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "caregiver_id", "timezone"}).AddRow(1, "John", 1, "America/New_York"))
	mock.ExpectQuery(qSettings).WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "wake_time", "bed_time"}).AddRow(1, "07:00", "21:00"))
	mock.ExpectQuery(qWidgets).WithArgs(1).WillReturnRows(rows)
}

/*
mockupWidgets fills the 4x3 grid, in reading order:

	photo   photo   message message
	photo   photo   weather  (empty)
	news    news    news    news
*/
func mockupWidgets() *sqlmock.Rows {
	return sqlmock.NewRows(widgetCols).
		AddRow(1, 1, "photo", 0, 0, 1, 1, []byte(`{}`), 4).
		AddRow(2, 1, "message", 2, 0, 3, 0, []byte(`{"text":"Ana","label":"Nurse"}`), nil).
		AddRow(5, 1, "weather", 2, 1, 2, 1, []byte(`{"zip":"14617"}`), nil).
		AddRow(3, 1, "news", 0, 2, 3, 2, []byte(`{}`), nil)
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

	require.Len(t, screen.Widgets, 4, "every stored widget, in reading order")
	photo, message, weatherTile, news := screen.Widgets[0], screen.Widgets[1], screen.Widgets[2], screen.Widgets[3]
	assert.Equal(t, widgetModel.GridArea{Col: 0, Row: 0, Width: 2, Height: 2}, areaOf(photo))
	assert.Equal(t, PhotoData{Caption: "Sarah"}, photo.Data, "caption falls back to the photo's name")
	assert.Equal(t, "/media/4/file", photo.Photo.URL)
	assert.Equal(t, widgetModel.GridArea{Col: 2, Row: 0, Width: 2, Height: 1}, areaOf(message))
	assert.Equal(t, "Nurse", jsonField(t, message.Config, "label"))
	assert.Nil(t, message.Data, "messages are all config")
	assert.Equal(t, widgetModel.GridArea{Col: 2, Row: 1, Width: 1, Height: 1}, areaOf(weatherTile))
	assert.Equal(t, 48, weatherTile.Data.(*weather.WeatherOverviewResponse).Temp)
	assert.Equal(t, widgetModel.GridArea{Col: 0, Row: 2, Width: 4, Height: 1}, areaOf(news))
	assert.NoError(t, mock.ExpectationsWereMet())

	// The JSON a tablet receives: each widget's area in the shape of GridArea in shared/
	raw, err := json.Marshal(screen)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"widgets":[{"id":1,"type":"photo","col":0,"row":0,"width":2,"height":2,`)
	assert.NotContains(t, string(raw), `"spots"`)
}

func areaOf(w *Widget) widgetModel.GridArea {
	return widgetModel.GridArea{Col: w.Col, Row: w.Row, Width: w.Width, Height: w.Height}
}

func jsonField(t *testing.T, raw json.RawMessage, key string) any {
	m := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &m))
	return m[key]
}

func TestGetScreen_SportsAndNews(t *testing.T) {
	s, mock, _, _ := newService(t, afternoon)
	expectBasics(mock, sqlmock.NewRows(widgetCols).
		AddRow(3, 1, "photo", 0, 0, 1, 1, []byte(`{"caption":"Grandkids"}`), 4).
		AddRow(1, 1, "sports", 2, 0, 2, 0, []byte(`{"league":"mlb","team":"NYY"}`), nil).
		AddRow(2, 1, "news", 3, 0, 3, 1, []byte(`{}`), nil).
		AddRow(4, 1, "reminder", 0, 2, 0, 2, []byte(`{"items":[{"time":"12:00","text":"Lunch","icon":"meal"}]}`), nil))
	expectPhoto(mock)

	screen, err := s.GetScreen(1)

	require.NoError(t, err)
	assert.Equal(t, PhotoData{Caption: "Grandkids"}, screen.Widgets[0].Data, "caption wins over the photo's name")
	assert.Equal(t, SportsData{League: "mlb", TeamName: "New York Yankees", Result: "won",
		OpponentTeamName: "Boston Red Sox", TeamScore: 5, OpponentScore: 3, Message: "won"}, screen.Widgets[1].Data)
	assert.Len(t, screen.Widgets[2].Data.(NewsData).Articles, MaxNewsArticles)
	assert.Equal(t, widgetModel.GridArea{Col: 3, Row: 0, Width: 1, Height: 2}, areaOf(screen.Widgets[2]))
	assert.Equal(t, "reminder", screen.Widgets[3].Type)
	assert.Nil(t, screen.Widgets[3].Data, "reminders are all config")
}

func TestGetScreen_UpstreamFailureOnlyAffectsThatWidget(t *testing.T) {
	s, mock, w, _ := newService(t, afternoon)
	w.err = errors.New("OpenWeather is down")
	expectBasics(mock, mockupWidgets())
	expectPhoto(mock)

	screen, err := s.GetScreen(1)

	require.NoError(t, err)
	weatherTile := screen.Widgets[2]
	assert.Nil(t, weatherTile.Data)
	require.NotNil(t, weatherTile.Error)
	assert.Equal(t, ErrorUpstreamUnavailable, *weatherTile.Error)
	assert.NotNil(t, screen.Widgets[0].Data, "other widgets still filled in")
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
	assert.NotNil(t, screen.Widgets[2].Data, "recovers on the next request")
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
	assert.Empty(t, screen.Widgets, "a new carepartner has an empty screen")
	raw, err := json.Marshal(screen)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"widgets":[]`, "an empty list, not null")
}

func TestGetScreen_Errors(t *testing.T) {
	s, mock, _, _ := newService(t, afternoon)
	_, err := s.GetScreen(0)
	assert.ErrorIs(t, err, ErrInvalidCarepartner)

	mock.ExpectQuery(qCarepartner).WithArgs(9).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "caregiver_id", "timezone"}))
	_, err = s.GetScreen(9)
	assert.ErrorIs(t, err, ErrCarepartnerMissing)
}
