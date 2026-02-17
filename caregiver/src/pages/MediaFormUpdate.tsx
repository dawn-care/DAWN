import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

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

const allowedExtensions: Record<MediaType, string[]> = {
  video: [".mp4", ".mov", ".avi"],
  photo: [".jpg", ".jpeg", ".png", ".heic"],
  audio: [".mp3", ".wav", ".ogg", ".flac", ".aac"],
};

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export default function MediaFormUpdate() {
  const navigate = useNavigate();

  const [mediaList, setMediaList] = useState<MediaOption[]>([]);
  const [selected, setSelected] = useState<MediaOption | null>(null);

  const [mediaName, setMediaName] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState("");
  const [status, setStatus] = useState("");

  /* ---------------- LOAD ALL MEDIA ---------------- */
  useEffect(() => {
    async function loadMedia() {
      try {
        const modulesRes = await fetch(
          `${API_URL}/modules?caregiver_id=1&carepartner_id=1`
        );

        if (!modulesRes.ok) throw new Error("Failed to load modules");

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
        setStatus("Failed to load media list.");
      }
    }

    loadMedia();
  }, []);

  /* ---------------- FILE VALIDATION ---------------- */
  function validateFile(f: File) {
    if (!selected) return;

    const ext = "." + f.name.split(".").pop()?.toLowerCase();

    if (!allowedExtensions[selected.type].includes(ext)) {
      setFileError(`Invalid file type for ${selected.type.toUpperCase()}.`);
      setFile(null);
    } else {
      setFileError("");
      setFile(f);
    }
  }

  /* ---------------- UPDATE ---------------- */
  async function handleUpdate() {
    if (!selected) {
      setStatus("Please select media to update.");
      return;
    }

    if (!mediaName.trim()) {
      setStatus("Please enter a new media name.");
      return;
    }

    try {
      /* -------- UPDATE METADATA -------- */
      const metaRes = await fetch(`${API_URL}/${selected.type}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id: selected.metadataId,
          displayed_name: mediaName,
        }),
      });

      if (!metaRes.ok) throw new Error(await metaRes.text());

      /* -------- UPDATE FILE (OPTIONAL) -------- */
      if (file) {
        const formData = new FormData();
        formData.append("id", selected.metadataId.toString());
        formData.append("file", file);

        const fileRes = await fetch(
          `${API_URL}/${selected.type}/file`,
          {
            method: "PUT",
            body: formData,
          }
        );

        if (!fileRes.ok) throw new Error(await fileRes.text());
      }

      setStatus("Updated successfully!");
      setFile(null);
      setMediaName("");

      setTimeout(() => window.location.reload(), 800);

    } catch (err: any) {
      setStatus("Update failed: " + err.message);
    }
  }

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">Update Media</h1>

      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">
        {/* Select Media */}
        <div className="mb-6">
          <label className="text-lg font-medium">
            Select Media to Update:
          </label>

          <select
            value={selected?.metadataId || ""}
            onChange={(e) => {
              const found = mediaList.find(
                (m) => m.metadataId === Number(e.target.value)
              );

              setSelected(found || null);
              setMediaName(found?.name || "");
              setFile(null);
              setFileError("");
              setStatus("");
            }}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
          >
            <option value="">-- Choose Media --</option>

            {mediaList.map((m) => (
              <option key={`${m.type}-${m.metadataId}`} value={m.metadataId}>
                {m.name} ({m.type})
              </option>
            ))}
          </select>
        </div>

        {/* New Name */}
        <div className="mb-6">
          <label className="text-lg font-medium">New Media Name:</label>
            <label>optional</label>
          <input
            type="text"
            value={mediaName}
            onChange={(e) => setMediaName(e.target.value)}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
          />
        </div>

        {/* Replace File */}
        {selected && (
          <div className="mb-6">
            <label className="text-lg font-medium">
              Replace File
            </label>
            <label>optional</label>
            <div
              className="mt-2 p-6 border-2 border-dashed border-gray-400 rounded cursor-pointer bg-purple-50 text-center"
              onClick={() =>
                document.getElementById("fileInput")?.click()
              }
            >
              {file ? file.name : "Drag and drop or click to upload"}
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
        )}

        {/* Update Button */}
        <button
          className="mt-4 w-full bg-teal-200 hover:bg-teal-300 font-semibold py-3 rounded"
          onClick={handleUpdate}
        >
          Update
        </button>

        {status && <p className="mt-2 text-sm">{status}</p>}

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
