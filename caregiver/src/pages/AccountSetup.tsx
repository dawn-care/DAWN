import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export default function AccountSetupPage() {
  const navigate = useNavigate();
  const { user, setUser } = useAuth(); // access logged-in user
  const [pairingCode, setPairingCode] = useState("");
  const [consentChecked, setConsentChecked] = useState(false);
  const [submitError, setSubmitError] = useState("");

  const handleSubmit = async () => {
    // Ensure pairing code is entered
    if (!pairingCode.trim()) {
      setSubmitError("Pairing code is required.");
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
      const API_URL = import.meta.env.VITE_API_URL;
        const res = await fetch(`${API_URL}/users/device-pairing/verify`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          code: pairingCode,
          caregiver_id: user.id, 
        }),
      });

      if (!res.ok) {
        const errText = await res.text();
        setSubmitError(errText || "Failed to verify pairing code. Please try again.");
        return;
      }

      // Update user context to have the correct carepartner id after pairing
      setUser({ ...user, carepartner_id: user.carepartner_id || 1 });

      // Pairing succeeded
      navigate("/library");
    } catch (err) {
      console.error(err);
      setSubmitError("Something went wrong. Please try again.");
    }
  };

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">
        Complete Account Setup
      </h1>

      <div className="border p-8 max-w-3xl w-full rounded-lg bg-white shadow">

        {/* Pairing Code */}
        <div className="mb-6">
          <label className="text-lg font-medium">Enter Care Partner Pairing Code:</label>
          <input
            type="text"
            value={pairingCode}
            onChange={(e) => setPairingCode(e.target.value)}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
            placeholder="Enter pairing code"
          />
        </div>

        {/* Consent Statement */}
        <div className="mb-6">
          <label className="text-lg font-medium">Consent Statement:</label>
          <div className="mt-2 p-4 border rounded bg-gray-50 max-h-48 overflow-y-auto text-sm text-gray-700">
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
            <label htmlFor="consentCheckbox" className="text-sm text-gray-700">
              I agree to the consent statement
            </label>
          </div>
        </div>

        {/* Error Message */}
        {submitError && (
          <p className="text-red-500 mb-4">{submitError}</p>
        )}

        {/* Submit Button */}
        <button
          onClick={handleSubmit}
          className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
        >
          Complete Setup
        </button>
      </div>
    </div>
  );
}