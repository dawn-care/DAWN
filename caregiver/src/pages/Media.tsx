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
 * Camera and voice recording are switched off for now, so uploading a file is
 * the only way to add something. Files picked in the upload panel are checked
 * and queued; naming and saving them is not built yet, so the queue is only
 * listed.
 */

import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import { Mic, Upload, Video, X } from "lucide-react";
import { useAuth } from "../context/AuthContext";
import IconButton from "../components/ui/IconButton";
import { FOCUS_RING } from "../components/ui/styles";
import { useToast } from "../components/ui/toastContext";
import { ApiError, deleteMedia, friendlyError, listMedia, listScheduledItems } from "../media/api";
import type { Media as MediaItem, ScheduledItem } from "../media/api";
import type { MediaKind } from "../media/constants";
import DetailPanel from "../media/DetailPanel";
import FilterPills from "../media/FilterPills";
import type { Pill } from "../media/FilterPills";
import MediaCard from "../media/MediaCard";
import type { PendingFile } from "../media/files";
import { formatDuration } from "../media/labels";
import UploadPanel from "../media/UploadPanel";

// Items the caregiver removed whose Undo toast is still open. The server is
// told to delete them only when the toast closes, so they stay hidden from
// every list load until then. Module level, because the toast outlives this page.
const pendingRemovalIds = new Set<number>();

type Filter = "all" | MediaKind;
type Panel = "detail" | "file";

const FILTERS: { value: Filter; label: string; empty: string }[] = [
  { value: "all", label: "All", empty: "No media files added yet." },
  { value: "photo", label: "Photos", empty: "No photos yet." },
  { value: "video", label: "Videos", empty: "No videos yet." },
  { value: "audio", label: "Audio", empty: "No audio yet." },
];

/* ============================
   Removing
============================ */

// The Undo toast has closed without Undo: delete for real. A failure brings the item back, with a toast saying so.
async function commitRemoval(
  removed: MediaItem,
  showToast: (text: string) => void,
  reload: () => void
) {
  try {
    await deleteMedia(removed.id);
  } catch (err) {
    // Already gone (deleted somewhere else) is the outcome the caregiver wanted
    if (!(err instanceof ApiError && err.status === 404)) {
      showToast(`Couldn't remove "${removed.display_name}". ${friendlyError(err)}`);
    }
  } finally {
    pendingRemovalIds.delete(removed.id);
    // Show the server's answer: the item is back if the delete failed, and the schedule reflects what went with it
    reload();
  }
}

/* ============================
   Header buttons
============================ */

interface AddButtonProps {
  icon: ReactNode;
  title: string;
  subtitle: string;
  open: boolean;
  disabled?: boolean;
  onClick: () => void;
}

