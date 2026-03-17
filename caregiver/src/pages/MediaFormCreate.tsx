import { useState } from "react";
import { useAuth } from '../context/AuthContext';

type MediaType = "video" | "photo" | "audio";

const allowedExtensions = {
  video: [".mp4", ".mov", ".avi"],
  photo: [".jpg", ".jpeg", ".png", ".heic"],
  audio: [".mp3", ".wav", ".ogg", ".flac", ".aac"],
};

const API_URL = import.meta.env.VITE_API_URL + "/api";

export default function MediaFormCreate() {
  const [mediaName, setMediaName] = useState("");
  const [mediaType, setMediaType] = useState<MediaType>("video");
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState("");
  const [uploadStatus, setUploadStatus] = useState("");
  const { user } = useAuth();

  function validateFile(f: File) {
    const ext = "." + f.name.split(".").pop()?.toLowerCase();
    if (!allowedExtensions[mediaType].includes(ext)) {
      setFileError(`Invalid file type for ${mediaType.toUpperCase()}.`);
      setFile(null);
    } else {
      setFileError("");
      setFile(f);
    }
  }

  async function handleUpload() {
    if (!mediaName.trim()) {
      setUploadStatus("Please enter a media name.");
      return;
    }

    if (!file) {
      setFileError("Please select a file.");
      return;
    }

    if (!user?.id || !user.carepartner_id) {
      setUploadStatus("Cannot upload: user information missing.");
      return;
    }

    const formData = new FormData();
    formData.append("file", file, file.name);
    formData.append("displayed_name", mediaName);
    formData.append("caregiver_id", String(user.id));
    formData.append("carepartner_id", String(user.carepartner_id));

    const endpoint =
      mediaType === "video"
        ? "/video/upload"
        : mediaType === "photo"
        ? "/photo/upload"
        : "/audio/upload";

    try {
      // Upload media metadata
      const response = await fetch(`${API_URL}${endpoint}`, {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        const text = await response.text();
        setUploadStatus(`Error: ${text}`);
        return;
      }

      const metadata = await response.json();
      const metadataId = metadata.id;
      if (!metadataId) {
        setUploadStatus("Upload succeeded but metadata ID not returned.");
        return;
      }

      // Create a module that references this metadata
      const moduleResponse = await fetch(`${API_URL}/modules`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          module_type: mediaType,
          caregiver_id: user.id,
          carepartner_id: user.carepartner_id,
          metadata_id: metadataId,
        }),
      });

      if (!moduleResponse.ok) {
        const text = await moduleResponse.text();
        setUploadStatus(`Media saved but failed to create module: ${text}`);
        return;
      }

      setUploadStatus("Saved!");
      setFile(null);
      setMediaName("");
    } catch (err: any) {
      setUploadStatus("Upload failed: " + err.message);
    }
  }

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">Add Media</h1>
      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">

        {/* Media Name */}
        <div className="mb-6">
          <label className="text-lg font-medium">Media Name:</label>
          <input
            type="text"
            value={mediaName}
            onChange={(e) => setMediaName(e.target.value)}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
            placeholder="Enter media name"
          />
        </div>

        {/* Media Type */}
        <div className="mb-6">
          <label className="text-lg font-medium">Media Type:</label>
          <select
            value={mediaType}
            onChange={(e) => setMediaType(e.target.value as MediaType)}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
          >
            <option value="video">Video</option>
            <option value="photo">Photo</option>
            <option value="audio">Audio</option>
          </select>
        </div>

        {/* Upload */}
        <div className="mb-6">
          <label className="text-lg font-medium">Upload Media:</label>
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
              f && validateFile(f);
            }}
          />

          {fileError && <p className="text-red-500 mt-2 text-sm">{fileError}</p>}
        </div>

        {/* Save */}
        <button
          className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
          onClick={handleUpload}
        >
          Save
        </button>

        {uploadStatus && <p className="mt-2 text-sm">{uploadStatus}</p>}
      </div>
    </div>
  );
}
