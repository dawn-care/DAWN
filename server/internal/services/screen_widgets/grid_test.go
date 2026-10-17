package screen_widgets

import (
	"testing"

	"github.com/stretchr/testify/assert"

	model "server/internal/models/screen_widgets"
)

func cell(col, row int) *model.Cell { return &model.Cell{Col: col, Row: row} }

func TestCheckArea(t *testing.T) {
	valid := []struct {
		name                 string
		topLeft, bottomRight *model.Cell
	}{
		{"1x1", cell(1, 1), cell(1, 1)},
		{"the bottom-right cell", cell(3, 2), cell(3, 2)},
		{"the whole grid", cell(0, 0), cell(GridColumns-1, GridRows-1)},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, checkArea(tc.topLeft, tc.bottomRight))
		})
	}

	invalid := []struct {
		name                 string
		topLeft, bottomRight *model.Cell
	}{
		{"missing top_left", nil, cell(0, 0)},
		{"missing bottom_right", cell(0, 0), nil},
		{"columns reversed", cell(2, 0), cell(1, 0)},
		{"rows reversed", cell(0, 2), cell(0, 1)},
		{"negative column", cell(-1, 0), cell(0, 0)},
		{"negative row", cell(0, -1), cell(0, 0)},
		{"column 4 is off the grid", cell(3, 0), cell(4, 0)},
		{"row 3 is off the grid", cell(0, 2), cell(0, 3)},
		{"both corners off the grid", cell(4, 3), cell(5, 4)},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			assert.ErrorIs(t, checkArea(tc.topLeft, tc.bottomRight), ErrInvalidPosition)
		})
	}
}

func area(col, row, width, height int) model.GridArea {
	return model.GridArea{Col: col, Row: row, Width: width, Height: height}
}

func TestFindOverlap(t *testing.T) {
	cases := []struct {
		name          string
		areas         []model.GridArea
		found         bool
		first, second int
	}{
		{"no widgets", nil, false, 0, 0},
		{"one widget covering the grid", []model.GridArea{area(0, 0, 4, 3)}, false, 0, 0},
		{"separate widgets", []model.GridArea{area(0, 0, 1, 1), area(3, 2, 1, 1)}, false, 0, 0},
		{"touching edge to edge", []model.GridArea{area(0, 0, 2, 2), area(2, 0, 2, 2), area(0, 2, 4, 1)}, false, 0, 0},
		{"a 1x1 inside a 2x2", []model.GridArea{area(0, 0, 2, 2), area(1, 1, 1, 1)}, true, 0, 1},
		{"an exact duplicate", []model.GridArea{area(1, 0, 1, 1), area(1, 0, 1, 1)}, true, 0, 1},
		{"overlap only in the last cell", []model.GridArea{area(0, 0, 4, 3), area(3, 2, 1, 1)}, true, 0, 1},
		{"reports the widgets involved, not the first two", []model.GridArea{area(0, 0, 1, 1), area(2, 0, 2, 1), area(3, 0, 1, 3)}, true, 1, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			first, second, found := findOverlap(tc.areas)
			assert.Equal(t, tc.found, found)
			if tc.found {
				assert.Equal(t, []int{tc.first, tc.second}, []int{first, second})
			}
		})
	}
}
