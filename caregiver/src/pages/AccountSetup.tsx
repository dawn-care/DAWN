/**
 * AccountSetup.tsx
 * 
 * User: Caregiver
 * 
 * This component handles the final onboarding step for caregivers.
 * It allows users to:
 *  - Enter a care partner name
 *  - Enter a pairing code
 *  - Accept a consent statement
 * 
 * Features:
 *  - Input validation with clear error messaging
 *  - Pairing code verification via API
 *  - Care partner selection/creation
 *  - User context refresh after successful setup
 * 
 * API interactions:
 *  - POST /users/device-pairing/verify (validate pairing code)
 *  - GET /me (refresh user data)
 *
 * A caregiver who came from Settings gets a back arrow; one with no carepartner
 * yet (straight after signing up) gets Log out instead, since nothing else is
 * open to them.
 */

import { useRef, useState } from "react";
import { ArrowLeft, LogOut } from "lucide-react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import TextField from "../components/ui/form/TextField";
import { FOCUS_RING } from "../components/ui/styles";
import { useAuth } from "../context/AuthContext";
import { fetchCurrentUser, useLogout } from "../context/session";
import { pairingErrorWording } from "./pairingErrors";

/* ============================
   Static Data
============================ */

/**
 * Pairing codes are 6 capital letters and numbers (generateCode in the server's
 * device_pairing_service.go)
 */
const PAIRING_CODE_LENGTH = 6;

/** What's typed, as a code: capitals, letters and numbers only, at most 6 */
const toPairingCode = (typed: string) =>
  typed.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, PAIRING_CODE_LENGTH);

/**
 * Base API URL from environment config
 */
const API_URL = import.meta.env.VITE_API_URL;

/**
 * The caregiver's wording for a failed verify, from the server's { "error": ... }
 * body (VerifyPairingCode in device_pairing_service.go)
 */
async function pairingError(res: Response): Promise<string> {
  const text = await res.text();
  let message = text;
  try {
    message = JSON.parse(text).error ?? text;
  } catch {
    // Not JSON (e.g. a proxy's error page); keep the text for the console
  }
  console.error(`POST /users/device-pairing/verify answered ${res.status}:`, message);
  return pairingErrorWording(message);
}

/* ============================
   Component
============================ */

