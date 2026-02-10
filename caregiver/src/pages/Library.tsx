import { useEffect, useMemo, useState } from "react";
import { Trash2, Plus, Check } from "lucide-react";
import { useNavigate } from "react-router-dom";

type MediaType = "video" | "photo" | "audio";

interface MediaItem {
  id: number;
  moduleId: number;
  name: string;
  filename: string;
  type: MediaType;
  uri: string;
}

type Filter = {
  key: string;
  label: string;
  value: string;
};

const FILTERS: Filter[] = [
  { key: "pictures", label: "Pictures", value: "photo" },
  { key: "videos", label: "Videos", value: "video" },
  { key: "music", label: "Music", value: "audio" },
  { key: "png", label: ".png", value: ".png" },
  { key: "jpg", label: ".jpg", value: ".jpg" },
  { key: "mov", label: ".mov", value: ".mov" },
  { key: "wav", label: ".wav", value: ".wav" },
  { key: "mp3", label: ".mp3", value: ".mp3" },
  { key: "ogg", label: ".ogg", value: ".ogg" },
];

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export default function Library() {
  const [media, setMedia] = useState<MediaItem[]>([]);
  const [search, setSearch] = useState("");
  const [activeFilter, setActiveFilter] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    async function fetchMedia() {
      try {
        // Get modules
        const modulesRes = await fetch(
          `${API_URL}/modules?caregiver_id=1&carepartner_id=1`
        );

        if (!modulesRes.ok) {
          throw new Error("Failed to fetch modules");
        }

        const text = await modulesRes.text();
        console.log("RAW MODULE RESPONSE:", text);

        const modules = JSON.parse(text);

        // **TEMP**
        const mediaModules = modules.filter(
          (m: any) =>
            ["photo", "video", "audio"].includes(m.module_type)
        );

        // Fetch metadata for each module
        const mediaResults = await Promise.all(
          mediaModules.map(async (m: any) => {
            const type = m.module_type as MediaType;

            const res = await fetch(`${API_URL}/${type}?id=${m.metadata_id}`);
            if (!res.ok) return null;

            const data = await res.json();

            return {
              id: data.id,
              moduleId: m.id,
              name: data.displayed_name || data.filename,
              filename: data.filename,
              type,
              uri: `${API_URL}/${type}/file?id=${data.id}`,
            } as MediaItem;
          })
        );

        setMedia(mediaResults.filter(Boolean));
      } catch (err) {
        console.error("Failed to fetch media:", err);
      }
    }

    fetchMedia();
  }, []);

  const handleDelete = async (
    // metadataId: number,
    moduleId: number,
    // type: MediaType
  ) => {
    if (!confirm("Are you sure you want to delete this item?")) return;

    try {
      // // Delete module first
      // await fetch(`${API_URL}/modules/${moduleId}`, {
      //   method: "DELETE",
      // });

      // // Delete actual media metadata/file
      // await fetch(`${API_URL}/${type}?id=${metadataId}`, {
      //   method: "DELETE",
      // });

      setMedia((prev) =>
        prev.filter((m) => m.moduleId !== moduleId)
      );
    } catch (err) {
      console.error("Delete failed:", err);
    }
  };

  const handleRename = async (
    metadataId: number,
    // type: MediaType,
    newName: string
  ) => {
    try {
      // const res = await fetch(`${API_URL}/${type}`, {
      //   method: "PUT",
      //   headers: { "Content-Type": "application/json" },
      //   body: JSON.stringify({
      //     id: metadataId,
      //     displayed_name: newName,
      //   }),
      // });

      // if (!res.ok) throw new Error("Rename failed");

      setMedia((prev) =>
        prev.map((m) =>
          m.id === metadataId ? { ...m, name: newName } : m
        )
      );
    } catch (err) {
      console.error("Rename failed:", err);
    }
  };

  const filteredMedia = useMemo(() => {
    return media.filter((item) => {
      const matchesSearch =
        item.name.toLowerCase().includes(search.toLowerCase()) ||
        item.filename.toLowerCase().includes(search.toLowerCase());

      const matchesFilter = !activeFilter
        ? true
        : activeFilter.startsWith(".")
        ? item.filename.toLowerCase().endsWith(activeFilter)
        : item.type === activeFilter;

      return matchesSearch && matchesFilter;
    });
  }, [media, search, activeFilter]);

  return (
    <div className="relative w-full px-8 py-10">
      <h1 className="text-4xl font-medium text-center mb-8">
        Your Uploaded Content
      </h1>

      <div className="max-w-3xl mx-auto mb-4">
        <input
          type="text"
          placeholder="Search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full px-6 py-3 rounded-full bg-gray-100 focus:outline-none"
        />
      </div>

      <div className="flex flex-wrap justify-center gap-2 mb-10">
        {FILTERS.map((filter) => (
          <button
            key={filter.key}
            onClick={() =>
              setActiveFilter(
                activeFilter === filter.value ? null : filter.value
              )
            }
            className={`px-4 py-1 rounded-full border text-sm transition ${
              activeFilter === filter.value
                ? "bg-teal-200 border-teal-300"
                : "bg-white hover:bg-gray-100"
            }`}
          >
            {filter.label}
          </button>
        ))}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-12 max-w-6xl mx-auto">
        {filteredMedia.map((item) => (
          <MediaCard
            key={item.moduleId}
            item={item}
            onDelete={handleDelete}
            onRename={handleRename}
          />
        ))}
      </div>

      <button
        onClick={() => navigate("/upload")}
        className="fixed bottom-6 right-6 bg-teal-500 hover:bg-teal-600 text-white p-4 rounded-full shadow-lg flex items-center justify-center"
      >
        <Plus className="h-6 w-6" />
      </button>
    </div>
  );
}

