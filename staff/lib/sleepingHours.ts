/**
 * Sleeping-hours rule.
 *
 * Staff cannot play media on a resident's tablet during their sleeping hours.
 * The window normally crosses midnight (e.g. 20:30 → 07:00), so the comparison
 * wraps.
 *
 * This is enforced in the UI only for now. It must also be enforced server-side
 * once the backend exists — a client-side check is a usability guard, not a
 * safety guarantee.
 */

/** Minutes since midnight for an "HH:MM" string. */
const toMinutes = (hhmm: string): number => {
  const [h, m] = hhmm.split(':').map(Number);
  return h * 60 + m;
};

export const isAsleep = (
  hours: { start: string; end: string },
  now: Date = new Date()
): boolean => {
  const current = now.getHours() * 60 + now.getMinutes();
  const start = toMinutes(hours.start);
  const end = toMinutes(hours.end);

  // Window crosses midnight: asleep if after start OR before end.
  if (start > end) return current >= start || current < end;
  return current >= start && current < end;
};

/** "8:30 pm" for an "HH:MM" string. */
export const formatTime = (hhmm: string): string => {
  const [h, m] = hhmm.split(':').map(Number);
  const period = h >= 12 ? 'pm' : 'am';
  const hour12 = h % 12 === 0 ? 12 : h % 12;
  return `${hour12}:${String(m).padStart(2, '0')} ${period}`;
};

export const describeWindow = (hours: { start: string; end: string }): string =>
  `${formatTime(hours.start)} – ${formatTime(hours.end)}`;
