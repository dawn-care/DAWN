/**
 * RequireLogin.tsx
 *
 * User: Caregiver
 *
 * Layout route that sends anyone who isn't logged in to /login, passing the
 * page they wanted so Login can return them there. The session is a cookie the
 * browser keeps, so after a reload the app asks the server who is signed in
 * (GET /me) before showing anything. A signed-in answer is kept in AuthContext,
 * so it asks once per page load. useSignIn also picks the carepartner to manage:
 * the saved one if it is still theirs, otherwise their first.
 *
 * Only a 401 means logged out. Any other failure (server down, a 500, no
 * network) shows a message with Try again, so a signed-in caregiver isn't sent
 * to sign in with Google again just because the server didn't answer.
 */

import { useEffect, useState } from "react";
import { Navigate, Outlet, useLocation } from "react-router-dom";
import { FOCUS_RING } from "./ui/styles";
import { useAuth } from "../context/AuthContext";
import { fetchCurrentUser, useSignIn } from "../context/session";

export default function RequireLogin() {
  const { isAuthenticated } = useAuth();
  const signIn = useSignIn();
  const location = useLocation();
  // What the server said: still waiting, nobody signed in, or no usable answer
  const [answer, setAnswer] = useState<"checking" | "loggedOut" | "failed">("checking");
  // Bumped by Try again to ask once more
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    if (isAuthenticated) return;

    let current = true;
    fetchCurrentUser()
      .then((user) => {
        if (!current) return;
        if (user) signIn(user);
        else setAnswer("loggedOut");
      })
      .catch((err) => {
        if (!current) return;
        console.error("Checking who is signed in failed:", err);
        setAnswer("failed");
      });

    return () => {
      current = false;
    };
  }, [isAuthenticated, signIn, attempt]);

  if (isAuthenticated) return <Outlet />;
  if (answer === "loggedOut") {
    return <Navigate to="/login" replace state={{ from: location.pathname + location.search + location.hash }} />;
  }
  if (answer === "failed") {
    return (
      <div className="min-h-screen flex items-center justify-center px-6 bg-page text-ink font-dawn">
        <div
          role="alert"
          className="w-full max-w-md flex flex-col items-center gap-4 p-8 bg-surface border-[1.5px] border-border rounded-[14px] text-center"
        >
          <p className="m-0 text-xl font-bold">Couldn't reach the server.</p>
          <p className="m-0 text-base text-ink-muted">Check your connection and try again.</p>
          <button
            type="button"
            onClick={() => {
              setAnswer("checking");
              setAttempt((n) => n + 1);
            }}
            className={`h-11 px-[18px] rounded-[10px] border-2 border-accent bg-accent hover:brightness-95 text-on-accent text-base font-bold cursor-pointer ${FOCUS_RING}`}
          >
            Try again
          </button>
        </div>
      </div>
    );
  }
  // Nothing to show until we know; the answer takes a moment
  return null;
}
