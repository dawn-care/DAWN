package screen_widgets

import "encoding/json"

// The screen layout API's request and response bodies. They live in this package, next
// to ScreenWidget and Cell, because swag can't follow a type in the service package to
// one here: both packages are named screen_widgets.

/*
WidgetInput is one widget of a saved layout. With ID it updates that widget (its type
can't change); without ID it creates a new one. TopLeft and BottomRight are the cells at
its corners, both inclusive.
*/
type WidgetInput struct {
	ID           *int            `json:"id"`
	Type         string          `json:"type"`
	TopLeft      *Cell           `json:"top_left"`
	BottomRight  *Cell           `json:"bottom_right"`
	Config       json.RawMessage `json:"config" swaggertype:"object"`
	PhotoMediaID *int            `json:"photo_media_id"`
}

/*
LayoutInput is every widget on the carepartner's screen. Widgets is a pointer so a
missing key or null can be told apart from [], which clears the screen on purpose.
*/
type LayoutInput struct {
	Widgets *[]WidgetInput `json:"widgets"`
}

// Layout is a carepartner's widgets, as the screen editor loads and saves them
type Layout struct {
	Widgets []ScreenWidget `json:"widgets"`
}
