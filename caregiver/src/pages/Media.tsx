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
 *  - A grid of cards (3 columns, 2 beside the panel on wide screens), newest first
 *  - A right-hand panel that shows the selected item, or the add flow that
 *    was picked in the header
 *
 * Camera and voice recording are switched off for now, so uploading a file is
 * the only way to add something. Files picked in the upload panel are checked
 * and queued, then named and saved one at a time. Saved items go to the top
 * of the grid with a "New" badge, selected, with a toast.
 */

import { useEffect, useRef, useState } from "react";
import type { ReactNode, Ref } from "react";
import { Mic, Upload, Video } from "lucide-react";
import { useAuth } from "../context/AuthContext";
import { FOCUS_RING } from "../components/ui/styles";
import { useToast } from "../components/ui/toastContext";
import {
  ApiError,
  deleteMedia,
  deleteMediaOnExit,
  friendlyError,
  listFamilyMembers,
  listMedia,
  listScheduledItems,
} from "../media/api";
import type { FamilyMember, Media as MediaItem, ScheduledItem } from "../media/api";
import type { MediaKind } from "../media/constants";
import DetailPanel from "../media/DetailPanel";
import FilterPills from "../media/FilterPills";
import type { Pill } from "../media/FilterPills";
import MediaCard from "../media/MediaCard";
import type { PendingFile } from "../media/files";
import SavePanel from "../media/SavePanel";
import UploadPanel from "../media/UploadPanel";

// Items the caregiver removed whose Undo toast is still open. The server is
// told to delete them only when the toast closes, so they stay hidden from
// every list load until then. Module level, because the toast outlives this page.
// If the page is closed or reloaded first, the pagehide listener below sends
// the deletes then, so a removed item doesn't come back.
const pendingRemovalIds = new Set<number>();

window.addEventListener("pagehide", () => pendingRemovalIds.forEach(deleteMediaOnExit));

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
  ref?: Ref<HTMLButtonElement>;
}

