/**
 * Fake data for the parts of the staff app with no backend yet: family folders
 * and guidance notes, sleeping hours, ratings and requests to the caregiver.
 * Residents and clips are real (data/server.ts).
 *
 * Every function returns a Promise, so replacing one with a real call later
 * doesn't change the screens.
 */

import type { FamilyFolder, Rating, Situation, SleepingHours } from './types';

export const SITUATIONS: Situation[] = [
  {
    id: 'asking_for_someone',
    label: 'Asking for someone',
    icon: 'people-outline',
    routesThroughFamily: true,
  },
  {
    id: 'wants_to_go_home',
    label: 'Wants to go home',
    icon: 'home-outline',
    routesThroughFamily: false,
  },
  {
    id: 'doesnt_know_where',
    label: "Doesn't know where they are",
    icon: 'location-outline',
    routesThroughFamily: false,
  },
  {
    id: 'anxious_or_upset',
    label: 'Anxious or upset',
    icon: 'heart-outline',
    routesThroughFamily: false,
  },
  {
    id: 'starting_the_day',
    label: 'Starting the day',
    icon: 'sunny-outline',
    routesThroughFamily: false,
  },
  {
    id: 'winding_down',
    label: 'Winding down for night',
    icon: 'moon-outline',
    routesThroughFamily: false,
  },
];

/** The same fake family for every resident, until family folders exist on the server. */
const FAMILY_FOLDERS: FamilyFolder[] = [
  { id: 1, name: 'Elena', relationship: 'Daughter' },
  { id: 2, name: 'Marco', relationship: 'Son' },
  { id: 3, name: 'Sofia', relationship: 'Granddaughter' },
  {
    id: 4,
    name: 'Daniel',
    relationship: 'Husband',
    guidanceNote:
      "He passed away in 2023. Please don't tell her again. Reassure her and talk about her garden.",
  },
];

/** Matches the server's default awake hours (wake 07:00, bed 21:00). */
const DEFAULT_SLEEPING_HOURS: SleepingHours = { start: '21:00', end: '07:00' };

/** Edits made in the app this session. Not saved anywhere yet. */
const sleepingHours = new Map<number, SleepingHours>();

/** Simulated latency so loading states are visible during development. */
const delay = <T>(value: T, ms = 150): Promise<T> =>
  new Promise((resolve) => setTimeout(() => resolve(value), ms));

export const getFamilyFolders = () => delay([...FAMILY_FOLDERS]);

export const getFamilyFolder = (folderId: number) =>
  delay(FAMILY_FOLDERS.find((f) => f.id === folderId));

export const getSleepingHours = (residentId: number) =>
  delay(sleepingHours.get(residentId) ?? DEFAULT_SLEEPING_HOURS);

/** Staff and caregivers may both edit these. Kept in memory only for now. */
export const setSleepingHours = (residentId: number, hours: SleepingHours) => {
  sleepingHours.set(residentId, hours);
  return delay(hours);
};

/** Ratings aren't saved yet; the screen still behaves as if they were. */
export const rateClip = (_playbackId: number, _rating: Rating) => delay(undefined);
