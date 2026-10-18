package screen_widgets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAreaFrom(t *testing.T) {
	cases := []struct {
		name                 string
		topLeft, bottomRight Cell
		want                 GridArea
	}{
		{"1x1: the same cell for both corners", Cell{2, 1}, Cell{2, 1}, GridArea{Col: 2, Row: 1, Width: 1, Height: 1}},
		{"2x2", Cell{0, 0}, Cell{1, 1}, GridArea{Col: 0, Row: 0, Width: 2, Height: 2}},
		{"4x1 along the bottom", Cell{0, 2}, Cell{3, 2}, GridArea{Col: 0, Row: 2, Width: 4, Height: 1}},
		{"the whole 4x3 grid", Cell{0, 0}, Cell{3, 2}, GridArea{Col: 0, Row: 0, Width: 4, Height: 3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, AreaFrom(tc.topLeft, tc.bottomRight))
		})
	}
}
