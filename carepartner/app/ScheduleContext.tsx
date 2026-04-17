/**
 * ScheduleContext.tsx
 * 
 * User: Carepartner
 * 
 * Provides scheduled start for daily agenda and access to today's schedule items.
 * 
 * Responsibilities:
 * - Fetch and store carepartner data
 * - Fetch and store daily schedule configurations
 * - Track played items to prevent repeats
 * - Handle auto-start and scheduled triggers
 * - Provide refresh functionality across the app
 * 
 * Navigation:
 * - Auto-start trigger → `/agendaView`
 * - Scheduled trigger → `/agendaView`
 * 
 * Notes:
 * - Polls backend every 5 minutes
 * - Resets state at midnight
 */

import React, { createContext, useContext, useEffect, useRef, useState } from 'react';
import { router } from 'expo-router';
import { getCarepartnerId } from './carepartnerId';
import { API_URL } from './config';
import type { ConfigurationWithMetadata } from './agendaView';

/* ============================
   Type Definitions
============================ */

/**
 * Carepartner data structure from database
 */
type Carepartner = {
  id: number;
  name: string;
  auto_start: string; // "HH:MM:SS"
};

/* ============================
   Context Setup
============================ */

interface ScheduleContextType {
  todaysItems: ConfigurationWithMetadata[];
  loading: boolean;
  error: string | null;
  carepartner: Carepartner | null;
  playedIds: Set<number>;
  markPlayed: (id: number) => void;
  refresh: () => Promise<void>;
}

const ScheduleContext = createContext<ScheduleContextType | null>(null);

export function useSchedule() {
  const ctx = useContext(ScheduleContext);
  if (!ctx) throw new Error('useSchedule must be used within ScheduleProvider');
  return ctx;
}

/* ============================
   Helpers
============================ */

/**
 * Formats backend time (HH:MM:SS) into user-friendly 12-hour format.
 * Example:
 *  "08:00:00" → "8:00 AM"
 */