function AddButton({ icon, title, subtitle, open, disabled = false, onClick }: AddButtonProps) {
  return (
    <button
      type="button"
      aria-pressed={open}
      disabled={disabled}
      onClick={onClick}
      className={`h-[76px] flex items-center gap-3.5 px-[18px] rounded-[14px] border-2 text-left text-dawn-ink ${
        disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"
      } ${open ? "bg-dawn-accent/10 border-dawn-accent" : "bg-white border-dawn-line-strong"} ${FOCUS_RING}`}
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
  const toast = useToast();

  const [items, setItems] = useState<MediaItem[] | null>(null);
  // Null if the schedule could not be loaded; the page still works, with plainer schedule text
  const [scheduled, setScheduled] = useState<ScheduledItem[] | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  // Bumped to fetch the list again (the "Try again" button)
  const [reloadKey, setReloadKey] = useState(0);

  const [filter, setFilter] = useState<Filter>("all");
  const [panel, setPanel] = useState<Panel>("detail");
  const [selectedId, setSelectedId] = useState<number | null>(null);
  // Files that passed the upload checks and are waiting to be named and saved, plus a message for each that didn't pass
  const [queue, setQueue] = useState<PendingFile[]>([]);
  const [notices, setNotices] = useState<string[]>([]);

  /* ============================
     Effects
  ============================ */

  useEffect(() => {
    if (activeCarePartnerId === null) return;

    let cancelled = false;
    // The schedule only adds detail, so a failure there must not hide the media
    Promise.all([listMedia(activeCarePartnerId), listScheduledItems(activeCarePartnerId).catch(() => null)])
      .then(([list, schedule]) => {
        if (cancelled) return;
        setItems(list.filter((item) => !pendingRemovalIds.has(item.id)));
        setScheduled(schedule);
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

  const handleRenamed = (updated: MediaItem) => {
    setItems((list) => list && list.map((item) => (item.id === updated.id ? updated : item)));
    toast.show(`Renamed to "${updated.display_name}".`);
  };

  // Hide it from the grid right away and select its neighbour. The server
  // isn't told to delete it until the Undo toast closes, so Undo only has to
  // cancel that and show the item again.
  const handleRemoved = (removed: MediaItem) => {
    const index = visible.findIndex((item) => item.id === removed.id);
    const neighbour = visible[index + 1] ?? visible[index - 1] ?? null;
    const reload = () => setReloadKey((key) => key + 1);
    const showToast = toast.show;

    pendingRemovalIds.add(removed.id);
    setItems((list) => list && list.filter((item) => item.id !== removed.id));
    setSelectedId(neighbour?.id ?? null);

    toast.show(`Removed "${removed.display_name}".`, {
      actionLabel: "Undo",
      onAction: () => {
        pendingRemovalIds.delete(removed.id);
        setSelectedId(removed.id);
        reload();
      },
      onDismiss: () => void commitRemoval(removed, showToast, reload),
    });
  };

  const handleChosen = (accepted: PendingFile[], rejected: string[]) => {
    setQueue(accepted);
    setNotices(rejected);
  };

  // Back to the empty upload panel, for closing the queue
  const clearQueue = () => {
    setQueue([]);
    setNotices([]);
  };

  const closeUpload = () => {
    clearQueue();
    setPanel("detail");
  };

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
          subtitle="Coming soon"
          open={false}
          disabled
          onClick={() => {}}
        />
        <AddButton
          icon={<Mic size={34} aria-hidden="true" />}
          title="Record your voice"
          subtitle="Coming soon"
          open={false}
          disabled
          onClick={() => {}}
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
          aria-label={panel === "detail" ? "Selected item" : "Upload a file"}
          className="w-[420px] shrink-0 overflow-y-auto flex flex-col gap-3 p-[18px] bg-white border-[1.5px] border-dawn-line rounded-2xl"
        >
          {panel === "detail" ? (
            selected ? (
              <DetailPanel
                key={selected.id}
                media={selected}
                scheduled={scheduled}
                onRenamed={handleRenamed}
                onRemoved={() => handleRemoved(selected)}
              />
            ) : items === null ? null : (
              <p className="m-auto text-center text-2xl font-bold text-dawn-muted">
                {all.length === 0 ? "Add a media file to see it here." : "Pick something to see it here."}
              </p>
            )
          ) : queue.length === 0 ? (
            <UploadPanel onClose={closeUpload} onChosen={handleChosen} />
          ) : (
            <>
              <div className="flex items-center justify-between">
                <h2 className="m-0 text-xl font-extrabold">Ready to save</h2>
                <IconButton
                  label="Close"
                  icon={<X size={20} strokeWidth={2.4} aria-hidden="true" />}
                  onClick={clearQueue}
                />
              </div>
              {notices.length > 0 && (
                <ul role="alert" className="m-0 pl-5 flex flex-col gap-1 text-base font-bold text-dawn-accent">
                  {notices.map((message) => (
                    <li key={message}>{message}</li>
                  ))}
                </ul>
              )}
              <ul className="m-0 pl-5 flex flex-col gap-1 text-base">
                {queue.map((pending) => (
                  <li key={pending.id} className="break-words">
                    {pending.file.name} ({pending.kind}
                    {pending.seconds !== null && `, ${formatDuration(pending.seconds)}`})
                  </li>
                ))}
              </ul>
            </>
          )}
        </aside>
      </div>
    </div>
  );
}
