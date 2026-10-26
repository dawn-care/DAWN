/**
 * Settings.tsx
 *
 * User: Caregiver
 *
 * This component allows users to manage account and care partner settings.
 *
 * Responsibilities:
 *  - Change the caregiver's name (its own Save button; a success shows a toast, an error stays under it)
 *  - Switch the active carepartner (applies as soon as it is picked)
 *  - Log out
 *  - Go to the pairing page to pair another tablet
 *
 * The caregiver comes from AuthContext, which RequireLogin has already loaded.
 */

import { useState } from "react";
import { ArrowLeft, LogOut, Smartphone } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import SelectField from "../components/ui/form/SelectField";
import TextField from "../components/ui/form/TextField";
import { FOCUS_RING } from "../components/ui/styles";
import { useToast } from "../components/ui/toastContext";
import { useAuth } from "../context/AuthContext";
import { useLogout } from "../context/session";

// The same two buttons as the rest of the app: one filled (the main action), one outlined (everything else)
const PRIMARY_BUTTON = `h-11 px-[18px] flex items-center justify-center gap-2 rounded-[10px] border-2 border-accent bg-accent hover:brightness-95 text-on-accent text-base font-bold cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`;

/**
 * Base API URL
 */
const API_URL = import.meta.env.VITE_API_URL;

/** The caregiver's wording for a failed save, from the server's { "error": ... } body */
async function saveNameError(res: Response): Promise<string> {
  const text = await res.text();
  let message = text;
  try {
    message = JSON.parse(text).error ?? text;
  } catch {
    // Not JSON (e.g. a proxy's error page); keep the text for the console
  }
  console.error(`PUT /caregivers answered ${res.status}:`, message);
  if (message === "name cannot be empty") return "Enter your name before saving.";
  return "Couldn't save your name. Please try again.";
}

/* ============================
   Component
============================ */

/**
 * SettingsPage Component
 *
 * Provides account settings, carepartner configuration,
 * and session management actions.
 */
export default function SettingsPage() {

  /* ============================
     Routing / Context
  ============================ */

  const navigate = useNavigate();
  const toast = useToast();
  const logout = useLogout();

  const { user, setUser, activeCarePartnerId, setActiveCarePartnerId } = useAuth();

  const carepartners = user?.carepartners ?? [];

  /* ============================
     State
  ============================ */

  /**
   * The name as typed
   */
  const [name, setName] = useState(user?.name ?? "");

  /**
   * True while Save is under way
   */
  const [saving, setSaving] = useState(false);

  /**
   * Why the last save failed; stays until the next save, unlike a toast
   */
  const [saveError, setSaveError] = useState<string | null>(null);

  /* ============================
     Derived Values
  ============================ */

  /**
   * Option labels; a carepartner who shares a name with another gets their ID
   * too, so the two can be told apart
   */
  const carepartnerOptions = carepartners.map((carepartner) => {
    const shared = carepartners.filter((other) => other.name === carepartner.name).length > 1;
    return {
      value: String(carepartner.id),
      label: shared ? `${carepartner.name} (ID ${carepartner.id})` : carepartner.name,
    };
  });

  /* ============================
     Handlers
  ============================ */

  /**
   * Saves the caregiver's name
   */
  const handleSave = async () => {
    if (!user) return;

    const trimmed = name.trim();
    if (!trimmed) {
      setSaveError("Enter your name before saving.");
      return;
    }

    setSaving(true);
    setSaveError(null);
    try {
      const res = await fetch(`${API_URL}/caregivers`, {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        // The server sets the role from this too, so send the one it gave us
        body: JSON.stringify({ id: user.id, name: trimmed, role: user.role ?? "" }),
      });

      if (!res.ok) {
        setSaveError(await saveNameError(res));
        return;
      }

      setUser((currentUser) => (currentUser ? { ...currentUser, name: trimmed } : currentUser));
      setName(trimmed);
      toast.show("Name saved.");
    } catch (err) {
      console.error("Saving the name failed:", err);
      setSaveError("Couldn't reach the server. Check your connection and try again.");
    } finally {
      setSaving(false);
    }
  };

  /* ============================
     Render
  ============================ */
  return (
    <div className="min-h-screen flex flex-col items-center px-6 py-10 bg-page text-ink font-dawn">
      <div className="w-full max-w-xl flex flex-col gap-6">
        <header className="flex items-center gap-3">
          <Link
            to="/"
            aria-label="Back to the carepartner screen"
            className={`w-11 h-11 shrink-0 flex items-center justify-center rounded-full text-ink hover:bg-border/60 ${FOCUS_RING}`}
          >
            <ArrowLeft size={24} aria-hidden="true" />
          </Link>
          <h1 className="m-0 text-3xl font-extrabold">Settings</h1>
        </header>

        {/* Account */}
        <section
          aria-labelledby="settings-account"
          className="flex flex-col gap-5 p-6 bg-surface border-[1.5px] border-border rounded-[14px]"
        >
          <h2 id="settings-account" className="m-0 text-xl font-extrabold">
            Your account
          </h2>
          <TextField
            label="Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleSave();
            }}
          />
          {saveError && (
            <p role="alert" className="m-0 px-3 py-2.5 rounded-lg bg-status-red-bg text-base font-bold text-status-red-text">
              {saveError}
            </p>
          )}
          <div className="flex flex-wrap items-center justify-between gap-3">
            <button type="button" onClick={handleSave} disabled={saving} className={PRIMARY_BUTTON}>
              {saving ? "Saving…" : "Save name"}
            </button>
            <button type="button" onClick={logout} className={PRIMARY_BUTTON}>
              <LogOut size={18} aria-hidden="true" />
              Log out
            </button>
          </div>
        </section>

        {/* Carepartner */}
        <section
          aria-labelledby="settings-carepartner"
          className="flex flex-col gap-5 p-6 bg-surface border-[1.5px] border-border rounded-[14px]"
        >
          <h2 id="settings-carepartner" className="m-0 text-xl font-extrabold">
            Your carepartner
          </h2>
          <SelectField
            label="Active carepartner"
            value={activeCarePartnerId ?? ""}
            onChange={(e) => setActiveCarePartnerId(Number(e.target.value))}
            disabled={carepartners.length === 0}
            options={
              carepartners.length === 0
                ? [{ value: "", label: "No connected carepartners" }]
                : carepartnerOptions
            }
          />
          <button
            type="button"
            onClick={() => navigate("/setup", { state: { from: "settings" } })}
            className={`${PRIMARY_BUTTON} self-start`}
          >
            <Smartphone size={18} aria-hidden="true" />
            Pair a new device
          </button>
        </section>
      </div>
    </div>
  );
}