function todayDateStr() {
  const now = new Date();
  const y = now.getFullYear();
  const m = String(now.getMonth() + 1).padStart(2, '0');
  const d = String(now.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

function nowHHMM() {
  const now = new Date();
  return `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
}

/* ============================
   Component
============================ */

/**
 * ScheduleProvider Component
 * 
 * Handles:
 *  - Global schedule state
 *  - API fetching
 *  - Auto triggers
 *  - Played-item tracking
 */
export function ScheduleProvider({ children }: { children: React.ReactNode }) {

  /* ============================
     State
  ============================ */
  const [carepartner, setCarepartner] = useState<Carepartner | null>(null);
  const [todaysItems, setTodaysItems] = useState<ConfigurationWithMetadata[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const lastAutoStartDateRef = useRef<string | null>(null);
  /** Dedup ref so the scheduled-item trigger navigates at most once per item */
  const navFiredRef = useRef<Set<number>>(new Set());

  /** Track which items have already been played today */
  const [playedIds, setPlayedIds] = useState<Set<number>>(new Set());
  const playedIdsRef = useRef<Set<number>>(playedIds);
  useEffect(() => { playedIdsRef.current = playedIds; }, [playedIds]);

  /** Track today's schedule items */
  const todaysItemsRef = useRef<ConfigurationWithMetadata[]>(todaysItems);
  useEffect(() => { todaysItemsRef.current = todaysItems; }, [todaysItems]);

  const markPlayed = (id: number) => {
    setPlayedIds(prev => {
      const next = new Set(prev);
      next.add(id);
      return next;
    });
  };

  /** Fetch carepartner info */
  const fetchCarepartner = async (): Promise<Carepartner | null> => {
    const cpId = await getCarepartnerId();
    if (!cpId) return null;
    const res = await fetch(`${API_URL}/users/carepartner/${cpId}`);
    if (!res.ok) return null;
    const data: Carepartner = await res.json();
    setCarepartner(data);
    return data;
  };

  /** Fetch today's configurations */
  const fetchConfigurations = async () => {
    try {
      setLoading(true);
      setError(null);

      const cpId = await getCarepartnerId();
      if (!cpId) {
        setError('No carepartner ID found.');
        return;
      }

      const now = new Date();
      const params = new URLSearchParams({
        carepartner_id: String(cpId),
        day: String(now.getDate()),
        month: String(now.getMonth() + 1),
        day_of_week: String(now.getDay()),
        date: todayDateStr(),
      });

      const res = await fetch(`${API_URL}/configurations/day?${params}`);
      if (!res.ok) throw new Error('Failed to fetch configurations');
      const raw = await res.json();
      const data: ConfigurationWithMetadata[] = Array.isArray(raw) ? raw : [];

      setTodaysItems(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  /** Sync with caregiver */
  const refresh = async () => {
    try {
      await fetchCarepartner();
      await fetchConfigurations();
    } catch (error) {
      console.error('Error refreshing schedule:', error); 
    }
  };

  /** Initial fetch */
  useEffect(() => {
    refresh();
  }, []);

  /** Poll for configuration & auto-start updates every 5 minutes */
  useEffect(() => {
    const FIVE_MINUTES = 5 * 60 * 1000;
    const id = setInterval(() => {
      refresh();
    }, FIVE_MINUTES);
    return () => clearInterval(id);
  }, []);

  /** Reset played-item tracking at midnight */
  useEffect(() => {
    const now = new Date();
    const tomorrow = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1);
    const msUntilMidnight = tomorrow.getTime() - now.getTime();
    const timer = setTimeout(() => {
      setPlayedIds(new Set());
      navFiredRef.current = new Set();
      lastAutoStartDateRef.current = null;
    }, msUntilMidnight);
    return () => clearTimeout(timer);
  }, []);

  /** Global auto-start trigger */
  useEffect(() => {
    const check = () => {
      if (!carepartner?.auto_start) return;
      const now = new Date();
      const dateKey = now.toDateString();
      const [hStr, mStr] = carepartner.auto_start.split(':');
      const h = parseInt(hStr, 10);
      const m = parseInt(mStr, 10);

      if (now.getHours() !== h || now.getMinutes() !== m) return;
      if (lastAutoStartDateRef.current === dateKey) return;

      /** Only trigger if there are unplayed non-timed items */
      const hasUnplayed = todaysItemsRef.current.some(
        i => !i.timed && !playedIdsRef.current.has(i.id)
      );
      if (!hasUnplayed) return;

      lastAutoStartDateRef.current = dateKey;
      router.navigate({ pathname: '/agendaView', params: { trigger: 'auto_start' } } as any);
    };

    check();
    const id = setInterval(check, 15000);
    return () => clearInterval(id);
  }, [carepartner]);

  /** Global scheduled-item trigger (navigate only, no queue management) */
  useEffect(() => {
    const scheduled = todaysItems.filter(i => i.timed);
    if (!scheduled.length) return;

    const check = () => {
      const currentTime = nowHHMM();
      let shouldNavigate = false;

      for (const item of scheduled) {
        if (navFiredRef.current.has(item.id)) continue;
        if (playedIdsRef.current.has(item.id)) continue;
        const itemTime = (item.time ?? '').slice(0, 5);
        if (itemTime === currentTime) {
          navFiredRef.current.add(item.id);
          shouldNavigate = true;
        }
      }

      if (shouldNavigate) {
        router.navigate({ pathname: '/agendaView', params: { trigger: 'scheduled' } } as any);
      }
    };

    check();
    const id = setInterval(check, 15000);
    return () => clearInterval(id);
  }, [todaysItems]);

  return (
    <ScheduleContext.Provider
      value={{
        todaysItems,
        loading,
        error,
        carepartner,
        playedIds,
        markPlayed,
        refresh,
      }}
    >
      {children}
    </ScheduleContext.Provider>
  );
}
