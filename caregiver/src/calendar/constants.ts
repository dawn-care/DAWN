/**
 * constants.ts
 *
 * User: Caregiver
 *
 * Fixed values for the calendar page: the times the wake and bed selects
 * offer, and the wording for the server's schedule and
 * awake-hours error codes (docs/api/screen-and-media.md).
 */

// Shown when bedtime isn't later than wake-up time, by the editor and for the server's INVALID_BED_TIME
export const BED_AFTER_WAKE = "Bedtime must be later than wake-up time.";

// Shown when the server gives a code we have no wording for
export const GENERIC_ERROR = "Something went wrong. Please try again.";

// The times the wake and bed selects offer, as minutes after midnight: the whole day,
// 12:00 am to 11:30 pm. The server takes any times as long as bed is later than wake
// on the same day (screen_settings_service.go, and the database checks the same)
export const AWAKE_RANGE = { earliest: 0, latest: 23 * 60 + 30 };

// The awake-hours selects move in half hours
export const AWAKE_STEP_MINUTES = 30;

// Caregiver-facing wording for the server's error codes, plus NETWORK_ERROR
export const ERROR_MESSAGES: Record<string, string> = {
  INVALID_REQUEST: GENERIC_ERROR,
  INTERNAL_ERROR: GENERIC_ERROR,
  // Not a server code: the request never got an answer
  NETWORK_ERROR: "Can't reach the server. Check your connection and try again.",
  NOT_FOUND: "This was deleted. Refresh to see the latest.",
  INVALID_REPEAT: "Choose how often this repeats.",
  INVALID_DATE: "Choose a valid date.",
  INVALID_TIME: "Choose a valid time.",
  NOT_ON_15_MIN_STEP: "Times move in 15-minute steps.",
  OUTSIDE_AWAKE_HOURS: "That time is outside the carepartner's awake hours. Pick a time that finishes before bedtime.",
  INVALID_SHAPE: "That video, audio or photo choice can't be scheduled together.",
  INVALID_MEDIA: "That item was deleted. Pick another.",
  NOT_AN_OCCURRENCE: "This doesn't happen on that day any more. Refresh to see the latest.",
  INVALID_SCOPE: GENERIC_ERROR,
  INVALID_RANGE: GENERIC_ERROR,
  INVALID_WAKE_TIME: "Choose a wake-up time.",
  INVALID_BED_TIME: BED_AFTER_WAKE,
};