export default function AccountSetupPage() {

  /* ============================
     Routing / Context
  ============================ */

  const navigate = useNavigate();
  // Settings sends its state along, so the page can offer a way back (a new caregiver arrives with none)
  const cameFromSettings = (useLocation().state as { from?: string } | null)?.from === "settings";
  const { user, setUser, setActiveCarePartnerId } = useAuth();
  const logout = useLogout();

  /* ============================
     State
  ============================ */

  const [pairingCode, setPairingCode] = useState("");
  const [consentChecked, setConsentChecked] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [carePartnerName, setCarePartnerName] = useState("");
  // True while pairing is under way, so a second click can't send the code again. The ref
  // blocks a click before the disabled button has rendered; the state draws it
  const [submitting, setSubmitting] = useState(false);
  const submittingRef = useRef(false);

  /* ============================
     Handlers
  ============================ */

  /**
   * Handles account setup submission
   * - Validates inputs
   * - Verifies pairing code
   * - Refreshes user data
   */
  const handleSubmit = async () => {
    if (submittingRef.current) return;

    // Ensure name is typed
    if (!carePartnerName.trim()) {
      setSubmitError("Please enter who this is for.");
      return;
    }

    // Ensure pairing code is entered
    if (!pairingCode) {
      setSubmitError("Pairing code is required.");
      return;
    }
    if (pairingCode.length !== PAIRING_CODE_LENGTH) {
      setSubmitError(`The pairing code is ${PAIRING_CODE_LENGTH} letters and numbers.`);
      return;
    }

    // Ensure checkbox is clicked
    if (!consentChecked) {
      setSubmitError("You must agree to the consent statement to proceed.");
      return;
    }

    // Get rid of "user is possibly null" error
    if (!user?.id) {
      setSubmitError("User is not logged in. Please log in again.");
      return;
    }

    setSubmitError("");
    submittingRef.current = true;
    setSubmitting(true);

    try {
      // Verify that the pairing code matches one in the pairing code table
      const res = await fetch(`${API_URL}/users/device-pairing/verify`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          code: pairingCode,
          caregiver_id: user.id, 
          carepartner_name: carePartnerName.trim(),
        }),
      });

      if (!res.ok) {
        setSubmitError(await pairingError(res));
        return;
      }

      const verifyData: { carepartner_id: number } = await res.json();

      // Set the active carepartner from the pairing response
      setActiveCarePartnerId(verifyData.carepartner_id);

      // The code is used up now, so go on even if the refresh fails: add the new
      // carepartner here instead (the next reload loads the real list)
      try {
        const refreshedUser = await fetchCurrentUser();
        if (!refreshedUser) throw new Error("GET /me answered 401");
        setUser(refreshedUser);
      } catch (err) {
        console.error("Pairing worked, but refreshing the caregiver failed:", err);
        setUser((currentUser) =>
          currentUser
            ? {
                ...currentUser,
                carepartners: [
                  ...(currentUser.carepartners ?? []),
                  { id: verifyData.carepartner_id, name: carePartnerName.trim() },
                ],
              }
            : currentUser
        );
      }

      // Pairing succeeded
      navigate("/");
    } catch (err) {
      console.error("Pairing failed:", err);
      setSubmitError("Couldn't reach the server. Check your connection and try again.");
    } finally {
      submittingRef.current = false;
      setSubmitting(false);
    }
  };

  /* ============================
     Render
  ============================ */

  return (
    <div className="min-h-screen flex flex-col items-center px-6 py-10 bg-page text-ink font-dawn">
      <div className="w-full max-w-xl flex flex-col gap-6">
        <header className="flex items-center gap-3">
          {cameFromSettings && (
            <Link
              to="/settings"
              aria-label="Back to settings"
              className={`w-11 h-11 shrink-0 flex items-center justify-center rounded-full text-ink hover:bg-border/60 ${FOCUS_RING}`}
            >
              <ArrowLeft size={24} aria-hidden="true" />
            </Link>
          )}
          <h1 className="m-0 text-3xl font-extrabold">Pair a new device</h1>
          {!cameFromSettings && (
            <button
              type="button"
              onClick={logout}
              className={`ml-auto h-11 px-3.5 flex items-center gap-2 rounded-[10px] border-0 bg-transparent text-base font-bold text-ink cursor-pointer hover:bg-border/60 ${FOCUS_RING}`}
            >
              <LogOut size={18} aria-hidden="true" />
              Log out
            </button>
          )}
        </header>

        <div className="flex flex-col gap-6 p-8 bg-surface border-[1.5px] border-border rounded-[14px]">
          <p className="m-0 text-lg text-ink-muted">
            Connect your carepartner's tablet to your account so you can choose what it shows.
          </p>

          {/* ============================
              Care Partner Name
          ============================ */}
          <TextField
            label="Who is this for?"
            value={carePartnerName}
            onChange={(e) => setCarePartnerName(e.target.value)}
            placeholder="Enter name"
          />

          {/* ============================
              Pairing Code
          ============================ */}
          <TextField
            label="Pairing code"
            hint="Shown on your carepartner's tablet"
            value={pairingCode}
            // Codes are always capitals, so typing in lowercase still shows (and sends) them that
            // way; anything but letters and numbers, and anything past 6, is left out
            onChange={(e) => setPairingCode(toPairingCode(e.target.value))}
            maxLength={PAIRING_CODE_LENGTH}
            autoCapitalize="characters"
            autoCorrect="off"
            autoComplete="off"
            spellCheck={false}
            className="uppercase tracking-[0.3em] text-xl font-bold"
            placeholder="ABC123"
          />

          {/* ============================
              Consent Statement
          ============================ */}
          <div className="flex flex-col gap-2.5">
            <div className="text-base font-bold">Consent statement</div>
            <div className="flex flex-col gap-2 p-4 bg-page border-[1.5px] border-border rounded-[10px] max-h-48 overflow-y-auto text-[15px] leading-snug">
              <p className="m-0">
                By checking this box, you acknowledge that your photos and videos will be
                stored on the D.A.W.N. server. Your data will not be sold to third parties.
              </p>
              <p className="m-0">
                Please make sure you have discussed this with your care partner before
                proceeding. You must agree to this statement to complete account setup.
              </p>
            </div>

            <label htmlFor="consentCheckbox" className="flex items-center gap-3 min-h-11 text-base cursor-pointer">
              <input
                type="checkbox"
                id="consentCheckbox"
                checked={consentChecked}
                onChange={(e) => setConsentChecked(e.target.checked)}
                className={`w-5 h-5 shrink-0 accent-accent cursor-pointer ${FOCUS_RING}`}
              />
              I agree to the consent statement
            </label>
          </div>

          {/* ============================
              Error Message
          ============================ */}
          {submitError && (
            <p role="alert" className="m-0 px-3 py-2.5 rounded-lg bg-status-red-bg text-base font-bold text-status-red-text">
              {submitError}
            </p>
          )}

          {/* ============================
              Submit Button
          ============================ */}
          <button
            type="button"
            onClick={handleSubmit}
            disabled={submitting}
            className={`h-12 w-full rounded-[10px] border-2 border-accent bg-accent text-base font-bold text-on-accent cursor-pointer hover:brightness-95 disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`}
          >
            {submitting ? "Pairing…" : "Pair device"}
          </button>
        </div>
      </div>
    </div>
  );
}
