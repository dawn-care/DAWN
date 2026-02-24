import { useState } from "react";
import { useNavigate } from "react-router-dom";

export default function AccountSetupPage() {
  const navigate = useNavigate();

  const [pairingCode, setPairingCode] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState("");
  const [submitError, setSubmitError] = useState("");

  const validateFile = (f: File) => {
    const allowedTypes = ["application/pdf", "image/png", "image/jpeg"];

    if (!allowedTypes.includes(f.type)) {
      setFileError("Only PDF, PNG, or JPG files are allowed.");
      return;
    }

    if (f.size > 5 * 1024 * 1024) {
      setFileError("File size must be under 5MB.");
      return;
    }

    setFileError("");
    setFile(f);
  };

  const handleSubmit = async () => {
    if (!pairingCode.trim()) {
      setSubmitError("Pairing code is required.");
      return;
    }

    if (!file) {
      setSubmitError("Written consent document is required.");
      return;
    }

    setSubmitError("");

    const formData = new FormData();
    formData.append("pairingCode", pairingCode);
    formData.append("consentForm", file);

    try {
        const res = await fetch("http://localhost:8080/api/me", {
        method: "PUT",
        credentials: "include",
        body: formData,
        });

      if (!res.ok) {
        setSubmitError("Failed to complete setup. Please try again.");
        return;
      }

      navigate("/library");
    } catch (err) {
      console.error(err);
      setSubmitError("Something went wrong.");
    }
  };

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">
        Complete Account Setup
      </h1>

      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">

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

        {/* Written Consent Upload */}
        <div className="mb-6">
          <label className="text-lg font-medium">
            Upload Written Consent (PDF or Image):
          </label>

          <div
            className="mt-2 p-6 border-2 border-dashed border-gray-400 rounded cursor-pointer bg-purple-50 text-center"
            onClick={() => document.getElementById("fileInput")?.click()}
          >
            {file ? <span>{file.name}</span> : <span>Drag and drop or click to upload</span>}
          </div>

          <input
            id="fileInput"
            type="file"
            className="hidden"
            onChange={(e) => {
              const f = e.target.files?.[0];
              if (f) validateFile(f);
            }}
          />

          {fileError && (
            <p className="text-red-500 mt-2 text-sm">{fileError}</p>
          )}
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