function MediaCard({
  item,
  onDelete,
  onRename,
}: {
  item: MediaItem;
  onDelete: (metadataId: number, moduleId: number, type: MediaType) => void;
  onRename: (id: number, type: MediaType, newName: string) => void;
}){
  const [isEditing, setIsEditing] = useState(false);
  const [tempName, setTempName] = useState(item.name);

  const handleBlurOrEnter = async () => {
    if (tempName.trim() && tempName !== item.name) {
      await onRename(item.id, item.type, tempName.trim());
    }
    setIsEditing(false);
  };

  return (
    <div className="relative flex flex-col items-center text-center">
      {item.type === "photo" ? (
        <img
          src={item.uri}
          alt={item.name}
          className="w-full aspect-video object-cover rounded mb-4"
        />
      ) : item.type === "video" ? (
        <div className="w-full aspect-video bg-gray-300 rounded mb-4 flex items-center justify-center">
          <span className="text-gray-700 font-medium">VIDEO</span>
        </div>
      ) : (
        <div className="w-full aspect-video bg-gray-200 rounded mb-4 flex items-center justify-center">
          <audio controls className="w-full px-4">
            <source src={item.uri} type="audio/mpeg" />
            Your browser does not support the audio element.
          </audio>
        </div>
      )}

      {/* Editable Name */}
      {isEditing ? (
        <div className="flex items-center gap-1">
          <input
            type="text"
            value={tempName}
            autoFocus
            onChange={(e) => setTempName(e.target.value)}
            onBlur={handleBlurOrEnter}
            onKeyDown={(e) => e.key === "Enter" && handleBlurOrEnter()}
            className="border-b border-teal-500 focus:outline-none text-lg font-medium text-center"
          />
          <button onClick={handleBlurOrEnter}>
            <Check className="h-4 w-4 text-teal-500" />
          </button>
        </div>
      ) : (
        <h3
          className="text-lg font-medium cursor-pointer hover:text-teal-500"
          onClick={() => setIsEditing(true)}
        >
          {item.name}
        </h3>
      )}

      <p className="text-sm text-gray-500">{item.filename}</p>

      {/* Delete Button */}
      <button
        onClick={() => onDelete(item.id, item.moduleId, item.type)}
        className="absolute bottom-2 right-2 bg-red-500 hover:bg-red-600 text-white p-1 rounded-full"
        title="Delete"
      >
        <Trash2 className="h-4 w-4" />
      </button>
    </div>
  );
}
