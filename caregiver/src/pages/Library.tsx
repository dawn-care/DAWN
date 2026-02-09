import { useEffect, useMemo, useState } from "react";
import { Trash2, Plus, Check } from "lucide-react";
import { useNavigate } from "react-router-dom";

type ModuleType =
  | "photo"
  | "video"
  | "audio"
  | "weather"
  | "news"
  | "message_prompt"
  | "caregiver_location_prompt"
  | "carepartner_location_prompt";

interface ModuleItem {
  module_id: number;
  metadata_id: number;
  type: ModuleType;

  // Media fields
  displayed_name?: string;
  filename?: string;
  uri?: string;
  extension?: string;

  // Widget fields
  zipcode?: string;
  category?: string;
  language?: string;
  country?: string;

  // Prompt fields
  message?: string;
  location?: string;
  image_uri?: string;
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
  { key: "widgets", label: "Widgets", value: "widget" },
  { key: "prompts", label: "Prompts", value: "prompt" },
];

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export default function Library() {
  const [modules, setModules] = useState<ModuleItem[]>([]);
  const [search, setSearch] = useState("");
  const [activeFilter, setActiveFilter] = useState<string | null>(null);
  const navigate = useNavigate();

  const caregiverID = 1; // temporary
  const carepartnerID = 1; // temporary

  // Fetch all metadata on mount
  useEffect(() => {
    async function fetchMetadata() {
      try {
        const url = new URL(`${API_URL}/modules/metadata/all`);
        url.searchParams.append("caregiver_id", caregiverID.toString());
        url.searchParams.append("carepartner_id", carepartnerID.toString());

        const res = await fetch(url.toString());
        if (!res.ok) {
          console.error("Failed to fetch metadata");
          return;
        }

        const data: ModuleItem[] = await res.json();
        setModules(data);
      } catch (err) {
        console.error("Error fetching metadata:", err);
      }
    }

    fetchMetadata();
  }, []);

  const handleDelete = async (moduleId: number) => {
    if (!confirm("Are you sure you want to delete?")) return;

    try {
      // const res = await fetch(`${API_URL}/module?id=${moduleId}`, {
      //   method: "DELETE",
      // });

      setModules((prev) => prev.filter((m) => m.module_id !== moduleId));

      // if (!res.ok) throw new Error("Failed to delete module");
    } catch (err) {
      console.error("Failed to delete module:", err);
    }
  };

  const handleRename = async (moduleId: number, newName: string) => {
    try {
      const res = await fetch(`${API_URL}/module/rename`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ module_id: moduleId, name: newName }),
      });

      if (!res.ok) throw new Error("Rename failed");

      setModules((prev) =>
        prev.map((m) =>
          m.module_id === moduleId
            ? { ...m, displayed_name: newName }
            : m
        )
      );
    } catch (err) {
      console.error("Rename failed:", err);
    }
  };

  const filteredModules = useMemo(() => {
    return modules.filter((item) => {
      const name =
        item.displayed_name || item.filename || item.message || "";

      const matchesSearch = name
        .toLowerCase()
        .includes(search.toLowerCase());

      if (!activeFilter) return matchesSearch;

      if (activeFilter === "widget") {
        return (
          (item.type === "weather" || item.type === "news") &&
          matchesSearch
        );
      }

      if (activeFilter === "prompt") {
        return item.type.includes("prompt") && matchesSearch;
      }

      return item.type === activeFilter && matchesSearch;
    });
  }, [modules, search, activeFilter]);

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
            className={`px-4 py-1 rounded-full border text-sm transition
              ${
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
        {filteredModules.map((item) => (
          <ModuleCard
            key={item.module_id}
            item={item}
            onDelete={handleDelete}
            onRename={handleRename}
          />
        ))}
      </div>

      <button
        onClick={() => navigate("/upload")}
        className="fixed bottom-6 right-6 bg-teal-500 hover:bg-teal-600 text-white p-4 rounded-full shadow-lg"
      >
        <Plus className="h-6 w-6" />
      </button>
    </div>
  );
}

function ModuleCard({
  item,
  onDelete,
  onRename,
}: {
  item: ModuleItem;
  onDelete: (moduleId: number) => void;
  onRename: (moduleId: number, newName: string) => void;
}) {
  const [isEditing, setIsEditing] = useState(false);
  const [tempName, setTempName] = useState(
    item.displayed_name || item.filename || item.message || ""
  );

  const handleSave = async () => {
    if (tempName.trim()) {
      await onRename(item.module_id, tempName.trim());
    }
    setIsEditing(false);
  };

  const renderContent = () => {
    switch (item.type) {
      case "photo":
        return (
          <img
            src={item.uri}
            alt={item.displayed_name}
            className="w-full aspect-video object-cover rounded mb-4"
          />
        );

      case "video":
        return (
          <video
            controls
            className="w-full aspect-video rounded mb-4"
            src={item.uri}
          />
        );

      case "audio":
        return (
          <audio controls className="w-full mb-4">
            <source src={item.uri} />
          </audio>
        );

      case "weather":
        return (
          <div className="w-full aspect-video bg-blue-100 rounded mb-4 flex items-center justify-center">
            Weather Widget ({item.zipcode})
          </div>
        );

      case "news":
        return (
          <div className="w-full aspect-video bg-yellow-100 rounded mb-4 flex items-center justify-center">
            News ({item.category})
          </div>
        );

      default:
        return (
          <div className="w-full aspect-video bg-gray-200 rounded mb-4 flex items-center justify-center text-center px-4">
            {item.message}
          </div>
        );
    }
  };

  return (
    <div className="relative flex flex-col items-center text-center">
      {renderContent()}

      {isEditing ? (
        <div className="flex items-center gap-1">
          <input
            value={tempName}
            autoFocus
            onChange={(e) => setTempName(e.target.value)}
            onBlur={handleSave}
            onKeyDown={(e) => e.key === "Enter" && handleSave()}
            className="border-b border-teal-500 focus:outline-none text-lg font-medium text-center"
          />
          <button onClick={handleSave}>
            <Check className="h-4 w-4 text-teal-500" />
          </button>
        </div>
      ) : (
        <h3
          className="text-lg font-medium cursor-pointer hover:text-teal-500"
          onClick={() => setIsEditing(true)}
        >
          {tempName}
        </h3>
      )}

      <button
        onClick={() => onDelete(item.module_id)}
        className="absolute bottom-2 right-2 bg-red-500 hover:bg-red-600 text-white p-1 rounded-full"
      >
        <Trash2 className="h-4 w-4" />
      </button>
    </div>
  );
}