function AddButton({ icon, title, subtitle, open, disabled = false, onClick, ref }: AddButtonProps) {
  return (
    <button
      ref={ref}
      type="button"
      aria-pressed={open}
      disabled={disabled}
      onClick={onClick}
      className={`min-h-[76px] flex items-center gap-3.5 px-[18px] py-2 rounded-[14px] border-2 text-left text-dawn-ink ${
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
  // Null if the family could not be loaded; media can still be saved, just not linked to anyone
  const [family, setFamily] = useState<FamilyMember[] | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  // Bumped to fetch the list again (the "Try again" button)
  const [reloadKey, setReloadKey] = useState(0);

  const [filter, setFilter] = useState<Filter>("all");
  const [panel, setPanel] = useState<Panel>("detail");
  const [selectedId, setSelectedId] = useState<number | null>(null);
  // Files that passed the upload checks and are waiting to be named and saved, plus a message for each that didn't pass
  const [queue, setQueue] = useState<PendingFile[]>([]);
  const [notices, setNotices] = useState<string[]>([]);
  // How many files were picked together, for "File 2 of 5"
  const [batchSize, setBatchSize] = useState(0);
  // Saved in this session of the page; shown with a "New" badge until the caregiver picks something else
  const [newIds, setNewIds] = useState<number[]>([]);
  // The item whose detail panel says "Saved. You can use it any time."
  const [justSavedId, setJustSavedId] = useState<number | null>(null);
  // What has been saved from the files picked together, for one toast when they are done
  const savedInBatch = useRef<MediaItem[]>([]);
  // Focus goes back here when the upload panel is closed
  const uploadButton = useRef<HTMLButtonElement>(null);
  // The carepartner whose list is on screen; null until it has loaded
  const loadedFor = useRef<number | null>(null);

  /* ============================
     Effects
  ============================ */

  useEffect(() => {
    if (activeCarePartnerId === null) return;

    // Another carepartner's list must not show while this one loads
    if (loadedFor.current !== activeCarePartnerId) {
      loadedFor.current = null;
      setItems(null);
    }

    let cancelled = false;
    // The schedule and the family only add detail, so a failure there must not hide the media
    Promise.all([
      listMedia(activeCarePartnerId),
      listScheduledItems(activeCarePartnerId).catch((err) => {
        console.warn("Could not load the schedule; showing plainer schedule text", err);
        return null;
      }),
      listFamilyMembers(activeCarePartnerId).catch((err) => {
        console.warn("Could not load the family; media can't be linked to anyone for now", err);
        return null;
      }),
    ])
      .then(([list, schedule, members]) => {
        if (cancelled) return;
        loadedFor.current = activeCarePartnerId;
        setItems(list.filter((item) => !pendingRemovalIds.has(item.id)));
        setScheduled(schedule);
        setFamily(members);
        setLoadError(null);
      })
      .catch((err) => {
        if (cancelled) return;
        // A refresh that fails leaves what is already on screen; only a first load has nothing to show
        if (loadedFor.current === activeCarePartnerId) {
          toast.show(`Couldn't refresh the list. ${friendlyError(err)}`, {
            actionLabel: "Try again",
            onAction: () => setReloadKey((key) => key + 1),
          });
        } else {
          setLoadError(friendlyError(err));
        }
      });

    return () => {
      cancelled = true;
    };
  }, [activeCarePartnerId, reloadKey, toast]);

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

  const handleTagsChanged = (updated: MediaItem) => {
    setItems((list) => list && list.map((item) => (item.id === updated.id ? updated : item)));
    toast.show(`Saved how staff find "${updated.display_name}".`);
  };

  // Added from either panel's picker; it was created on the server already
  const handleFamilyAdded = (member: FamilyMember) => {
    setFamily((members) => [...(members ?? []), member]);
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
    savedInBatch.current = [];
    setBatchSize(accepted.length);
    setQueue(accepted);
    setNotices(rejected);
  };

  // One toast for everything saved from the files picked together
  const announceSaved = () => {
    const saved = savedInBatch.current;
    savedInBatch.current = [];
    if (saved.length === 0) return;
    toast.show(saved.length === 1 ? `Added "${saved[0].display_name}".` : `Added ${saved.length} files.`);
  };

  // Back to the empty upload panel, ending the queue
  const clearQueue = () => {
    announceSaved();
    setQueue([]);
    setNotices([]);
  };

  const closeUpload = () => {
    clearQueue();
    setPanel("detail");
    uploadButton.current?.focus();
  };

  // Moves on to the next picked file. After the last one: to the saved item if anything was
  // saved, otherwise back to the upload panel.
  const nextInQueue = () => {
    const rest = queue.slice(1);
    setQueue(rest);
    if (rest.length > 0) return;

    const savedAny = savedInBatch.current.length > 0;
    announceSaved();
    setNotices([]);
    if (savedAny) setPanel("detail");
  };

  const handleSaved = (saved: MediaItem) => {
    savedInBatch.current.push(saved);
    // Newest first, like the server's list
    setItems((list) => [saved, ...(list ?? [])]);
    setNewIds((ids) => [...ids, saved.id]);
    setFilter("all");
    setSelectedId(saved.id);
    setJustSavedId(saved.id);
    nextInQueue();
  };

  const selectItem = (id: number) => {
    setSelectedId(id);
    setPanel("detail");
    setNewIds([]);
    setJustSavedId(null);
  };

  const retry = () => {
    setLoadError(null);
    setReloadKey((key) => key + 1);
  };

  /* ============================
     Render
  ============================ */

  return (
    // Wide screens: the page fits the window and the grid and panel scroll on their own. Narrower: the panel sits below the grid and the page scrolls.
    <div className="lg:h-screen flex flex-col gap-3 px-7 py-5">
      <h1 className="m-0 min-h-11 flex items-center text-3xl font-extrabold">Photos, videos &amp; audio</h1>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
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
          ref={uploadButton}
        />
      </div>

      <div className="flex flex-col lg:flex-row gap-6 grow min-h-0">
        <section aria-label="Your media" className="grow min-w-0 flex flex-col gap-3">
          <FilterPills pills={pills} value={filter} onChange={setFilter} label="Show" />

          {/* p-[3px] leaves room for the selected card's border, which the scroll area would clip */}
          <div className="grow min-h-0 lg:overflow-y-auto p-[3px]">
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
              <p role="status" className="m-0 text-base text-dawn-muted">
                Loading…
              </p>
            ) : visible.length === 0 ? (
              <p className="m-0 text-base text-dawn-muted">{FILTERS.find((f) => f.value === filter)?.empty}</p>
            ) : (
              <div className="grid grid-cols-3 lg:grid-cols-2 xl:grid-cols-3 gap-3 content-start">
                {visible.map((item) => (
                  <MediaCard
                    key={item.id}
                    media={item}
                    selected={panel === "detail" && item.id === selected?.id}
                    isNew={newIds.includes(item.id)}
                    onSelect={() => selectItem(item.id)}
                  />
                ))}
              </div>
            )}
          </div>
        </section>

        <aside
          aria-label={panel === "detail" ? "Selected item" : "Upload a file"}
          className="w-full lg:w-[420px] shrink-0 lg:overflow-y-auto flex flex-col gap-3 p-[18px] bg-white border-[1.5px] border-dawn-line rounded-2xl"
        >
          {panel === "detail" ? (
            selected ? (
              <DetailPanel
                key={selected.id}
                media={selected}
                scheduled={scheduled}
                family={family}
                onFamilyAdded={handleFamilyAdded}
                justSaved={selected.id === justSavedId}
                onRenamed={handleRenamed}
                onTagsChanged={handleTagsChanged}
                onRemoved={() => handleRemoved(selected)}
              />
            ) : items === null ? null : (
              <p className="m-auto text-center text-2xl font-bold text-dawn-muted">
                {all.length === 0 ? "Add a media file to see it here." : "Pick something to see it here."}
              </p>
            )
          ) : queue.length === 0 ? (
            <UploadPanel onClose={closeUpload} onChosen={handleChosen} />
          ) : activeCarePartnerId !== null ? (
            <SavePanel
              key={queue[0].id}
              carepartnerId={activeCarePartnerId}
              pending={queue[0]}
              family={family}
              onFamilyAdded={handleFamilyAdded}
              position={batchSize - queue.length + 1}
              total={batchSize}
              notices={notices}
              onSaved={handleSaved}
              onSkip={nextInQueue}
              onClose={clearQueue}
            />
          ) : null}
        </aside>
      </div>
    </div>
  );
}
