import { useEffect, useMemo, useState } from "react";
import { Trash2, Plus, Check } from "lucide-react";
import { useNavigate } from "react-router-dom";
import WidgetForm from "./WidgetForm";
import PromptForm from "./PromptForm";
// import type { Module } from "../components/AgendaList";

type ModuleType = "photo" | "video" | "audio" | "weather" | "sports" | "news" | "messages";
type MediaType = "video" | "photo" | "audio";

interface ModuleItem {
  id: number;
  caregiver_id: number;
  carepartner_id: number;
  module_type: ModuleType;
  metadata_id: number;
}

interface MediaItem extends ModuleItem {
  name: string;
  filename: string;
  type: MediaType;
  uri: string;
}

interface WeatherItem extends ModuleItem {
  module_type: "weather";
  zipcode: string;
}

interface NewsItem extends ModuleItem {
  module_type: "news" | "sports";
  category: string;
  language: string;
  country: string;
}

interface PromptItem extends ModuleItem {
  type: "messages";
  message: string;
  image_uri?: string;
}

type WidgetItem = WeatherItem | NewsItem;

type Filter = {
  key: string;
  label: string;
  value: string;
};

const FILTERS: Filter[] = [
  { key: "all", label: "All", value: "" },
  { key: "pictures", label: "Pictures", value: "photo" },
  { key: "videos", label: "Videos", value: "video" },
  { key: "music", label: "Music", value: "audio" },
  { key: "weather", label: "Weather", value: "weather" },
  { key: "sports", label: "Sports News", value: "sports" },
  { key: "news", label: "Good News", value: "news" },
  { key: "messages", label: "Messages", value: "messages" },
  { key: "png", label: ".png", value: ".png" },
  { key: "jpg", label: ".jpg", value: ".jpg" },
  { key: "mov", label: ".mov", value: ".mov" },
  { key: "wav", label: ".wav", value: ".wav" },
  { key: "mp3", label: ".mp3", value: ".mp3" },
  { key: "ogg", label: ".ogg", value: ".ogg" },
];

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

// Props for Library when it is acting as the agenda selector
interface LibraryProps {
  onSelect?: (module: { moduleId: number; moduleType: string; label: string }) => void;
}

