package weather

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetWeatherForZip_Live(t *testing.T) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENWEATHER_API_KEY not set; the fake-server tests cover this offline")
	}
	weather, err := NewWeatherService(apiKey).GetWeatherForZip("11357")
	require.NoError(t, err)
	require.NotNil(t, weather)
}
