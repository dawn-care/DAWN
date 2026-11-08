package tablet_events

import "time"

// Screens touches are recorded on, matching the CHECK on "TabletTouch".screen. The
// pair screen is for staff, so the tablet never records touches there.
const (
	ScreenAlwaysOn = "always_on"
	ScreenDefault  = "default"
	ScreenNight    = "night"
	ScreenPlaying  = "playing"
)

// What was under the finger. A touch with nothing to name (the night screen) has
// no target type at all.
const (
	TargetWidget = "widget"
	TargetEmpty  = "empty"
	TargetMedia  = "media"
)

const (
	GestureTap       = "tap"
	GestureLongPress = "long_press"
	GestureSwipe     = "swipe"
)

const (
	MediaVideo = "video"
	MediaAudio = "audio"
	MediaPhoto = "photo"
)

var (
	Screens     = []string{ScreenAlwaysOn, ScreenDefault, ScreenNight, ScreenPlaying}
	TargetTypes = []string{TargetWidget, TargetEmpty, TargetMedia}
	Gestures    = []string{GestureTap, GestureLongPress, GestureSwipe}
	MediaKinds  = []string{MediaVideo, MediaAudio, MediaPhoto}
)

/*
TouchEvent is one event as the tablet sends it and as a "TabletTouch" row stores it
(see carepartner/src/telemetry/touchLog.ts). Quick repeats of the same gesture on the
same thing are one event with a count, first_at the first touch and last_at the last.
WidgetType and WidgetID are set only for a widget, MediaKind only for media.
WidgetID is the ScreenWidget id as a string, kept as sent even after the widget is
deleted.
*/
type TouchEvent struct {
	Screen     string    `json:"screen"`
	TargetType *string   `json:"target_type"`
	WidgetType *string   `json:"widget_type"`
	WidgetID   *string   `json:"widget_id"`
	MediaKind  *string   `json:"media_kind"`
	Gesture    string    `json:"gesture"`
	Count      int       `json:"count"`
	FirstAt    time.Time `json:"first_at"`
	LastAt     time.Time `json:"last_at"`
}
