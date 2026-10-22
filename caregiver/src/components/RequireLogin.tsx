/**
 * RequireLogin.tsx
 *
 * User: Caregiver
 *
 * Layout route that sends anyone who isn't logged in to /login. The session is
 * a cookie the browser keeps, so after a reload the app asks the server who is
 * signed in (GET /me) before showing anything, and keeps the answer in
 * AuthContext so it only asks once. A caregiver who has carepartners but no
 * saved selection (cleared site data, another browser) gets the first one.
 */

import { useEffect, useState } from "react";
import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import type { User } from "../context/AuthContext";

const API_URL = import.meta.env.VITE_API_URL;

export default function RequireLogin() {
  const { isAuthenticated, setIsAuthenticated, setUser, setActiveCarePartnerId } = useAuth();
  // "checking" until the server has answered
  const [state, setState] = useState<"checking" | "in" | "out">(isAuthenticated ? "in" : "checking");

  useEffect(() => {
    if (isAuthenticated) return;

    let current = true;
    fetch(`${API_URL}/me`, { credentials: "include" })
      .then(async (res) => {
        if (!current) return;
        if (!res.ok) {
          setState("out");
          return;
        }
        const user: User = await res.json();
        if (!current) return;
        setUser(user);
        setActiveCarePartnerId((saved) => saved ?? user.carepartners?.[0]?.id ?? null);
        setIsAuthenticated(true);
        setState("in");
      })
      .catch(() => current && setState("out"));

    return () => {
      current = false;
    };
  }, [isAuthenticated, setIsAuthenticated, setUser, setActiveCarePartnerId]);

  if (state === "out" && !isAuthenticated) return <Navigate to="/login" replace />;
  // Nothing to show until we know; the answer takes a moment
  if (state === "checking" && !isAuthenticated) return null;
  return <Outlet />;
}
