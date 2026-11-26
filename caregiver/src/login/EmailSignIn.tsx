/**
 * EmailSignIn.tsx
 *
 * User: Caregiver
 *
 * Sign in with an emailed 6-digit code, for caregivers without a Google account.
 * Two steps: enter the email address, then the code that was sent to it.
 *
 *  - The code step checks the code against the address it was sent to, which
 *    can't be edited while the code is being sent
 *  - "Already have a code? Enter here" goes to the code step without sending one, for a
 *    caregiver who reloaded the page or reached the sending limit with a code
 *    that still works
 *  - "Send a new code" waits RESEND_SECONDS between sends
 *  - Five wrong codes burn the code (the server says so), and the field is cleared
 *  - Anything typed in the code field that is not a digit is dropped
 *
 * Once the server has accepted the code it has started the session and used
 * the code up, and `onSignedIn` loads who signed in; the Login page then moves
 * on. If that load fails, the caregiver is told to reload, not to re-enter the
 * code, which no longer works.
 */

import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";
import TextField from "../components/ui/form/TextField";
import { ApiError } from "../media/api";
import { FOCUS_RING } from "../components/ui/styles";
import { PRIMARY_BUTTON } from "../media/styles";
import {
  CODE_LENGTH,
  RESEND_SECONDS,
  friendlyError,
  requestEmailCode,
  verifyEmailCode,
} from "./api";

interface EmailSignInProps {
  /** Loads the caregiver the server just signed in. Resolves false if nobody is signed in */
  onSignedIn: () => Promise<boolean>;
}

// A right code has been used up and the session cookie is set, so reloading
// signs the caregiver in; entering the code again would only be refused
const SIGNED_IN_BUT_NOT_LOADED =
  "You're signed in, but we couldn't load your account. Reload the page to continue.";

const ALERT_CLASSES =
  "m-0 px-3 py-2.5 rounded-lg bg-status-red-bg text-base font-bold text-status-red-text";

// The two buttons under Sign in. They share a row when the card is wide enough for both and
// stack when it isn't (a phone): nowrap makes each as wide as its label, and flex-1 then shares
// what is left equally.
const ROW_BUTTON = `h-11 px-3 flex-1 whitespace-nowrap rounded-[10px] border-2 border-ink-muted bg-surface text-sm font-bold text-ink cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`;

const ALREADY_HAVE_CODE = "Already have a code? Enter here";

// A text link under Send Code, so Send Code stays the one obvious next step. A button that looks
// like a link, since it changes the step rather than the page; underlined and 44px tall so it
// reads as clickable and is easy to hit.
const TEXT_LINK_BUTTON = `self-center min-h-11 px-2 rounded-md bg-transparent border-0 text-base font-bold text-link underline underline-offset-4 cursor-pointer hover:no-underline disabled:opacity-60 disabled:no-underline disabled:cursor-default ${FOCUS_RING}`;

const EmailSignIn = ({ onSignedIn }: EmailSignInProps) => {
  const [step, setStep] = useState<"email" | "code">("email");
  // What is typed in the email field
  const [email, setEmail] = useState("");
  // The address the code step is for: the one the code was sent to, or the one
  // typed before "Already have a code? Enter here"
  const [sentTo, setSentTo] = useState("");
  // False when the code step was reached without sending a code on this visit
  const [codeSentHere, setCodeSentHere] = useState(false);
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
    const resend = step === "code";
    // Read once, so the code step is for the address the code went to
    const address = resend ? sentTo : email.trim();
    if (busy || !address) return;
    setBusy(true);
    setError(null);
    setNotice(null);
    try {
      await requestEmailCode(address);
      setSentTo(address);
      setCodeSentHere(true);
      setCode("");
      setWait(RESEND_SECONDS);
      setStep("code");
      if (resend) setNotice("We sent a new code. Only the newest one works.");
    } catch (err) {
      setError(friendlyError(err));
      // An earlier code may still work, so point to where it can be entered
      if (!resend && err instanceof ApiError && err.code === "RATE_LIMITED")
        setNotice(`If you already have a code, choose "${ALREADY_HAVE_CODE}" above.`);
    } finally {
      setBusy(false);
    }
  };

  const enterExistingCode = () => {
    const address = email.trim();
    if (busy || !address) return;
    setSentTo(address);
    setCodeSentHere(false);
    setCode("");
    setError(null);
    setNotice(null);
    setWait(0);
    setStep("code");
  };

  const signIn = async (event: FormEvent) => {
    event.preventDefault();
    if (busy || code.length !== CODE_LENGTH) return;
    setBusy(true);
    setError(null);
    setNotice(null);
    try {
      await verifyEmailCode(sentTo, code);
    } catch (err) {
      setError(friendlyError(err));
      // The code is burned, so there is nothing left to correct
      if (err instanceof ApiError && err.code === "TOO_MANY_ATTEMPTS")
        setCode("");
      setBusy(false);
      return;
    }
    try {
      if (!(await onSignedIn())) {
        console.error("Signed in with an emailed code, but GET /me says nobody is signed in");
        setError(SIGNED_IN_BUT_NOT_LOADED);
      }
    } catch (err) {
      console.error("Signed in with an emailed code, but loading the caregiver failed", err);
      setError(SIGNED_IN_BUT_NOT_LOADED);
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
            // Read-only while sending, so the address can't change under the request
            readOnly={busy}
            onChange={(event) => setEmail(event.target.value)}
          />
          <button
            type="submit"
            disabled={busy || !email.trim()}
            className={PRIMARY_BUTTON}
          >
            {busy ? "Sending…" : "Send Code"}
          </button>
          <button
            type="button"
            onClick={enterExistingCode}
            disabled={busy || !email.trim()}
            className={TEXT_LINK_BUTTON}
          >
            {ALREADY_HAVE_CODE}
          </button>
        </form>
      ) : (
        <form onSubmit={signIn} noValidate className="flex flex-col gap-3">
          {codeSentHere ? (
            <p className="m-0 text-base text-ink">
              We sent a {CODE_LENGTH}-digit code to{" "}
              <strong className="break-all">{sentTo}</strong>. It can take a
              minute to arrive. If you don&apos;t see it, check your spam or
              junk folder.
            </p>
          ) : (
            <p className="m-0 text-base text-ink">
              Enter the {CODE_LENGTH}-digit code we emailed to{" "}
              <strong className="break-all">{sentTo}</strong>. Only the newest
              code works, and a code stops working after 10 minutes.
            </p>
          )}
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
