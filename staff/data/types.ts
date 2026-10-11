/**
 * Domain types for the staff app.
 *
 * These mirror the shapes the backend is expected to return once Phase 1 of the
 * implementation plan lands. Until then they are served from data/mock.ts.
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
  room: string;
  /** True when family has uploaded content the staff haven't seen yet. */
  hasNewContent: boolean;
  /** Local times, "HH:MM". Staff cannot play media between these. */
  sleepingHours: { start: string; end: string };
  lastPlayed?: LastPlayed;
}

export interface LastPlayed {
  title: string;
  fromName: string;
  /** "8:12 am" */
  at: string;
  rating?: Rating;
}

/** Family members are folders for grouping clips, not accounts. */
export interface FamilyFolder {
  id: number;
  residentId: number;
  name: string;
  relationship: string;
  clipCount: number;
  /** Caregiver-authored. Only caregivers may edit this. */
  guidanceNote?: string;
}

export interface Clip {
  id: number;
  residentId: number;
  familyFolderId: number | null;
  title: string;
  fromName: string;
  /** Seconds. */
  duration: number;
  recordedOn: string;
  situations: SituationId[];
  playCount: number;
  helpedCount: number;
  /** Caregiver-authored note shown to staff before playing. */
  familyNote?: string;
  /** True for generic facility content rather than family content. */
  isFacilityContent?: boolean;
}

export type Rating = 'helped' | 'no_change' | 'worse';
