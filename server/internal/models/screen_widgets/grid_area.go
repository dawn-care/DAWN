package screen_widgets

// Cell is one square of the tablet's grid. Columns and rows count from 0.
type Cell struct {
	Col int `json:"col"`
	Row int `json:"row"`
}

/*
GridArea is where a widget sits on the grid, in the shape the tablet draws from
(GridArea in shared/widgets/grid.ts): its top-left cell and its size in cells.
*/
type GridArea struct {
	Col    int `json:"col"`
	Row    int `json:"row"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

/*
AreaFrom converts a widget's corners to a GridArea. Both corners are inclusive, so a
widget whose corners are the same cell is 1x1. This is the only place the server
converts corners to a size; cornersToArea in shared/widgets/grid.ts is the
frontend's copy of the same rule.
*/
func AreaFrom(topLeft, bottomRight Cell) GridArea {
	return GridArea{
		Col:    topLeft.Col,
		Row:    topLeft.Row,
		Width:  bottomRight.Col - topLeft.Col + 1,
		Height: bottomRight.Row - topLeft.Row + 1,
	}
}
