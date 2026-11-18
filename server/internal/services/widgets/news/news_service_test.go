package news

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetGoodNews_Live(t *testing.T) {
	apiKey := os.Getenv("NEWS_API_KEY")
	if apiKey == "" {
		t.Skip("NEWS_API_KEY not set; the fake-server tests cover this offline")
	}
	news, err := NewNewsService(apiKey).GetGoodNews()
	require.NoError(t, err)
	require.NotNil(t, news)
}
