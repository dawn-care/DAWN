/**
 * constants.ts
 *
 * User: Caregiver
 *
 * Fixed values for the calendar page: the range of times the server accepts
 * for waking and bedtime, and the wording for the server's schedule and
 * awake-hours error codes (docs/api/screen-and-media.md).
 */

// Shown when the server gives a code we have no wording for
export const GENERIC_ERROR = "Something went wrong. Please try again.";

// Wake and bed times the server accepts (EarliestWake .. LatestBed), as minutes after midnight
export const WAKE_RANGE = { earliest: 6 * 60, latest: 10 * 60 };
export const BED_RANGE = { earliest: 19 * 60, latest: 22 * 60 };

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
  INVALID_WAKE_TIME: "Wake-up time must be between 6:00 am and 10:00 am.",
  INVALID_BED_TIME: "Bedtime must be between 7:00 pm and 10:00 pm.",
};
