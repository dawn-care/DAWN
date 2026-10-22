/**
 * Settings.tsx
 * 
 * User: Caregiver
 * 
 * This component allows users to manage account and care partner settings.
 * 
 * Responsibilities:
 *  - View and update caregiver profile information
 *  - View active carepartner details
 *  - Logout user and clear session
 *  - Navigate to device repair/setup flow
 */

import { useEffect, useState } from "react";
import { ArrowLeft, LogOut, Smartphone } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import SelectField from "../components/ui/form/SelectField";
import TextField from "../components/ui/form/TextField";
import { FOCUS_RING } from "../components/ui/styles";
import { useToast } from "../components/ui/toastContext";
import { useAuth } from "../context/AuthContext";

// The same two buttons as the rest of the app: one filled (the main action), one outlined (everything else)
const PRIMARY_BUTTON = `h-11 px-[18px] flex items-center justify-center gap-2 rounded-[10px] border-2 border-accent bg-accent hover:brightness-95 text-on-accent text-base font-bold cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`;
const SECONDARY_BUTTON = `h-11 px-[18px] flex items-center justify-center gap-2 rounded-[10px] border-2 border-ink-muted bg-surface hover:bg-page text-ink text-base font-bold cursor-pointer ${FOCUS_RING}`;

/* ============================
   Types
============================ */

/**
 * Caregiver profile structure
 */
interface Caregiver {
  id: number;
  name: string;
  role: string;
}

/**
 * Default fallback caregiver state
 */
const DEFAULT_CAREGIVER: Caregiver = {
  id: 0,
  name: "Your Name",
  role: "Primary Caregiver",
};

/* ============================
   Component
============================ */

/**
 * SettingsPage Component
 * 
 * Provides account settings, carepartner configuration,
 * and session management actions. Results show as toasts.
 */
export default function SettingsPage() {

  /* ============================
     Routing / Context
  ============================ */

  const navigate = useNavigate();
  const toast = useToast();

  const {
    user,
    setIsAuthenticated,
    setUser,
    activeCarePartnerId,
    setActiveCarePartnerId,
  } = useAuth();

  const carepartners = user?.carepartners ?? [];

  /**
   * Base API URL
   */
  const API_URL = import.meta.env.VITE_API_URL;

  /* ============================
     State
  ============================ */

  /**
   * Caregiver profile data
   */
  const [caregiver, setCaregiver] = useState<Caregiver>(DEFAULT_CAREGIVER);

  /**
   * True while Save is under way
   */
  const [saving, setSaving] = useState(false);

  /* ============================
     Effects
  ============================ */

  /**
   * Fetch caregiver and carepartner settings on load
   */
  useEffect(() => {
    const fetchCaregiverAndCarepartner = async () => {
      try {
        /* ============================
           Fetch caregiver profile
        ============================ */
        const resCg = await fetch(`${API_URL}/me`, { method: "GET", credentials: "include" });
        if (!resCg.ok) return;

        const cgData = await resCg.json();
        setCaregiver(cgData);
        setUser(cgData);

        const connectedCarepartners = Array.isArray(cgData.carepartners) ? cgData.carepartners : [];

        if (connectedCarepartners.length === 0) {
          setActiveCarePartnerId(null);
          return;
        }

        const hasActiveCarepartner = activeCarePartnerId !== null
          && connectedCarepartners.some((cp: { id: number }) => cp.id === activeCarePartnerId);

        if (!hasActiveCarepartner) {
          setActiveCarePartnerId(connectedCarepartners[0].id);
        }
      } catch (err) {
        console.error("Error fetching data:", err);
      }
    };

    fetchCaregiverAndCarepartner();
  }, [activeCarePartnerId]);


  /* ============================
     Handlers
  ============================ */

  /**
   * Logs the user out and clears authentication state
   */
  const handleLogout = async () => {
    try {
      // Call the backend to clear the cookie
      const res = await fetch(`${API_URL}/auth/logout`, {
        method: "POST",
        credentials: "include",
      });

      if (!res.ok) {
        throw new Error("Failed to logout");
      }

      // Update global auth state
      setIsAuthenticated(false);
      setUser(null);
      setActiveCarePartnerId(null);

      // Redirect to login
      navigate("/login");
    } catch (err) {
      toast.show("Logout failed: " + (err instanceof Error ? err.message : String(err)));
    }
  };

  /**
   * Updates caregiver profile information; returns an error message, or null when it worked
   */
  const handleUpdate = async (): Promise<string | null> => {
    try {
      const res = await fetch(`${API_URL}/caregivers`, {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          ...caregiver,
        }),
      });

      if (!res.ok) {
        const text = await res.text();
        return "Update failed: " + text;
      }

      setUser((currentUser) => {
        if (!currentUser) {
          return { ...caregiver, carepartners: [] };
        }

        return {
          ...currentUser,
          name: caregiver.name,
          role: caregiver.role,
        };
      });
      return null;
    } catch (err) {
      return "Update failed: " + (err instanceof Error ? err.message : String(err));
    }
  };

  /**
   * Saves the caregiver's profile and says how it went
   */
  const handleSave = async () => {
    setSaving(true);
    const error = await handleUpdate();
    setSaving(false);
    toast.show(error ?? "Settings saved.");
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
            value={caregiver.name}
            onChange={(e) => setCaregiver({ ...caregiver, name: e.target.value })}
          />
          <button type="button" onClick={handleLogout} className={`${SECONDARY_BUTTON} self-start`}>
            <LogOut size={18} aria-hidden="true" />
            Log out
          </button>
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
                : carepartners.map((carepartner) => ({ value: String(carepartner.id), label: carepartner.name }))
            }
          />
          <button
            type="button"
            onClick={() => navigate("/setup", { state: { from: "settings" } })}
            className={`${SECONDARY_BUTTON} self-start`}
          >
            <Smartphone size={18} aria-hidden="true" />
            Pair a new device
          </button>
        </section>

        <div className="flex justify-end">
          <button type="button" onClick={handleSave} disabled={saving} className={PRIMARY_BUTTON}>
            {saving ? "Saving…" : "Save changes"}
          </button>
        </div>
      </div>
    </div>
  );
}
