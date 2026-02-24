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
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchCaregiver = async () => {
      try {
        const res = await fetch("/api/me", {
          method: "GET",
          credentials: "include",
        });

        if (res.ok) {
          const data = await res.json();
          setCaregiver(data);
        } else {
          console.warn("Failed to fetch caregiver, using defaults");
        }
      } catch (err) {
        console.error("Error fetching caregiver, using defaults:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchCaregiver();
  }, []);

  const handleLogout = () => {
    setIsAuthenticated(false);
    navigate("/login");
  };

  const handleUpdate = () => {
    alert(`Update functionality not implemented yet. Name: ${caregiver.name}`);
  };

  if (loading) return <p>Loading...</p>;

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">Settings</h1>

      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">
        <div className="mb-6">
          <label className="text-lg font-medium">Name:</label>
          <input
            type="text"
            value={caregiver.name}
            onChange={(e) => setCaregiver({ ...caregiver, name: e.target.value })}
            className="w-full mt-2 p-3 border rounded bg-white text-gray-900"
          />
        </div>

        <div className="mb-6">
          <label className="text-lg font-medium">Care Partner ID:</label>
          <input
            type="text"
            value={caregiver.carepartner_id}
            className="w-full mt-2 p-3 border rounded bg-gray-200 text-gray-700"
            readOnly
          />
        </div>

        <div className="mb-6">
          <label className="text-lg font-medium">Role:</label>
          <input
            type="text"
            value={caregiver.role}
            className="w-full mt-2 p-3 border rounded bg-gray-200 text-gray-700"
            readOnly
          />
        </div>

        <div className="flex gap-4 mt-4">
          <button
            className="flex-1 bg-teal-200 hover:bg-teal-300 text-white font-semibold py-3 rounded"
            onClick={handleUpdate}
          >
            Update
          </button>

          <button
            className="flex-1 bg-red-500 hover:bg-red-600 text-white font-semibold py-3 rounded"
            onClick={handleLogout}
          >
            Logout
          </button>
        </div>
      </div>
    </div>
  );
}