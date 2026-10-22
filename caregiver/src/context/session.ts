/**
 * session.ts
 *
 * User: Caregiver
 *
 * Signing in and out, shared by RequireLogin, Login, Settings and AccountSetup:
 * who the server says is signed in (GET /me), which carepartner to manage, and
 * where the caregiver goes after logging in.
 */

import { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { useToast } from "../components/ui/toastContext";
import { useAuth } from "./AuthContext";
import type { User } from "./AuthContext";

const API_URL = import.meta.env.VITE_API_URL;

/** The signed-in caregiver, or null when nobody is (401). Throws on any other answer */
export async function fetchCurrentUser(): Promise<User | null> {
  const res = await fetch(`${API_URL}/me`, { credentials: "include" });
  if (res.status === 401) return null;
  if (!res.ok) throw new Error(`GET /me answered ${res.status}`);
  return res.json();
}

/**
 * The carepartner to manage: the saved one while it is still the caregiver's
 * (it may be from another account, or unpaired since), otherwise their first
 */
export function pickCarepartner(saved: number | null, user: User): number | null {
  const carepartners = user.carepartners ?? [];
  if (saved !== null && carepartners.some((carepartner) => carepartner.id === saved)) return saved;
  return carepartners[0]?.id ?? null;
}

/**
 * Where a caregiver goes once logged in: pairing when they have no carepartner
 * yet, otherwise the page they were sent to /login from (`from`), or home
 */
export function pathAfterLogin(user: User, from: string | undefined): string {
  if (!user.carepartners?.length) return "/setup";
  return from && !from.startsWith("/login") ? from : "/";
}

/** Puts a signed-in caregiver into AuthContext */
export function useSignIn(): (user: User) => void {
  const { setUser, setIsAuthenticated, setActiveCarePartnerId } = useAuth();
  return useCallback(
    (user: User) => {
      setUser(user);
      setActiveCarePartnerId((saved) => pickCarepartner(saved, user));
      setIsAuthenticated(true);
    },
    [setUser, setIsAuthenticated, setActiveCarePartnerId],
  );
}

/** Ends the session on the server, clears AuthContext and goes to /login; a failure shows a toast */
export function useLogout(): () => Promise<void> {
  const { setUser, setIsAuthenticated, setActiveCarePartnerId } = useAuth();
  const navigate = useNavigate();
  const toast = useToast();

  return async () => {
    try {
      const res = await fetch(`${API_URL}/auth/logout`, { method: "POST", credentials: "include" });
      if (!res.ok) throw new Error(`POST /auth/logout answered ${res.status}`);
    } catch (err) {
      console.error("Logout failed:", err);
      toast.show("Couldn't log out. Check your connection and try again.");
      return;
    }
    setIsAuthenticated(false);
    setUser(null);
    setActiveCarePartnerId(null);
    navigate("/login");
  };
}
