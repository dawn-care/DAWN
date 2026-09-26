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
 *  - Configure carepartner auto-start time
 *  - Logout user and clear session
 *  - Navigate to device repair/setup flow
 */

import { useEffect, useState } from "react";
import { useAuth } from "../context/AuthContext";
import { useNavigate } from "react-router-dom";

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
 * and session management actions.
 */
export default function SettingsPage() {

  /* ============================
     Routing / Context
  ============================ */

  const navigate = useNavigate();

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
   * Carepartner auto-start time (HH:mm format)
   */
  const [autoStart, setAutoStart] = useState<string>("08:00");

  /**
   * UI status messages for updates
   */
  const [updateStatus, setUpdateStatus] = useState<string | null>(null);

  /**
   * Logout error message
   */
  const [logoutError, setLogoutError] = useState<string | null>(null);

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
          setAutoStart("08:00");
          return;
        }

        const hasActiveCarepartner = activeCarePartnerId !== null
          && connectedCarepartners.some((cp: { id: number }) => cp.id === activeCarePartnerId);

        if (!hasActiveCarepartner) {
          setActiveCarePartnerId(connectedCarepartners[0].id);
          return;
        }

        /* ============================
           Fetch carepartner settings
        ============================ */
        if (!activeCarePartnerId) return;
        const resCp = await fetch(`${API_URL}/users/carepartner/${activeCarePartnerId}`, {
          method: "GET",
          credentials: "include",
        });
        if (!resCp.ok) return;

        const cpData = await resCp.json();

        if (cpData.auto_start) {
          const [hour, minute] = cpData.auto_start.split(":");
          setAutoStart(`${hour.padStart(2, "0")}:${minute.padStart(2, "0")}`);
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
      setLogoutError("Logout failed: " + (err instanceof Error ? err.message : String(err)));
    }
  };

  /**
   * Updates caregiver profile information
   */
  const handleUpdate = async () => {
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
        setUpdateStatus("Update failed: " + text);
        return;
      }

      const result = await res.json();
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
      setUpdateStatus(result.message || "Settings updated!");
    } catch (err) {
      setUpdateStatus("Update failed: " + (err instanceof Error ? err.message : String(err)));
    }
  };

  /**
   * Updates carepartner auto-start configuration
   */
  const handleUpdateAutoStart = async () => {
    if (!activeCarePartnerId) return;

    try {
      const res = await fetch(`${API_URL}/users/carepartner/update`, {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id: activeCarePartnerId,
          auto_start: autoStart + ":00", // backend expects HH:MM:SS
        }),
      });

      if (!res.ok) {
        const text = await res.text();
        setUpdateStatus("Update failed: " + text);
        return;
      }

      const result = await res.json();
      setUpdateStatus(result.message || "Carepartner settings updated!");
    } catch (err) {
      setUpdateStatus("Update failed: " + (err instanceof Error ? err.message : String(err)));
    }
  };

  /* ============================
     Render
  ============================ */
  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">Settings</h1>

      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">
        
        {/* Caregiver Name */}
        <div className="mb-6">
          <label className="text-lg font-medium">Name:</label>
          <input
            type="text"
            value={caregiver.name}
            onChange={(e) =>
              setCaregiver({ ...caregiver, name: e.target.value })
            }
            className="w-full mt-2 p-3 border rounded bg-white text-gray-900"
          />
        </div>

        {/* Active Care Partner */}
        <div className="mb-6">
          <label className="text-lg font-medium">Active Care Partner:</label>
          <select
            value={activeCarePartnerId ?? ""}
            onChange={(e) => setActiveCarePartnerId(Number(e.target.value))}
            className="w-full mt-2 p-3 border rounded bg-white text-gray-900"
            disabled={carepartners.length === 0}
          >
            {carepartners.length === 0 ? (
              <option value="">No connected care partners</option>
            ) : (
              carepartners.map((carepartner) => (
                <option key={carepartner.id} value={carepartner.id}>
                  {carepartner.name} (ID: {carepartner.id})
                </option>
              ))
            )}
          </select>
        </div>

        {/* Role */}
        <div className="mb-6">
          <label className="text-lg font-medium">Role:</label>
          <input
            type="text"
            value={caregiver.role}
            className="w-full mt-2 p-3 border rounded bg-gray-200 text-gray-700"
            readOnly
          />
        </div>

        {/* Auto Start Time */}
        <div className="mb-6">
          <label className="text-lg font-medium">Auto Start Time:</label>
          <input
            type="time"
            value={autoStart}
            onChange={(e) => setAutoStart(e.target.value)}
            className="w-full mt-2 p-3 border rounded bg-white text-gray-900"
          />
        </div>

        {/* Update & Logout */}
        <div className="flex gap-4 mt-4">
          <button
            className="mt-4 flex-1 bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
            onClick={async () => {
              await handleUpdate();          // Caregiver update
              await handleUpdateAutoStart(); // Carepartner auto start update
            }}
          >
            Save
          </button>

          <button
            className="mt-4 flex-1 bg-yellow-400 hover:bg-yellow-500 text-gray-800 font-semibold py-3 rounded"
            onClick={() => navigate("/setup")}
          >
            Pair New Device
          </button>

          <button
            className="mt-4 flex-1 bg-red-500 hover:bg-red-600 text-gray-800 font-semibold py-3 rounded"
            onClick={handleLogout}
          >
            Logout
          </button>
        </div>

        {/* Status messages */}
        {updateStatus && (
          <p className="mt-2 text-sm text-gray-800">{updateStatus}</p>
        )}
        {logoutError && (
          <p className="mt-2 text-sm text-red-500">{logoutError}</p>
        )}
      </div>
    </div>
  );
}