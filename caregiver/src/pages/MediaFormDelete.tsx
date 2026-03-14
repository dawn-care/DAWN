import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from '../context/AuthContext';

type MediaType = "video" | "photo" | "audio";

interface Module {
  id: number;
  module_type: MediaType;
  metadata_id: number;
}

interface MediaOption {
  moduleId: number;
  metadataId: number;
  type: MediaType;
  name: string;
}

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export default function MediaFormDelete() {
  const navigate = useNavigate();

  const [mediaList, setMediaList] = useState<MediaOption[]>([]);
  const [selected, setSelected] = useState<MediaOption | null>(null);
  const [status, setStatus] = useState("");
  const { user } = useAuth();

  /* ---------------- LOAD ALL MEDIA ---------------- */
  useEffect(() => {
    async function loadMedia() {
      try {
        if (!user?.id || !user.carepartner_id) {
          setStatus("Cannot load media: user information missing.");
          return;
        }

        const modulesRes = await fetch(
          `${API_URL}/modules?caregiver_id=${user.id}&carepartner_id=${user.carepartner_id}`
        );
        const modules: Module[] = await modulesRes.json();

        const mediaModules = modules.filter((m) =>
          ["photo", "video", "audio"].includes(m.module_type)
        );

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

  /* ---------------- DELETE ---------------- */
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
      // 1️⃣ Delete metadata (photo/video/audio)
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

      // 2️⃣ Delete module reference
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

      setStatus("Deleted successfully.");

      // Remove from dropdown immediately
      setMediaList((prev) =>
        prev.filter((m) => m.metadataId !== selected.metadataId)
      );

      setSelected(null);
    } catch (err: any) {
      setStatus("Delete failed: " + err.message);
    }
  }

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
