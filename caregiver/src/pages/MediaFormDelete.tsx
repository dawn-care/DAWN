/**
 * MediaFormDelete.tsx
 * 
 * User: Caregiver
 *
 * This component allows users to delete previously uploaded media.
 * Supported media types:
 *  - Video
 *  - Photo
 *  - Audio
 *
 * Features:
 *  - Fetches all media modules for the current caregiver + carepartner
 *  - Displays media in a dropdown for selection
 *  - Deletes both:
 *      1. Media metadata (photo/video/audio)
 *      2. Associated module reference
 *  - Updates UI immediately after deletion
 *
 * API interactions:
 *  - GET modules (filtered by caregiver + carepartner)
 *  - GET metadata for each media item
 *  - DELETE media metadata
 *  - DELETE module
 */

import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from '../context/AuthContext';

/* ============================
   Type Definitions
============================ */

/**
 * Supported media types
 */
type MediaType = "video" | "photo" | "audio";

/**
 * Represents a module returned from backend
 */
interface Module {
  id: number;
  module_type: MediaType;
  metadata_id: number;
}

/**
 * Represents a selectable media item in the UI
 */
interface MediaOption {
  moduleId: number;
  metadataId: number;
  type: MediaType;
  name: string;
}

/* ============================
   Static Data
============================ */

/**
 * Base API URL from environment config
 */
const API_URL = import.meta.env.VITE_API_URL;

/* ============================
   Component
============================ */

/**
 * MediaFormDelete Component
 *
 * Handles:
 *  - Fetching media list
 *  - Selecting media item
 *  - Deleting media + associated module
 */
export default function MediaFormDelete() {
  
  /* ============================
      Routing / Context
  ============================ */

  const navigate = useNavigate();
  const { user, activeCarePartnerId } = useAuth();

  /* ============================
      State
  ============================ */

  const [mediaList, setMediaList] = useState<MediaOption[]>([]);
  const [selected, setSelected] = useState<MediaOption | null>(null);
  const [status, setStatus] = useState("");

  /* ============================
      Effects
  ============================ */

  /**
   * Load all media associated with caregiver + carepartner
   */
  useEffect(() => {
    async function loadMedia() {
      try {
        if (!user?.id || !activeCarePartnerId) {
          setStatus("Cannot load media: user information missing.");
          return;
        }

        /* ============================
           Fetch Modules
        ============================ */
        const modulesRes = await fetch(
          `${API_URL}/modules?caregiver_id=${user.id}&carepartner_id=${activeCarePartnerId}`
        );
        const modules: Module[] = await modulesRes.json();

        /**
         * Filter only media-related modules
         */
        const mediaModules = modules.filter((m) =>
          ["photo", "video", "audio"].includes(m.module_type)
        );

        /* ============================
           Fetch Metadata for Each Media
        ============================ */
        const results = await Promise.all(
          mediaModules.map(async (m) => {
            const res = await fetch(
              `${API_URL}/${m.module_type}?id=${m.metadata_id}`
            );
            if (!res.ok) return null;

            const metadata = await res.json();

            return {
              moduleId: m.id,
              metadataId: m.metadata_id,
              type: m.module_type,
              name: metadata.displayed_name || metadata.filename,
            } as MediaOption;
          })
        );

        setMediaList(results.filter(Boolean) as MediaOption[]);
      } catch (err) {
        setStatus("Failed to load media.");
      }
    }

    loadMedia();
  }, []);

  /* ============================
      Delete Handler
  ============================ */

  /**
   * Deletes selected media and its module
   */
  async function handleDelete() {
    if (!selected) {
      setStatus("Please select media to delete.");
      return;
    }

    const confirmDelete = window.confirm(
      `Are you sure you want to delete "${selected.name}"?`
    );

    if (!confirmDelete) return;

    try {
      /* ============================
         Delete Media Metadata
      ============================ */
      const metaRes = await fetch(
        `${API_URL}/${selected.type}?id=${selected.metadataId}`,
        {
          method: "DELETE",
        }
      );

      if (!metaRes.ok) {
        const text = await metaRes.text();
        throw new Error(text);
      }

      /* ============================
         Delete Module Reference
      ============================ */
      const moduleRes = await fetch(
        `${API_URL}/modules?id=${selected.moduleId}`,
        {
          method: "DELETE",
        }
      );

      if (!moduleRes.ok) {
        const text = await moduleRes.text();
        throw new Error(text);
      }

      /* ============================
         Success Handling
      ============================ */
      setStatus("Deleted successfully.");

      // Remove deleted item from UI immediately
      setMediaList((prev) =>
        prev.filter((m) => m.metadataId !== selected.metadataId)
      );

      setSelected(null);
    } catch (err: any) {
      setStatus("Delete failed: " + err.message);
    }
  }

  /* ============================
      Render
  ============================ */
  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">Delete Media</h1>

      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">
        {/* Select Media */}
        <div className="mb-6">
          <label className="text-lg font-medium">
            Select Media to Delete:
          </label>
          <select
            value={selected?.metadataId || ""}
            onChange={(e) => {
              const found = mediaList.find(
                (m) => m.metadataId === Number(e.target.value)
              );
              setSelected(found || null);
            }}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
          >
            <option value="">-- Choose Media --</option>
            {mediaList.map((m) => (
              <option key={m.metadataId} value={m.metadataId}>
                {m.name} ({m.type})
              </option>
            ))}
          </select>
        </div>

        {/* Delete Button */}
        <button
          className="mt-4 w-full bg-red-300 hover:bg-red-400 text-gray-800 font-semibold py-3 rounded"
          onClick={handleDelete}
        >
          Delete
        </button>

        {status && <p className="mt-3 text-sm">{status}</p>}

        <button
          onClick={() => navigate("/library")}
          className="mt-4 w-full bg-gray-200 hover:bg-gray-300 font-semibold py-3 rounded"
        >
          Back to Library
        </button>
      </div>
    </div>
  );
}
