/**
 * Media.tsx
 *
 * User: Caregiver
 *
 * Photos, videos & audio page at `/media`: everything the caregiver has
 * added for the active carepartner.
 *
 *  - Header with three ways to add something: camera, voice, file upload
 *  - Filter pills (All / Photos / Videos / Audio) with counts
 *  - A three column grid of cards, newest first
 *  - A right-hand panel that shows the selected item, or the add flow that
 *    was picked in the header
 *
 * The panel contents are placeholders until the detail, camera, voice and
 * upload panels are built.
 */

import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import { Mic, Upload, Video, X } from "lucide-react";
import { useAuth } from "../context/AuthContext";
import IconButton from "../components/ui/IconButton";
import { FOCUS_RING } from "../components/ui/styles";
import { friendlyError, listMedia } from "../media/api";
import type { Media as MediaItem } from "../media/api";
import type { MediaKind } from "../media/constants";
import FilterPills from "../media/FilterPills";
import type { Pill } from "../media/FilterPills";
import MediaCard from "../media/MediaCard";

type Filter = "all" | MediaKind;
type Panel = "detail" | "camera" | "mic" | "file";

const PANEL_TITLES: Record<Exclude<Panel, "detail">, string> = {
  camera: "Use your camera",
  mic: "Record your voice",
  file: "Upload a file",
};

const FILTERS: { value: Filter; label: string; empty: string }[] = [
  { value: "all", label: "All", empty: "No media files added yet." },
  { value: "photo", label: "Photos", empty: "No photos yet." },
  { value: "video", label: "Videos", empty: "No videos yet." },
  { value: "audio", label: "Audio", empty: "No audio yet." },
];

/* ============================
   Header buttons
============================ */

interface AddButtonProps {
  icon: ReactNode;
  title: string;
  subtitle: string;
  open: boolean;
  onClick: () => void;
}

function AddButton({ icon, title, subtitle, open, onClick }: AddButtonProps) {
  return (
    <button
      type="button"
      aria-pressed={open}
      onClick={onClick}
      className={`h-[76px] flex items-center gap-3.5 px-[18px] rounded-[14px] border-2 text-left text-dawn-ink cursor-pointer ${
        open ? "bg-dawn-accent/10 border-dawn-accent" : "bg-white border-dawn-line-strong"
      } ${FOCUS_RING}`}
    >
      <span className="shrink-0 text-dawn-accent">{icon}</span>
      <span className="flex flex-col gap-0.5">
        <span className="text-lg font-extrabold">{title}</span>
        <span className="text-[15px] text-dawn-muted">{subtitle}</span>
      </span>
    </button>
  );
}

/* ============================
   Component
============================ */

export default function Media() {
  const { activeCarePartnerId } = useAuth();

  const [items, setItems] = useState<MediaItem[] | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  // Bumped to fetch the list again (the "Try again" button)
  const [reloadKey, setReloadKey] = useState(0);

  const [filter, setFilter] = useState<Filter>("all");
  const [panel, setPanel] = useState<Panel>("detail");
  const [selectedId, setSelectedId] = useState<number | null>(null);

  /* ============================
     Effects
  ============================ */

  useEffect(() => {
    if (activeCarePartnerId === null) return;

    let cancelled = false;
    listMedia(activeCarePartnerId)
      .then((list) => {
        if (cancelled) return;
        setItems(list);
        setLoadError(null);
      })
      .catch((err) => {
        if (!cancelled) setLoadError(friendlyError(err));
      });

    return () => {
      cancelled = true;
    };
  }, [activeCarePartnerId, reloadKey]);

  /* ============================
     Derived Values
  ============================ */

  const all = items ?? [];
  const visible = filter === "all" ? all : all.filter((item) => item.kind === filter);
  // Nothing picked yet, or the picked item is filtered out or gone: show the first one
  const selected = visible.find((item) => item.id === selectedId) ?? visible[0] ?? null;

  const pills: Pill<Filter>[] = FILTERS.map(({ value, label }) => ({
    value,
    label,
    count: value === "all" ? all.length : all.filter((item) => item.kind === value).length,
  }));

  const retry = () => {
    setLoadError(null);
    setReloadKey((key) => key + 1);
  };

  /* ============================
     Render
  ============================ */

  return (
    <div className="h-screen flex flex-col gap-3 px-7 py-5">
      <h1 className="m-0 min-h-11 flex items-center text-3xl font-extrabold">Photos, videos &amp; audio</h1>

      <div className="grid grid-cols-3 gap-3">
        <AddButton
          icon={<Video size={34} aria-hidden="true" />}
          title="Use your camera"
          subtitle="Take a photo or record a video"
          open={panel === "camera"}
          onClick={() => setPanel("camera")}
        />
        <AddButton
          icon={<Mic size={34} aria-hidden="true" />}
          title="Record your voice"
          subtitle="A message or a song"
          open={panel === "mic"}
          onClick={() => setPanel("mic")}
        />
        <AddButton
          icon={<Upload size={34} aria-hidden="true" />}
          title="Upload a file"
          subtitle="Photos, videos, audio or music"
          open={panel === "file"}
          onClick={() => setPanel("file")}
        />
      </div>

      <div className="flex gap-6 grow min-h-0">
        <section aria-label="Your media" className="grow min-w-0 flex flex-col gap-3">
          <FilterPills pills={pills} value={filter} onChange={setFilter} label="Show" />

          {/* p-[3px] leaves room for the selected card's border, which the scroll area would clip */}
          <div className="grow min-h-0 overflow-y-auto p-[3px]">
            {loadError ? (
              <div role="alert" className="flex flex-col items-start gap-2 text-base">
                <p className="m-0">{loadError}</p>
                <button
                  type="button"
                  onClick={retry}
                  className={`h-11 px-4 rounded-[10px] border-2 border-dawn-control bg-white font-bold text-dawn-ink cursor-pointer ${FOCUS_RING}`}
                >
                  Try again
                </button>
              </div>
            ) : items === null ? (
              <p className="m-0 text-base text-dawn-muted">Loading…</p>
            ) : visible.length === 0 ? (
              <p className="m-0 text-base text-dawn-muted">{FILTERS.find((f) => f.value === filter)?.empty}</p>
            ) : (
              <div className="grid grid-cols-3 gap-3 content-start">
                {visible.map((item) => (
                  <MediaCard
                    key={item.id}
                    media={item}
                    selected={panel === "detail" && item.id === selected?.id}
                    onSelect={() => {
                      setSelectedId(item.id);
                      setPanel("detail");
                    }}
                  />
                ))}
              </div>
            )}
          </div>
        </section>

        <aside
          aria-label={panel === "detail" ? "Selected item" : PANEL_TITLES[panel]}
          className="w-[420px] shrink-0 overflow-y-auto flex flex-col gap-3 p-[18px] bg-white border-[1.5px] border-dawn-line rounded-2xl"
        >
          {panel === "detail" ? (
            selected ? (
              <h2 className="m-0 text-xl font-extrabold break-words">{selected.display_name}</h2>
            ) : items === null ? null : (
              <p className="m-auto text-center text-2xl font-bold text-dawn-muted">
                {all.length === 0 ? "Add a media file to see it here." : "Pick something to see it here."}
              </p>
            )
          ) : (
            <div className="flex items-center justify-between">
              <h2 className="m-0 text-xl font-extrabold">{PANEL_TITLES[panel]}</h2>
              <IconButton
                label="Close"
                icon={<X size={20} strokeWidth={2.4} aria-hidden="true" />}
                onClick={() => setPanel("detail")}
              />
            </div>
          )}
        </aside>
      </div>
    </div>
  );
}
