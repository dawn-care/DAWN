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
 */

import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

/* ============================
   Static Data
============================ */

/**
 * Suggested names for care partner input
 */
const SUGGESTIONS = ["Mom", "Dad", "Grandma", "Grandpa", "Friend"];

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

/* ============================
   Component
============================ */

export default function AccountSetupPage() {

  /* ============================
     Routing / Context
  ============================ */

  const navigate = useNavigate();
  const { user, setUser, setActiveCarePartnerId } = useAuth();

  /* ============================
     State
  ============================ */

  const [pairingCode, setPairingCode] = useState("");
  const [consentChecked, setConsentChecked] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [carePartnerName, setCarePartnerName] = useState("");

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

    try {
      // Verify that the pairing code matches one in the pairing code table
      const res = await fetch(`${API_URL}/users/device-pairing/verify`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          code: pairingCode,
          caregiver_id: user.id, 
          carepartner_name: carePartnerName,
        }),
      });

      if (!res.ok) {
        const errText = await res.text();
        setSubmitError(errText || "Failed to verify pairing code. Please try again.");
        return;
      }

      const verifyData = await res.json();

      // Set the active carepartner from the pairing response
      setActiveCarePartnerId(verifyData.carepartner_id);

      const meRes = await fetch(`${API_URL}/me`, {
        credentials: "include",
      });

      if (!meRes.ok) {
        setSubmitError("Pairing succeeded, but failed to refresh account data.");
        return;
      }

      const refreshedUser = await meRes.json();
      setUser(refreshedUser);

      // Pairing succeeded
      navigate("/");
    } catch (err) {
      console.error(err);
      setSubmitError("Something went wrong. Please try again.");
    }
  };

  /* ============================
     Render
  ============================ */

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">
        Pair New Device
      </h1>

      <div className="border p-8 max-w-3xl w-full rounded-lg bg-surface shadow">

        {/* ============================
            Care Partner Name
        ============================ */}
        <div className="mb-6">
          <label className="text-lg font-medium">Who is this for?</label>

          <input
            type="text"
            value={carePartnerName}
            onChange={(e) => setCarePartnerName(e.target.value)}
            className="w-full mt-2 p-3 border rounded bg-surface text-ink"
            placeholder="Enter name"
          />

          {/* Suggestions */}
          <div className="flex gap-2 mt-2 flex-wrap">
            {SUGGESTIONS.map((s) => (
              <button
                key={s}
                onClick={() => setCarePartnerName(s)}
                className="px-3 py-1 bg-border text-ink-muted rounded-full text-sm hover:brightness-95"
              >
                {s}
              </button>
            ))}
          </div>
        </div>

        {/* ============================
            Pairing Code
        ============================ */}
        <div className="mb-6">
          <label className="text-lg font-medium">Enter Care Partner Pairing Code:</label>
          <input
            type="text"
            value={pairingCode}
            // Codes are always capitals, so typing in lowercase still shows (and sends) them that
            // way; anything but letters and numbers, and anything past 6, is left out
            onChange={(e) => setPairingCode(toPairingCode(e.target.value))}
            maxLength={PAIRING_CODE_LENGTH}
            autoCapitalize="characters"
            autoCorrect="off"
            autoComplete="off"
            spellCheck={false}
            className="w-full mt-2 p-3 border rounded bg-surface text-ink uppercase"
            placeholder="Enter pairing code"
          />
        </div>

        {/* ============================
            Consent Statement
        ============================ */}
        <div className="mb-6">
          <label className="text-lg font-medium">Consent Statement:</label>
          <div className="mt-2 p-4 border rounded bg-page max-h-48 overflow-y-auto text-sm text-ink">
            <p>
              By checking this box, you acknowledge that your photos and videos will be
              stored on the StartUp server. Your data will not be sold to third parties.
            </p>
            <p className="mt-2">
              Please make sure you have discussed this with your care partner before
              proceeding. You must agree to this statement to complete account setup.
            </p>
          </div>

          <div className="mt-2 flex items-center gap-2">
            <input
              type="checkbox"
              id="consentCheckbox"
              checked={consentChecked}
              onChange={(e) => setConsentChecked(e.target.checked)}
            />
            <label htmlFor="consentCheckbox" className="text-sm text-ink">
              I agree to the consent statement
            </label>
          </div>
        </div>

        {/* ============================
            Error Message
        ============================ */}
        {submitError && (
          <p className="text-status-red-text mb-4">{submitError}</p>
        )}

        {/* ============================
            Submit Button
        ============================ */}
        <button
          onClick={handleSubmit}
          className="mt-4 w-full bg-accent hover:brightness-95 text-on-accent font-semibold py-3 rounded"
        >
          Pair Device
        </button>
      </div>
    </div>
  );
}