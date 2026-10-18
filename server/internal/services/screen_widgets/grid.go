package screen_widgets

import (
	"errors"
	"fmt"

	model "server/internal/models/screen_widgets"
)

// The tablet's grid size, in cells. Must match GRID in shared/widgets/grid.ts.
// Kept in Go rather than the database so resizing the grid needs no migration.
const (
	GridColumns = 4
	GridRows    = 3
)

var (
	ErrInvalidPosition = fmt.Errorf("top_left and bottom_right must be cells inside the %dx%d grid, with bottom_right not above or left of top_left", GridColumns, GridRows)
	ErrOverlap         = errors.New("widgets overlap")
)

func insideGrid(c model.Cell) bool {
	return c.Col >= 0 && c.Col < GridColumns && c.Row >= 0 && c.Row < GridRows
}

// checkArea checks a widget's corners: both present, both on the grid, and bottomRight
// not above or left of topLeft
func checkArea(topLeft, bottomRight *model.Cell) error {
	if topLeft == nil || bottomRight == nil {
		return fmt.Errorf("%w: top_left and bottom_right are required", ErrInvalidPosition)
	}
	if !insideGrid(*topLeft) || !insideGrid(*bottomRight) {
		return fmt.Errorf("%w: (%d,%d) to (%d,%d) is off the grid", ErrInvalidPosition,
			topLeft.Col, topLeft.Row, bottomRight.Col, bottomRight.Row)
	}
	if bottomRight.Col < topLeft.Col || bottomRight.Row < topLeft.Row {
		return fmt.Errorf("%w: bottom_right (%d,%d) is above or left of top_left (%d,%d)", ErrInvalidPosition,
			bottomRight.Col, bottomRight.Row, topLeft.Col, topLeft.Row)
	}
	return nil
}

/*
findOverlap reports the first two areas that cover the same cell, by their index in
areas, marking each area's cells in turn. The areas must already fit the grid
(checkArea).
*/
func findOverlap(areas []model.GridArea) (first, second int, found bool) {
	var owner [GridRows][GridColumns]int // index + 1 of the area covering each cell; 0 is free
	for i, a := range areas {
		for row := a.Row; row < a.Row+a.Height; row++ {
			for col := a.Col; col < a.Col+a.Width; col++ {
				if prev := owner[row][col]; prev != 0 {
					return prev - 1, i, true
				}
				owner[row][col] = i + 1
			}
		}
	}
	return 0, 0, false
}
