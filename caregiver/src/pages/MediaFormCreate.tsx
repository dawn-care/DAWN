/**
 * MediaFormCreate.tsx
 * 
 * User: Caregiver
 *
 * This component provides a form for uploading media content.
 * Supported media types:
 *  - Video
 *  - Photo
 *  - Audio
 *
 * Features:
 *  - File validation based on media type
 *  - Drag-and-drop upload support
 *  - Manual file selection
 *  - Uploads media to backend
 *  - Creates associated module after upload
 *
 * API interactions:
 *  - Upload media file (video/photo/audio endpoints)
 *  - Create module linked to uploaded media
 */

import { useState, useRef } from "react";
import { useAuth } from '../context/AuthContext';

/* ============================
   Type Definitions
============================ */

/**
 * Supported media types
 */
type MediaType = "video" | "photo" | "audio";

/* ============================
   Static Data
============================ */

/**
 * Allowed file extensions per media type
 */
const allowedExtensions = {
  video: [".mp4", ".mov", ".avi", ".webm", ".ogg"],
  photo: [".jpg", ".jpeg", ".png", ".heic", ".gif", ".webp"],
  audio: [".mp3", ".wav", ".ogg", ".flac", ".aac"],
};

/**
 * Base API URL from environment config
 */
const API_URL = import.meta.env.VITE_API_URL;

/* ============================
   Component
============================ */

/**
 * MediaFormCreate Component
 *
 * Handles:
 *  - Uploading media files
 *  - Validating file types
 *  - Creating modules linked to uploaded media
 *
 * @param onUploadSuccess - Callback triggered after successful upload
 */
export default function MediaFormCreate({ onUploadSuccess }: { onUploadSuccess: () => void }) {
  
  /* ============================
      State
  ============================ */

  const [mediaName, setMediaName] = useState("");
  const [mediaType, setMediaType] = useState<MediaType>("video");
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState("");
  const [uploadStatus, setUploadStatus] = useState("");
  const [isDragging, setIsDragging] = useState(false);

  /**
   * Reference to hidden file input
   */
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  /**
   * Auth context
   */
  const { user, activeCarePartnerId } = useAuth();

  /* ============================
      File Validation
  ============================ */

  /**
   * Validates selected file against allowed extensions
   */
  function validateFile(f: File) {
    const ext = "." + f.name.split(".").pop()?.toLowerCase();
    if (!allowedExtensions[mediaType].includes(ext)) {
      setFileError(`Invalid file type for ${mediaType.toUpperCase()}.`);
      setFile(null);

      // Reset input so same file can be reselected
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    } else {
      setFileError("");
      setFile(f);
    }
  }

  /* ============================
      Drag & Drop Handlers
  ============================ */

  function handleDragOver(e: React.DragEvent) {
    e.preventDefault();
    setIsDragging(true);
  }

  function handleDragLeave() {
    setIsDragging(false);
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault();
    setIsDragging(false);

    const droppedFile = e.dataTransfer.files?.[0];
    if (droppedFile) {
      validateFile(droppedFile);
    }
  }

  /* ============================
      Upload Handler
  ============================ */

  /**
   * Handles media upload and module creation
   */
  async function handleUpload() {

    // Input Validation
    if (!mediaName.trim()) {
      setUploadStatus("Please enter a media name.");
      return;
    }

    if (!file) {
      setFileError("Please select a file.");
      return;
    }

    if (!user?.id || !activeCarePartnerId) {
      setUploadStatus("Cannot upload: user information missing.");
      return;
    }

    // Prepare form data for upload
    const formData = new FormData();
    formData.append("file", file, file.name);
    formData.append("displayed_name", mediaName);
    formData.append("caregiver_id", String(user.id));
    formData.append("carepartner_id", String(activeCarePartnerId));

    /**
     * Select correct endpoint based on media type
     */
    const endpoint =
      mediaType === "video"
        ? "/video/upload"
        : mediaType === "photo"
        ? "/photo/upload"
        : "/audio/upload";

    try {
      /* ============================
         Upload Media File
      ============================ */
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

      /* ============================
         Create Module
      ============================ */
      const moduleResponse = await fetch(`${API_URL}/modules`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          module_type: mediaType,
          caregiver_id: user.id,
          carepartner_id: activeCarePartnerId,
          metadata_id: metadataId,
        }),
      });

      if (!moduleResponse.ok) {
        const text = await moduleResponse.text();
        setUploadStatus(`Media saved but failed to create module: ${text}`);
        return;
      }

      /* ============================
         Success Handling
      ============================ */
      setUploadStatus("Saved!");
      setFile(null);
      setMediaName("");

      if (onUploadSuccess) onUploadSuccess();

      // Reset input after success
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }

    } catch (err) {
      setUploadStatus("Upload failed: " + (err instanceof Error ? err.message : String(err)));
    }
  }

  /* ============================
      Render
  ============================ */
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
            onChange={(e) => {
              setMediaType(e.target.value as MediaType);
              setFile(null);
              setFileError("");
            }}
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
            className={`mt-2 p-6 border-2 border-dashed rounded cursor-pointer text-center transition
              ${isDragging ? "border-blue-500 bg-blue-50" : "border-gray-400 bg-purple-50"}
            `}
            onClick={() => fileInputRef.current?.click()}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
          >
            {file ? (
              <span>{file.name}</span>
            ) : (
              <span>Drag and drop or click to upload</span>
            )}
          </div>

          <input
            ref={fileInputRef}
            id="fileInput"
            type="file"
            className="hidden"
            onChange={(e) => {
              const f = e.target.files?.[0];
              if (f) validateFile(f);

              // allow selecting same file again
              e.target.value = "";
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
