package device_pairing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeCode(t *testing.T) {
	cases := map[string]string{
		"AB3K9X":      "AB3K9X",
		"ab3k9x":      "AB3K9X",
		"Ab3K9x":      "AB3K9X",
		"  ab3k9x \n": "AB3K9X",
		"   ":         "",
	}
	for typed, want := range cases {
		assert.Equal(t, want, normalizeCode(typed), "%q", typed)
	}
}

func TestVerifyPairingCode_Blank(t *testing.T) {
	_, err := NewDevicePairingService(nil).VerifyPairingCode("  ", 1, "Grandma")
	assert.EqualError(t, err, "pairing code cannot be empty")
}