export default function Library({ onSelect }: LibraryProps = {}) {
  const [modules, setModules] = useState<ModuleItem[]>([]);
  const [search, setSearch] = useState("");
  const [activeFilter, setActiveFilter] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetchModules();
  }, []);

  const fetchModules = async () => {
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

      // Fetch metadata for each module
      const moduleResults = await Promise.all(
        modules.map(async (m: any) => {
          const type = m.module_type as ModuleType;

          const res = await fetch(`${API_URL}/${type}?id=${m.metadata_id}`);
          if (!res.ok) return null;

          const data = await res.json();
          if (type === "video" || type === "photo" || type === "audio") {
            return {
              id: m.id,
              caregiver_id: 1,
              carepartner_id: 1,
              metadata_id: m.metadata_id,
              module_type: type,
              name: data.displayed_name || data.filename,
              filename: data.filename,
              type: type,
              uri: `${API_URL}/${type}/file?id=${data.id}`,
            } as MediaItem;
          } else if (type === "weather") {
            return {
              id: m.id,
              caregiver_id: 1,
              carepartner_id: 1,
              metadata_id: m.metadata_id,
              module_type: "weather",
              type: "weather",
              zipcode: data.zipcode,
            } as WeatherItem;
          } else if (type === "news" || type === "sports") {
            return {
              id: m.id,
              caregiver_id: 1,
              carepartner_id: 1,
              metadata_id: m.metadata_id,
              module_type: type,
              type: type,
              category: data.category,
              language: data.language,
              country: data.country,
            } as NewsItem;
          } else {
            return {
              id: m.id,
              caregiver_id: 1,
              carepartner_id: 1,
              metadata_id: m.metadata_id,
              module_type: "messages",
              type: "messages",
              message: data.message,
              image_uri: data.image_uri ? `${API_URL}${data.image_uri}` : null,
            } as PromptItem;
          }
        })
      );

      setModules(moduleResults.filter(Boolean));
    } catch (err) {
      console.error("Failed to fetch modules:", err);
    }
  };

  const handleWidgetDelete = async (moduleId: number, metadataId: number, widgetType: string,  isEditing: boolean) => {
    if (!isEditing) {
      // If Delete is coming from the library delete button instead of the widget form, confirm deletion and delete both module and widget
      if (!confirm("Are you sure you want to delete this widget?")) return;
      const moduleEndpoint = `/modules/delete?id=${moduleId}`;

      const widgetEndpoint =
          widgetType === "weather"
              ? `/weather/delete?id=${metadataId}`
              : widgetType === "sports"
              ? `/sports/delete?id=${metadataId}`
              : `/news/delete?id=${metadataId}`;
      try{
          const moduleResponse = await fetch(`${API_URL}${moduleEndpoint}`, {
              method: "DELETE",
          });

          if (!moduleResponse.ok) throw new Error("Delete failed");

          const widgetResponse = await fetch(`${API_URL}${widgetEndpoint}`, {
              method: "DELETE",
          });

          if (!widgetResponse.ok) throw new Error("Delete failed");
      }
      catch (err: any) {
            return
      }
    }
    setModules((prev) => prev.filter((m) => m.id !== moduleId));
  };

  const handlePromptDelete = async (moduleId: number, metadataId: number, isEditing: boolean) => {
    if (!isEditing) {
      // If Delete is coming from the library delete button instead of the prompt form, confirm deletion and delete both module and prompt
      if (!confirm("Are you sure you want to delete this widget?")) return;

      try{
          const moduleRes = await fetch(`${API_URL}/modules/delete?id=${moduleId}`, {
            method: "DELETE",
          });

          if (!moduleRes.ok) throw new Error("Delete failed");

          const promptRes = await fetch(`${API_URL}/messages/delete/${metadataId}`, {
            method: "DELETE",
          });

          if (!promptRes.ok) throw new Error("Delete failed");
      }
      catch (err: any) {
            return
      }
    }
    setModules((prev) => prev.filter((m) => m.id !== moduleId));
  };

  const handleUpdate = () => {
    fetchModules();
  };

  const handleMediaDelete = async (
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

      setModules((prev) =>
        prev.filter((m) => m.id !== moduleId)
      );
    } catch (err) {
      console.error("Delete failed:", err);
    }
  };

  const handleMediaRename = async (
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

      setModules((prev) =>
        prev.map((m) =>
          m.metadata_id === metadataId ? { ...m, name: newName } : m
        )
      );
    } catch (err) {
      console.error("Rename failed:", err);
    }
  };

  const filteredModules = useMemo(() => {
    return modules.filter((item) => {
      if (item.module_type === "video" || item.module_type === "photo" || item.module_type === "audio") {
        var mediaItem = item as MediaItem;
        var matchesSearch =
          mediaItem.name.toLowerCase().includes(search.toLowerCase()) ||
          mediaItem.filename.toLowerCase().includes(search.toLowerCase());

        var matchesFilter = !activeFilter
          ? true
          : activeFilter.startsWith(".")
          ? mediaItem.filename.toLowerCase().endsWith(activeFilter)
          : item.module_type === activeFilter;

        return matchesSearch && matchesFilter;
      } else {
        const matchesFilter = !activeFilter
          ? true
          : item.module_type === activeFilter;

        return matchesFilter;
      }
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
        {filteredModules.map((item) =>
          item.module_type === "video" ||
          item.module_type === "photo" ||
          item.module_type === "audio" ? (
            <MediaCard
              key={item.id}
              item={item as MediaItem}
              onDelete={handleMediaDelete}
              onRename={handleMediaRename}
              onSelect={onSelect}
            />
          ) : item.module_type === "weather" ||
            item.module_type === "sports" ||
            item.module_type === "news" ? (
            <WidgetCard key={item.id} item={item as WidgetItem} onDelete={handleWidgetDelete} onUpdate={handleUpdate} onSelect={onSelect} />
          ) : item.module_type === "messages" ? (
            <PromptCard key={item.id} item={item as PromptItem} onDelete={handlePromptDelete} onUpdate={handleUpdate} onSelect={onSelect}/>
          ) : null
        )}
      </div>

      <button
        onClick={() => navigate("/media")}
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
  onSelect,
}: {
  item: MediaItem;
  onDelete: (metadataId: number, moduleId: number, type: MediaType) => void;
  onRename: (id: number, type: MediaType, newName: string) => void;
  onSelect?: (module: { moduleId: number; moduleType: string; label: string }) => void;
}){
  const [isEditing, setIsEditing] = useState(false);
  const [tempName, setTempName] = useState(item.name);

  const handleBlurOrEnter = async () => {
    if (tempName.trim() && tempName !== item.name) {
      await onRename(item.id, item.type, tempName.trim());
    }
    setIsEditing(false);
  };

  const handleClick = () => {
    if (onSelect) {
      onSelect({ moduleId: item.id, moduleType: item.module_type, label: `${item.type.charAt(0).toUpperCase() + item.type.slice(1)} - ${item.name}` });
    }
  };

  return (
    <div
      className={`relative flex flex-col items-center text-center border border-gray-200 rounded-xl shadow-sm hover:shadow-md p-6 w-full h-64 bg-white ${onSelect ? 'cursor-pointer ring-2 ring-transparent hover:ring-teal-400' : ''}`}
      onClick={onSelect ? handleClick : undefined}
    >
      {item.type === "photo" ? (
        <img
          src={item.uri}
          alt={item.name}
          className="w-full h-40 object-cover rounded mb-4"
        />
      ) : item.type === "video" ? (
        <div className="w-full h-40 object-cover rounded mb-4 bg-gray-200 flex items-center justify-center">
          <span className="text-gray-700 font-medium">VIDEO</span>
        </div>
      ) : (
        <div className="w-full h-40 object-cover rounded mb-4 bg-gray-200 flex items-center justify-center">
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
      {onSelect == undefined && (
        <button
          onClick={() => onDelete(item.metadata_id, item.id, item.type)}
          className="absolute bottom-2 right-2 bg-red-500 hover:bg-red-600 text-white p-1 rounded-full"
          title="Delete"
        >
          <Trash2 className="h-4 w-4" />
        </button>
      )}
    </div>
  );
}

function WidgetCard({
    item,
    onDelete,
    onUpdate,
    onSelect,
  }: {
    item: WidgetItem;
    onDelete: (moduleId: number, metadataId: number, moduleType: string, isEditing: boolean) => void;
    onUpdate: () => void;
    onSelect?: (module: { moduleId: number; moduleType: string; label: string }) => void;
  }) {
  const [isEditing, setIsEditing] = useState(false);

  return (
    <>
    {/* For editing widgets from the library */}
      {isEditing && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto relative">
            <button
              onClick={() => setIsEditing(false)}
              className="absolute top-4 right-4 text-gray-600 hover:text-gray-800 text-2xl z-10"
            >
              ✕
            </button>
            <div className="p-8">
              <WidgetForm
                moduleId={item.id}
                onDeleteSuccess={() => {
                  onDelete(item.id, item.metadata_id, item.module_type, isEditing);
                  setIsEditing(false);
                }}
                onUpdateSuccess={() => {
                  onUpdate();
                  setIsEditing(false);
                }}
              />
            </div>
          </div>
        </div>
      )}

    {/* For displaying widgets in the library */}
      <div className={`relative flex flex-col items-center justify-center text-center border border-gray-200 rounded-xl shadow-sm hover:shadow-md hover:text-teal-500 p-6 w-full h-64 bg-white ${onSelect ? 'cursor-pointer ring-2 ring-transparent hover:ring-teal-400' : ''}`}
          onClick={() => {
            if (onSelect) {
              const label = item.module_type === 'weather' ? `Weather - Zip: ${item.zipcode}` : item.module_type === 'sports' ? `Sports - ${(item as NewsItem).language}` : 'Good News';
              onSelect({ moduleId: item.id, moduleType: item.module_type, label });
            } else {
              setIsEditing(true);
            }
          }}>
        {item.module_type === "weather" ? (
          <>
            <h3
              className="text-3xl font-semibold cursor-pointer hover:text-teal-500"
              onClick={() => setIsEditing(true)}
            >
              Weather
            </h3>
            <p className="text-lg text-gray-500 mt-2">
              Zipcode: {item.zipcode}
            </p>
          </>
        ) : item.module_type === "sports" ? (
          <>
            <h3
              className="text-3xl font-semibold cursor-pointer hover:text-teal-500"
              onClick={() => setIsEditing(true)}
            >
              {item.country.toUpperCase()} News
            </h3>
            <p className="text-lg text-gray-500 mt-2">
              Team: {item.language}
            </p>
          </>
        ) : (
          <h3
            className="text-3xl font-semibold cursor-pointer hover:text-teal-500"
            onClick={() => setIsEditing(true)}
          >
            Good News
          </h3>
        )}
        {/* Delete Button */}
        {onSelect == undefined && (
          <button
            onClick={() => onDelete(item.id, item.metadata_id, item.module_type, isEditing)}
            className="absolute bottom-2 right-2 bg-red-500 hover:bg-red-600 text-white p-1 rounded-full"
            title="Delete"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        )}
      </div>
    </>
  );
}

function PromptCard({
    item,
    onDelete,
    onUpdate,
    onSelect,
  }: {
    item: PromptItem;
    onDelete: (moduleId: number, metadataId: number, isEditing: boolean) => void;
    onUpdate: () => void;
    onSelect?: (module: { moduleId: number; moduleType: string; label: string }) => void;
  }){
  const [isEditing, setIsEditing] = useState(false);

  return (
    <>
    {/* For editing prompts from the library */}
      {isEditing && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto relative">
            <button
              onClick={() => setIsEditing(false)}
              className="absolute top-4 right-4 text-gray-600 hover:text-gray-800 text-2xl z-10"
            >
              ✕
            </button>
            <div className="p-8">
              <PromptForm
                moduleId={item.id}
                onDeleteSuccess={() => {
                  onDelete(item.id, item.metadata_id, isEditing);
                  setIsEditing(false);
                }}
                onUpdateSuccess={() => {
                  onUpdate();
                  setIsEditing(false);
                }}
              />
            </div>
          </div>
        </div>
      )}

      {/* For displaying prompts in the library */}
      <div className={`relative flex flex-col items-center justify-center text-center border border-gray-200 rounded-xl shadow-sm hover:shadow-md hover:text-teal-500 p-6 w-full h-64 bg-white ${onSelect ? 'cursor-pointer ring-2 ring-transparent hover:ring-teal-400' : ''}`}
          onClick={() => {
            if (onSelect) {
              const label = item.message.length > 40 ? `Message - ${item.message.slice(0, 40)}...` : `Message - ${item.message}`;
              onSelect({ moduleId: item.id, moduleType: 'message', label });
            } else {
              setIsEditing(true);
            }
          }}>
          {item.image_uri ? (
            <>
              <img
                src={item.image_uri}
                alt={"Prompt Image"}
                className="w-full h-40 object-cover rounded mb-4"
                onClick={() => setIsEditing(true)}
              />
              <p className="text-sm text-gray-500">
                {item.message.length > 30
                  ? item.message.slice(0, 30) + "..."
                  : item.message}
              </p>
            </>
          ) : (
            <>
              <h3
                className="text-3xl font-semibold cursor-pointer hover:text-teal-500"
                onClick={() => setIsEditing(true)}
              >
                Message
              </h3>
              <p className="text-lg text-gray-500 mt-2">
                {item.message.length > 40
                  ? item.message.slice(0, 40) + "..."
                  : item.message}
              </p>
            </>
          )}
          {/* Delete Button */}
          {onSelect == undefined && (
            <button
              onClick={() => onDelete(item.id, item.metadata_id, isEditing)}
              className="absolute bottom-2 right-2 bg-red-500 hover:bg-red-600 text-white p-1 rounded-full"
              title="Delete"
            >
              <Trash2 className="h-4 w-4" />
          </button>
        )}
      </div>
    </>
  );
}