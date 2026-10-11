/**
 * Mock data + a thin async layer standing in for the backend.
 *
 * Every function here returns a Promise so that swapping these for real HTTP
 * calls later is a change to this file only — screens do not know the
 * difference. See the implementation plan: the staff app frontend is built
 * before any backend work.
 */

import type {
  Clip,
  FamilyFolder,
  Rating,
  Resident,
  Situation,
  SituationId,
} from './types';

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

const RESIDENTS: Resident[] = [
  {
    id: 1,
    name: 'Margaret Doyle',
    room: '210',
    hasNewContent: false,
    sleepingHours: { start: '20:00', end: '07:00' },
    lastPlayed: {
      title: 'Morning message',
      fromName: 'Claire',
      at: '8:12 am',
      rating: 'helped',
    },
  },
  {
    id: 2,
    name: 'Arthur Pennington',
    room: '212',
    hasNewContent: false,
    sleepingHours: { start: '21:00', end: '06:30' },
  },
  {
    id: 3,
    name: 'Rosa Alvarez',
    room: '214',
    hasNewContent: true,
    sleepingHours: { start: '20:30', end: '07:00' },
    lastPlayed: {
      title: "I'll see you Sunday",
      fromName: 'Elena',
      at: '8:12 am',
      rating: 'helped',
    },
  },
  {
    id: 4,
    name: 'Henry Osei',
    room: '216',
    hasNewContent: false,
    sleepingHours: { start: '20:00', end: '07:30' },
  },
  {
    id: 5,
    name: 'Dorothy Kaminski',
    room: '218',
    hasNewContent: false,
    sleepingHours: { start: '19:30', end: '06:00' },
    lastPlayed: {
      title: 'Garden in summer',
      fromName: 'Peter',
      at: '2:40 pm',
      rating: 'no_change',
    },
  },
];

const FAMILY_FOLDERS: FamilyFolder[] = [
  {
    id: 1,
    residentId: 3,
    name: 'Elena',
    relationship: 'Daughter',
    clipCount: 4,
  },
  { id: 2, residentId: 3, name: 'Marco', relationship: 'Son', clipCount: 2 },
  {
    id: 3,
    residentId: 3,
    name: 'Sofia',
    relationship: 'Granddaughter',
    clipCount: 1,
  },
  {
    id: 4,
    residentId: 3,
    name: 'Daniel',
    relationship: 'Husband',
    clipCount: 0,
    guidanceNote:
      "He passed away in 2023. Please don't tell her again. Reassure her and talk about her garden.",
  },
];

const CLIPS: Clip[] = [
  {
    id: 1,
    residentId: 3,
    familyFolderId: 1,
    title: "I'll see you Sunday",
    fromName: 'Elena',
    duration: 45,
    recordedOn: 'Sep 27',
    situations: ['asking_for_someone', 'anxious_or_upset'],
    playCount: 3,
    helpedCount: 2,
    familyNote: 'Works best if you sit with her.',
  },
  {
    id: 2,
    residentId: 3,
    familyFolderId: 1,
    title: "You're safe, Mom",
    fromName: 'Elena',
    duration: 30,
    recordedOn: 'Sep 14',
    situations: ['anxious_or_upset', 'doesnt_know_where'],
    playCount: 1,
    helpedCount: 1,
  },
  {
    id: 3,
    residentId: 3,
    familyFolderId: 2,
    title: 'Grandkids at the lake',
    fromName: 'Marco',
    duration: 52,
    recordedOn: 'Aug 30',
    situations: ['asking_for_someone', 'starting_the_day'],
    playCount: 2,
    helpedCount: 1,
  },
  {
    id: 4,
    residentId: 3,
    familyFolderId: null,
    title: 'This is your home now',
    fromName: "St. Ann's staff",
    duration: 30,
    recordedOn: 'Jun 02',
    situations: ['wants_to_go_home', 'doesnt_know_where'],
    playCount: 8,
    helpedCount: 5,
    isFacilityContent: true,
  },
];

/** Simulated network latency so loading states are visible during development. */
const delay = <T>(value: T, ms = 220): Promise<T> =>
  new Promise((resolve) => setTimeout(() => resolve(value), ms));

export const getResidents = () => delay([...RESIDENTS]);

export const getResident = (id: number) =>
  delay(RESIDENTS.find((r) => r.id === id));

export const getFamilyFolders = (residentId: number) =>
  delay(FAMILY_FOLDERS.filter((f) => f.residentId === residentId));

export const getFamilyFolder = (folderId: number) =>
  delay(FAMILY_FOLDERS.find((f) => f.id === folderId));

export const getClip = (clipId: number) =>
  delay(CLIPS.find((c) => c.id === clipId));

export const getClipsForFamilyFolder = (folderId: number) =>
  delay(CLIPS.filter((c) => c.familyFolderId === folderId));

export const getClipsForSituation = (
  residentId: number,
  situation: SituationId
) =>
  delay(
    CLIPS.filter(
      (c) => c.residentId === residentId && c.situations.includes(situation)
    )
  );

/** Updates sleeping hours locally. Staff and caregivers may both edit these. */
export const setSleepingHours = (
  residentId: number,
  hours: { start: string; end: string }
) => {
  const resident = RESIDENTS.find((r) => r.id === residentId);
  if (resident) resident.sleepingHours = hours;
  return delay(resident);
};

export const rateClip = (clipId: number, rating: Rating) => {
  const clip = CLIPS.find((c) => c.id === clipId);
  if (clip) {
    clip.playCount += 1;
    if (rating === 'helped') clip.helpedCount += 1;
  }
  return delay(clip);
};
