/**
 * SelectCarePartner.tsx
 * 
 * User: Caregiver
 * 
 * This component allows a logged-in caregiver to select an active care partner.
 * 
 * Responsibilities:
 *  - Display list of available carepartners from authenticated user
 *  - Allow user to select one carepartner
 *  - Store selection in AuthContext (and persisted storage)
 *  - Redirect user to the home page (`/`) after selection
 */

import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

/* ============================
   Component
============================ */

/**
 * SelectCarePartnerPage Component
 * 
 * Used when a user has multiple carepartners and must choose one
 * before accessing the main application.
 */
export default function SelectCarePartnerPage() {

  /* ============================
     Routing / Context
  ============================ */

  const navigate = useNavigate();
  const { user, setActiveCarePartnerId } = useAuth();

  /* ============================
     State
  ============================ */

  /**
   * Currently selected carepartner ID (string for select input binding)
   */
  const [selectedId, setSelectedId] = useState("");

  /* ============================
     Handlers
  ============================ */

  /**
   * Handles continuation after selecting a carepartner
   */
  const handleContinue = () => {
    if (!selectedId) return;

    // store selection via context (which syncs to localStorage)
    setActiveCarePartnerId(Number(selectedId));

    // Navigate to main application
    navigate("/");
  };

  /* ============================
     Render
  ============================ */
  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">
        Who is this for?
      </h1>

      <div className="border p-8 max-w-xl w-full rounded-lg bg-surface shadow">

        {/* Carepartner Selection Dropdown */}
        <select
          value={selectedId}
          onChange={(e) => setSelectedId(e.target.value)}
          className="w-full p-3 border rounded bg-surface text-ink"
        >
          <option value="">Select care partner</option>
          {user?.carepartners?.map((cp) => (
            <option key={cp.id} value={cp.id}>
              {cp.name}
            </option>
          ))}
        </select>

        {/* Info text */}
        <p className="mt-4 text-sm text-ink-muted italic text-center">
          You will be able to switch this in{" "}
          <span className="italic font-medium">Agenda</span>
        </p>

        {/* Continue Button */}
        <button
          onClick={handleContinue}
          className="mt-6 w-full bg-accent hover:brightness-95 text-on-accent font-semibold py-3 rounded"
        >
          Continue
        </button>
      </div>
    </div>
  );
}