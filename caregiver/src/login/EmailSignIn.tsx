/**
 * EmailSignIn.tsx
 *
 * User: Caregiver
 *
 * Sign in with an emailed 6-digit code, for caregivers without a Google account.
 * Two steps: enter the email address, then the code that was sent to it.
 *
 *  - "Send a new code" waits RESEND_SECONDS between sends
 *  - Five wrong codes burn the code (the server says so), and the field is cleared
 *  - A code is not needed to be typed exactly: spaces and dashes are dropped
 *
 * Once the server has accepted the code it has started the session, and
 * `onSignedIn` loads who signed in; the Login page then moves on.
 */

import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";
import TextField from "../components/ui/form/TextField";
import { ApiError } from "../media/api";
import { FOCUS_RING } from "../components/ui/styles";
import { PRIMARY_BUTTON } from "../media/styles";
import {
  CODE_LENGTH,
  GENERIC_ERROR,
  RESEND_SECONDS,
  friendlyError,
  requestEmailCode,
  verifyEmailCode,
} from "./api";

interface EmailSignInProps {
  /** Loads the caregiver the server just signed in. Resolves false if nobody is signed in */
  onSignedIn: () => Promise<boolean>;
}

const ALERT_CLASSES =
  "m-0 px-3 py-2.5 rounded-lg bg-status-red-bg text-base font-bold text-status-red-text";

// The two buttons under Sign in. They share a row when the card is wide enough for both and
// stack when it isn't (a phone): nowrap makes each as wide as its label, and flex-1 then shares
// what is left equally.
const ROW_BUTTON = `h-11 px-3 flex-1 whitespace-nowrap rounded-[10px] border-2 border-ink-muted bg-surface text-sm font-bold text-ink cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`;

const EmailSignIn = ({ onSignedIn }: EmailSignInProps) => {
  const [step, setStep] = useState<"email" | "code">("email");
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  // Seconds until a new code may be sent
  const [wait, setWait] = useState(0);
  const codeInput = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (wait <= 0) return;
    const timer = setTimeout(() => setWait((seconds) => seconds - 1), 1000);
    return () => clearTimeout(timer);
  }, [wait]);

  useEffect(() => {
    if (step === "code") codeInput.current?.focus();
  }, [step]);

  const sendCode = async (event?: FormEvent) => {
    event?.preventDefault();
    if (busy || !email.trim()) return;
    const resend = step === "code";
    setBusy(true);
    setError(null);
    setNotice(null);
    try {
      await requestEmailCode(email.trim());
      setCode("");
      setWait(RESEND_SECONDS);
      setStep("code");
      if (resend) setNotice("We sent a new code. Only the newest one works.");
    } catch (err) {
      setError(friendlyError(err));
    } finally {
      setBusy(false);
    }
  };

  const signIn = async (event: FormEvent) => {
    event.preventDefault();
    if (busy || code.length !== CODE_LENGTH) return;
    setBusy(true);
    setError(null);
    setNotice(null);
    try {
      await verifyEmailCode(email.trim(), code);
      if (!(await onSignedIn())) setError(GENERIC_ERROR);
    } catch (err) {
      setError(friendlyError(err));
      // The code is burned, so there is nothing left to correct
      if (err instanceof ApiError && err.code === "TOO_MANY_ATTEMPTS")
        setCode("");
    } finally {
      setBusy(false);
    }
  };

  const useDifferentEmail = () => {
    setStep("email");
    setCode("");
    setError(null);
    setNotice(null);
    setWait(0);
  };

  return (
    <div className="flex flex-col gap-4">
      {step === "email" ? (
        <form onSubmit={sendCode} noValidate className="flex flex-col gap-3">
          <TextField
            label="Email address"
            type="email"
            autoComplete="email"
            inputMode="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
          />
          <button
            type="submit"
            disabled={busy || !email.trim()}
            className={PRIMARY_BUTTON}
          >
            {busy ? "Sending…" : "Send Code"}
          </button>
        </form>
      ) : (
        <form onSubmit={signIn} noValidate className="flex flex-col gap-3">
          <p className="m-0 text-base text-ink">
            We sent a {CODE_LENGTH}-digit code to{" "}
            <strong className="break-all">{email.trim()}</strong>. It can take a
            minute to arrive. If you don&apos;t see it, check your spam or junk
            folder.
          </p>
          <TextField
            ref={codeInput}
            label={`${CODE_LENGTH}-digit code`}
            autoComplete="one-time-code"
            inputMode="numeric"
            maxLength={CODE_LENGTH + 4}
            value={code}
            onChange={(event) =>
              setCode(
                event.target.value.replace(/\D/g, "").slice(0, CODE_LENGTH),
              )
            }
          />
          <button
            type="submit"
            disabled={busy || code.length !== CODE_LENGTH}
            className={PRIMARY_BUTTON}
          >
            {busy ? "Signing in…" : "Sign in"}
          </button>
          <div className="flex flex-wrap gap-3">
            <button
              type="button"
              onClick={() => sendCode()}
              disabled={busy || wait > 0}
              // As wide as its longest label, "(30s)", so the row doesn't reflow as the count falls
              className={`${ROW_BUTTON} min-w-[11.5rem]`}
            >
              {wait > 0 ? `Send a new code (${wait}s)` : "Send a new code"}
            </button>
            <button
              type="button"
              onClick={useDifferentEmail}
              disabled={busy}
              className={ROW_BUTTON}
            >
              Use a different email
            </button>
          </div>
        </form>
      )}

      {notice && (
        <p role="status" className="m-0 text-base text-ink-muted">
          {notice}
        </p>
      )}
      {error && (
        <p role="alert" className={ALERT_CLASSES}>
          {error}
        </p>
      )}
    </div>
  );
};

export default EmailSignIn;
