import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export default function SelectCarePartnerPage() {
  const { user, setActiveCarePartnerId } = useAuth();
  const navigate = useNavigate();
  const [selectedId, setSelectedId] = useState("");

  const handleContinue = () => {
    if (!selectedId) return;

    // store selection via context (which syncs to localStorage)
    setActiveCarePartnerId(Number(selectedId));

    navigate("/library");
  };

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">
        Who is this for?
      </h1>

      <div className="border p-8 max-w-xl w-full rounded-lg bg-white shadow">

        {/* Dropdown */}
        <select
          value={selectedId}
          onChange={(e) => setSelectedId(e.target.value)}
          className="w-full p-3 border rounded bg-purple-50"
        >
          <option value="">Select care partner</option>
          {user?.carepartners?.map((cp: any) => (
            <option key={cp.id} value={cp.id}>
              {cp.name}
            </option>
          ))}
        </select>

        {/* Info text */}
        <p className="mt-4 text-sm text-gray-600 italic text-center">
          You will be able to switch this in{" "}
          <span className="italic font-medium">Agenda</span>
        </p>

        <button
          onClick={handleContinue}
          className="mt-6 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
        >
          Continue
        </button>
      </div>
    </div>
  );
}