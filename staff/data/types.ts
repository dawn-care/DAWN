/**
 * Domain types for the staff app.
 *
 * Residents, clips, family folders and what last played come from the server
 * (data/server.ts). Sleeping hours and ratings have no backend yet and come from
 * data/mock.ts.
 */

/** The six situations a staff member can report from the resident screen. */
export type SituationId =
  | 'asking_for_someone'
  | 'wants_to_go_home'
  | 'doesnt_know_where'
  | 'anxious_or_upset'
  | 'starting_the_day'
  | 'winding_down';

export interface Situation {
  id: SituationId;
  label: string;
  /** Ionicons glyph name. */
  icon: string;
  /** Whether this situation routes through the family-folder picker first. */
  routesThroughFamily: boolean;
}

/** A resident (called a "carepartner" elsewhere in the codebase). */
export interface Resident {
  id: number;
  name: string;
}

/** Local times, "HH:MM". Staff cannot play media between these. */
export interface SleepingHours {
  start: string;
  end: string;
}

/** Family members are folders for grouping clips, not accounts. Added by the caregiver. */
export interface FamilyFolder {
  id: number;
  name: string;
  relationship: string;
  /** Caregiver-authored. Only caregivers may edit this. */
  guidanceNote?: string;
}

/** A video or audio clip from the resident's media library on the server. */
export interface Clip {
  /** The server's media ID, which is what the play command sends. */
  id: number;
  kind: 'video' | 'audio';
  title: string;
  /** Seconds, or null if the server hasn't measured it. */
  duration: number | null;
  /** RFC 3339, when it was uploaded. */
  addedAt: string;
}

/** Something that reached the resident's tablet, from any source. */
export interface PlayedItem {
  /** The name it had when it played, even if the media was deleted since. */
  title: string;
  kind: 'photo' | 'video' | 'audio';
  /** Who started it: staff, the schedule, or a tap on the tablet. */
  source: 'staff' | 'scheduled' | 'tap';
  /** "started" means the tablet hasn't reported an end yet. */
  status: 'started' | 'finished' | 'stopped' | 'failed';
  /** RFC 3339, when it started on the tablet. */
  startedAt: string;
}

export type Rating = 'helped' | 'no_change' | 'worse';
