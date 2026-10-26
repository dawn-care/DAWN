/**
 * Login.tsx
 * 
 * User: Caregiver
 * 
 * This component handles user authentication via Google OAuth. The page is a
 * brand panel (the nav bar's dark surface, with a sunrise) beside the sign-in card.
 * 
 * Responsibilities:
 *  - Initiate Google login flow
 *  - Send ID token to backend for verification
 *  - Fetch authenticated user data
 *  - Update global AuthContext
 *  - Redirect user based on account state
 * 
 * Redirect logic (pathAfterLogin in context/session.ts):
 *  - No carepartners → Account setup page
 *  - Has carepartners → the page RequireLogin sent them here from, or home (`/`)
 *  - Already signed in (checked with GET /me on arrival) → straight there
 */

import { GoogleLogin } from "@react-oauth/google";
import type { CredentialResponse } from "@react-oauth/google";
import { useEffect, useState } from "react";
import { Calendar, Image, Monitor, Users } from "lucide-react";
import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { fetchCurrentUser, pathAfterLogin, useSignIn } from "../context/session";

const FEATURES = [
  { icon: Monitor, text: "Choose what your carepartner sees every day" },
  { icon: Calendar, text: "Schedule videos and audio throughout their day" },
  { icon: Image, text: "Keep familiar faces and voices close with custom videos, audio and photos" },
  { icon: Users, text: "Connect with staff at their facility, to best help your carepartner" },
];

/* ============================
   Component
============================ */

/**
 * Login Component
 * 
 * Handles:
 *  - Google authentication
 *  - Backend session creation
 *  - User state initialization
 */
const Login = () => {
  
  /* ============================
     Routing / Context
  ============================ */

  const { isAuthenticated, user } = useAuth();
  const signIn = useSignIn();
  // RequireLogin sends the page the caregiver was trying to open
  const from = (useLocation().state as { from?: string } | null)?.from;

  /**
   * Base API URL
   */
  const API_URL = import.meta.env.VITE_API_URL;

  /* ============================
     State
  ============================ */

  /**
   * Status of login process
   */
  const [status, setStatus] = useState<"idle" | "success" | "error">("idle");

  /**
   * Loading indicator for login request
   */
  const [loading, setLoading] = useState(false);

  /**
   * True until GET /me has said whether someone is already signed in (a
   * reload or a bookmark of /login), so the page doesn't flash before moving on
   */
  const [checking, setChecking] = useState(!isAuthenticated);

  /* ============================
     Effects
  ============================ */

  useEffect(() => {
    if (isAuthenticated) return;

    let current = true;
    fetchCurrentUser()
      .then((signedIn) => {
        if (!current) return;
        if (signedIn) signIn(signedIn);
        else setChecking(false);
      })
      .catch(() => current && setChecking(false));

    return () => {
      current = false;
    };
  }, [isAuthenticated, signIn]);

  /* ============================
     Handlers
  ============================ */

  /**
   * Handles successful Google authentication
   * 
   * Flow:
   *  1. Send Google ID token to backend
   *  2. Backend verifies and creates session
   *  3. Fetch authenticated user data
   *  4. Store user in global context, which redirects (see the render)
   */
  const handleSuccess = async (credentialResponse: CredentialResponse) => {
    setStatus("idle");
    setLoading(true);

    try {
      const idToken = credentialResponse.credential;

      /* ============================
         Step 1: Backend Login
      ============================ */
      const loginRes = await fetch(`${API_URL}/auth/google`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ idToken }),
      });

      if (!loginRes.ok) {
        setStatus("error");
        setLoading(false);
        return;
      }

      /* ============================
         Step 2: Fetch User
      ============================ */
      const signedIn = await fetchCurrentUser();

      if (!signedIn) {
        setStatus("error");
        setLoading(false);
        return;
      }

      /* ============================
         Step 3: Update Context (the render below then redirects)
      ============================ */
      signIn(signedIn);

      setStatus("success");
    } catch (err) {
      console.error(err);
      setStatus("error");
    } finally {
      setLoading(false);
    }
  };

  /**
   * Handles Google login error
   */
  const handleError = () => {
    setStatus("error");
  };

  /* ============================
     Render
  ============================ */
  if (isAuthenticated && user) return <Navigate to={pathAfterLogin(user, from)} replace />;
  if (checking) return null;

  return (
    <div className="min-h-screen flex flex-col md:flex-row bg-page text-ink font-dawn">
      {/* Brand panel */}
      <section
        aria-label="About D.A.W.N."
        className="relative overflow-hidden md:w-1/2 flex flex-col gap-8 px-8 pt-8 pb-44 md:px-14 md:pt-14 md:pb-72 bg-nav text-on-nav"
      >
        <div>
          <div className="text-4xl font-extrabold">D.A.W.N.</div>
          <div className="mt-1 text-xl font-bold">Daily Awareness Well-being Navigator</div>
          <div className="text-base text-on-nav/70">Caregiver</div>
        </div>

        <div className="hidden md:flex flex-col gap-6 max-w-md">
          <ul className="m-0 p-0 list-none flex flex-col gap-4 text-lg">
            {FEATURES.map(({ icon: Icon, text }) => (
              <li key={text} className="flex items-center gap-3.5">
                <span
                  aria-hidden="true"
                  className="w-11 h-11 shrink-0 flex items-center justify-center rounded-full bg-on-nav/15"
                >
                  <Icon size={22} />
                </span>
                {text}
              </li>
            ))}
          </ul>
        </div>

        {/* A sun rising over the horizon, drawn in the brand's apricot and gold over the nav color */}
        <svg
          aria-hidden="true"
          viewBox="0 0 800 200"
          preserveAspectRatio="xMidYMax slice"
          className="absolute inset-x-0 bottom-0 w-full h-40 md:h-72"
        >
          <circle cx="400" cy="170" r="120" className="fill-accent/25" />
          <circle cx="400" cy="170" r="84" className="fill-accent/60" />
          <circle cx="400" cy="170" r="48" className="fill-sun" />
          <rect x="0" y="170" width="800" height="30" className="fill-nav" />
          <g className="stroke-on-nav/30" strokeWidth="2">
            <line x1="0" y1="170" x2="800" y2="170" />
            <line x1="120" y1="182" x2="680" y2="182" />
            <line x1="240" y1="194" x2="560" y2="194" />
          </g>
        </svg>
      </section>

      {/* Sign in */}
      <main className="flex-1 flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-md flex flex-col gap-6 p-8 bg-surface border-[1.5px] border-border rounded-[14px] shadow-popover">
          <div className="flex flex-col gap-2">
            <h1 className="m-0 text-3xl font-extrabold">Welcome</h1>
            <p className="m-0 text-lg text-ink-muted">
              Log in or sign up with Google.
            </p>
          </div>

          <div className="flex justify-center min-h-11">
            <GoogleLogin onSuccess={handleSuccess} onError={handleError} />
          </div>

          {status === "error" && (
            <p
              role="alert"
              className="m-0 px-3 py-2.5 rounded-lg bg-status-red-bg text-base font-bold text-status-red-text"
            >
              Login failed. Please try again.
            </p>
          )}

          {loading && (
            <p role="status" className="m-0 text-base text-ink-muted">
              Logging in… please wait.
            </p>
          )}

          <p className="m-0 pt-4 border-t border-border text-sm text-ink-muted">
            New here? Signing in with Google creates your account.
          </p>
        </div>
      </main>
    </div>
  );
};

export default Login;
