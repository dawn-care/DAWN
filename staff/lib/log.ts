/**
 * Event logging stub.
 *
 * Call sites use the shape the backend is expected to accept, so wiring this to
 * a real endpoint later means changing only the body of `logEvent`. Until then
 * events go to the console and an in-memory buffer that can be inspected during
 * development.
 */

export type StaffEvent =
  | { type: 'resident_opened'; residentId: number }
  | { type: 'situation_selected'; residentId: number; situation: string }
  | { type: 'family_folder_opened'; residentId: number; folderId: number }
  | { type: 'play_requested'; residentId: number; clipId: number }
  | { type: 'play_blocked_sleeping_hours'; residentId: number; clipId: number }
  | { type: 'play_stopped'; residentId: number; clipId: number; atSecond: number }
  | { type: 'clip_rated'; residentId: number; clipId: number; rating: string }
  | { type: 'rating_window_expired'; residentId: number; clipId: number }
  | { type: 'sleeping_hours_changed'; residentId: number; start: string; end: string }
  | { type: 'caregiver_request_sent'; residentId: number; topic: string };

interface LoggedEvent {
  event: StaffEvent;
  at: string;
}

const buffer: LoggedEvent[] = [];

export const logEvent = (event: StaffEvent): void => {
  const entry: LoggedEvent = { event, at: new Date().toISOString() };
  buffer.push(entry);
  // eslint-disable-next-line no-console
  console.log('[event]', entry.at, JSON.stringify(event));
};

/** Development helper — returns everything logged this session. */
export const getLoggedEvents = (): LoggedEvent[] => [...buffer];
