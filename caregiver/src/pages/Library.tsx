import { useEffect, useMemo, useState } from "react";
import { Play } from "lucide-react";

type MediaType = "video" | "photo" | "audio";

interface MediaItem {
  id: number;
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

const MOCK_MEDIA: MediaItem[] = [
  { id: 1, name: "Sunset Beach", filename: "sunset.png", type: "photo", uri: "/temp_media/Sunset.png" },
  { id: 2, name: "Mountain View", filename: "mountain.jpg", type: "photo", uri: "/temp_media/Mountain.jpg" },
  { id: 3, name: "City Skyline", filename: "city.png", type: "photo", uri: "/temp_media/Skyline.png" },
  { id: 5, name: "Fly Me to the Moon", filename: "FrankSinatraMusic.mp3", type: "audio", uri: "/temp_media/Fly Me to the Moon.mp3" },
];

export default function Library() {
  const [media, setMedia] = useState<MediaItem[]>([]);
  const [search, setSearch] = useState("");
  const [activeFilter, setActiveFilter] = useState<string | null>(null);

  useEffect(() => {
    setMedia(MOCK_MEDIA);
  }, []);

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
    <div className="w-full px-8 py-10">
      <h1 className="text-4xl font-medium text-center mb-8">
        Your Uploaded Content
      </h1>

      {/* Search */}
      <div className="max-w-3xl mx-auto mb-4">
        <input
          type="text"
          placeholder="Search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full px-6 py-3 rounded-full bg-gray-100 focus:outline-none"
        />
      </div>

      {/* Filters */}
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

      {/* Media Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-12 max-w-6xl mx-auto">
        {filteredMedia.map((item) => (
          <MediaCard key={item.id} item={item} />
        ))}
      </div>
    </div>
  );
}

function MediaCard({ item }: { item: MediaItem }) {
  return (
    <div className="flex flex-col items-center text-center">
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

      <h3 className="text-lg font-medium">{item.name}</h3>
      <p className="text-sm text-gray-500">{item.filename}</p>
    </div>
  );
}
