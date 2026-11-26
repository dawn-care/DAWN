/**
 * api.ts
 *
 * User: Caregiver
 *
 * The calls for signing in with an emailed code (POST /api/auth/email/code and
 * /api/auth/email/verify). A successful verify starts the session: the server
 * sets the session cookie, as it does for Google, so these calls send cookies.
 */

import { ApiError } from "../media/api";

const API_URL = import.meta.env.VITE_API_URL;

/** The code is always this many digits */
export const CODE_LENGTH = 6;
/** How long "Send a new code" stays disabled, so it isn't pressed again and again */
export const RESEND_SECONDS = 30;

export const GENERIC_ERROR = "Something went wrong. Please try again.";

/** Wording for each `code` the server can answer with */
const ERROR_MESSAGES: Readonly<Record<string, string>> = {
  INVALID_EMAIL: "Please enter a valid email address.",
  INVALID_REQUEST: "Something went wrong. Please reload the page and try again.",
  RATE_LIMITED: "Too many codes requested. Try again later.",
  EMAIL_SEND_FAILED:
    "We couldn't send the email. Check the address, or try again in a few minutes.",
  CODE_REJECTED: "That code is incorrect or has expired.",
  TOO_MANY_ATTEMPTS:
    "Too many incorrect attempts. Send a new code and try again.",
  TOO_MANY_GUESSES:
    "Too many incorrect codes from this network. Try again later.",
  CROSS_ORIGIN_BLOCKED:
    "This page can't sign you in. Open D.A.W.N. from its own address and try again.",
  NETWORK_ERROR:
    "Couldn't reach the server. Check your connection and try again.",
};

/**
 * The rate limits: the server's own message says how long to wait ("Try again
 * in 15 minutes."), so it is shown instead of the general wording above
 */
const SERVER_WORDED = new Set(["RATE_LIMITED", "TOO_MANY_GUESSES"]);

async function post(path: string, body: Record<string, string>): Promise<void> {
  let res: Response;
  try {
    res = await fetch(`${API_URL}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(body),
    });
  } catch (err) {
    console.error(`POST ${path} could not reach the server`, err);
    throw new ApiError(0, "NETWORK_ERROR", "could not reach the server");
  }
  if (res.ok) return;

  const error = await res.json().catch(() => null);
  const code = error?.code ?? "INTERNAL_ERROR";
  if (!ERROR_MESSAGES[code])
    console.error(`POST ${path} failed`, res.status, error);
  throw new ApiError(res.status, code, error?.error ?? res.statusText);
}

/** Emails a sign-in code to the address */
export const requestEmailCode = (email: string) =>
  post("/auth/email/code", { email });

/** Checks the code and, if it is right, signs in (the server sets the session cookie) */
export const verifyEmailCode = (email: string, code: string) =>
  post("/auth/email/verify", { email, code });

/** Wording to show the caregiver for anything thrown by a call here */
export function friendlyError(err: unknown): string {
  if (!(err instanceof ApiError)) return GENERIC_ERROR;
  if (SERVER_WORDED.has(err.code) && err.message) return err.message;
  return ERROR_MESSAGES[err.code] ?? GENERIC_ERROR;
}
