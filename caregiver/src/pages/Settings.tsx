import { useEffect, useState } from "react";
import { useAuth } from "../context/AuthContext";
import { useNavigate } from "react-router-dom";

interface Caregiver {
  id: number;
  carepartner_id: number;
  name: string;
  role: string;
}

const DEFAULT_CAREGIVER: Caregiver = {
  id: 0,
  carepartner_id: 0,
  name: "Your Name",
  role: "Primary Caregiver",
};

export default function SettingsPage() {
  const { setIsAuthenticated } = useAuth();
  const navigate = useNavigate();

  const [caregiver, setCaregiver] = useState<Caregiver>(DEFAULT_CAREGIVER);
  const [updateStatus, setUpdateStatus] = useState<string | null>(null);
  const [logoutError, setLogoutError] = useState<string | null>(null);

  // Get the caregiver info
  useEffect(() => {
    const fetchCaregiver = async () => {
      try {
        const res = await fetch("http://localhost:8080/api/me", {
          method: "GET",
          credentials: "include",
        });

        if (!res.ok) {
          console.warn("Failed to fetch caregiver, using defaults");
          return;
        }

        const data = await res.json();
        setCaregiver(data);
      } catch (err) {
        console.error("Error fetching caregiver, using defaults:", err);
      }
    };

    fetchCaregiver();
  }, []);

  // Logout
const handleLogout = async () => {
  try {
    // Call the backend to clear the cookie
    const res = await fetch("http://localhost:8080/api/auth/logout", {
      method: "POST",
      credentials: "include",
    });

    if (!res.ok) {
      throw new Error("Failed to logout");
    }

    // Update global auth state
    setIsAuthenticated(false);

    // Redirect to login
    navigate("/login");
  } catch (err: any) {
    setLogoutError("Logout failed: " + err.message);
  }
};

  // Update caregiver info
  const handleUpdate = async () => {
    try {
      const res = await fetch("http://localhost:8080/caregivers", {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(caregiver),
      });

      if (!res.ok) {
        const text = await res.text();
        setUpdateStatus("Update failed: " + text);
        return;
      }

      const result = await res.json();
      setUpdateStatus(result.message || "Settings updated!");
    } catch (err: any) {
      setUpdateStatus("Update failed: " + err.message);
    }
  };

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">Settings</h1>

      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">
        {/* Name */}
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

        {/* Care Partner ID */}
        <div className="mb-6">
          <label className="text-lg font-medium">Care Partner ID:</label>
          <input
            type="text"
            value={caregiver.carepartner_id}
            className="w-full mt-2 p-3 border rounded bg-gray-200 text-gray-700"
            readOnly
          />
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

        {/* Update & Logout */}
        <div className="flex gap-4 mt-4">
          <button
            className="mt-4 flex-1 bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
            onClick={handleUpdate}
          >
            Save
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