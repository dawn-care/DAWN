package weather

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWeatherForZip_NetworkErrorsRedactAPIKey(t *testing.T) {
	for _, tc := range []struct {
		name    string
		path    string
		message string
	}{
		{"geocode", "/geo/1.0/zip", "failed geocode request"},
		{"current weather", "/data/2.5/weather", "failed current weather request"},
		{"overview", "/data/3.0/onecall/overview", "failed overview request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstream, _ := fakeOpenWeather(t, http.StatusOK)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == tc.path {
					// Close the connection without a response to produce a real network error.
					conn, _, err := w.(http.Hijacker).Hijack()
					if !assert.NoError(t, err) {
						return
					}
					conn.Close()
					return
				}
				upstream.Config.Handler.ServeHTTP(w, r)
			}))
			t.Cleanup(srv.Close)

			var logs bytes.Buffer
			previousOutput := log.Writer()
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(previousOutput) })

			weather, err := newTestWeatherService(srv.URL).GetWeatherForZip("14617")
			if tc.name == "overview" {
				// An optional overview failure is logged without breaking current weather.
				require.NoError(t, err)
				require.NotNil(t, weather)
				assert.Equal(t, 48, weather.Temp)
				assert.Empty(t, weather.Message)
				assert.Contains(t, logs.String(), tc.message)
			} else {
				require.ErrorContains(t, err, tc.message)
				assert.ErrorIs(t, err, io.EOF)
				assert.NotContains(t, err.Error(), "test-key")
				assert.NotContains(t, err.Error(), "appid=")
			}
			assert.NotContains(t, logs.String(), "test-key")
			assert.NotContains(t, logs.String(), "appid=")
		})
	}
}

// fakeOpenWeather imitates the three OpenWeather endpoints GetWeatherForZip calls
func fakeOpenWeather(t *testing.T, overviewStatus int) (*httptest.Server, *[]string) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path+"?"+r.URL.RawQuery)
		assert.Equal(t, "test-key", r.URL.Query().Get("appid"))
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/geo/1.0/zip":
			if r.URL.Query().Get("zip") != "14617,US" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Write([]byte(`{"lat":43.22,"lon":-77.59,"name":"Rochester","zip":"14617"}`))
		case "/data/2.5/weather":
			assert.Equal(t, "imperial", r.URL.Query().Get("units"))
			w.Write([]byte(`{"main":{"temp":47.6,"feels_like":44.2,"humidity":71,"temp_min":45.1,"temp_max":50.9},
				"weather":[{"main":"Clouds","description":"overcast clouds"}],"wind":{"speed":8.4},"name":"Irondequoit"}`))
		case "/data/3.0/onecall/overview":
			w.WriteHeader(overviewStatus)
			w.Write([]byte(`{"weather_overview":"A cool, cloudy day."}`))
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func newTestWeatherService(baseURL string) *WeatherService {
	s := NewWeatherService("test-key")
	s.baseURL = baseURL
	return s
}

func TestGetWeatherForZip(t *testing.T) {
	srv, calls := fakeOpenWeather(t, http.StatusOK)

	w, err := newTestWeatherService(srv.URL).GetWeatherForZip("14617")

	require.NoError(t, err)
	assert.Equal(t, &WeatherOverviewResponse{
		Location: "Irondequoit", Temp: 48, FeelsLike: 44, TempMin: 45, TempMax: 51,
		Humidity: 71, WindSpeed: 8, Condition: "Clouds", Description: "overcast clouds",
		Message: "A cool, cloudy day.",
	}, w)
	assert.Len(t, *calls, 3)
}

func TestGetWeatherForZip_OverviewIsOptional(t *testing.T) {
	// The overview needs a One Call 3.0 subscription; without it the widget still works
	srv, _ := fakeOpenWeather(t, http.StatusUnauthorized)

	w, err := newTestWeatherService(srv.URL).GetWeatherForZip("14617")

	require.NoError(t, err)
	assert.Equal(t, 48, w.Temp)
	assert.Equal(t, "", w.Message)
}

func TestGetWeatherForZip_Errors(t *testing.T) {
	srv, _ := fakeOpenWeather(t, http.StatusOK)
	s := newTestWeatherService(srv.URL)

	_, err := s.GetWeatherForZip("")
	assert.Error(t, err)

	_, err = s.GetWeatherForZip("99999")
	assert.ErrorContains(t, err, "geocode API returned status 404")

	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer down.Close()
	_, err = newTestWeatherService(down.URL).GetWeatherForZip("14617")
	assert.Error(t, err)
}
