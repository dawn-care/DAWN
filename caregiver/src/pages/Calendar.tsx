/**
 * Calendar.tsx
 *
 * User: Caregiver
 *
 * Calendar page at `/calendar`. Shows the videos and audio scheduled to play
 * on the carepartner's tablet as a day, week (the default) or month calendar
 * (ScheduleCalendar), with the carepartner's asleep hours hatched.
 *
 * The page owns the toolbar (previous, next, the dates showing, the info
 * icon and the view switch) and loads what the page needs: the awake hours,
 * the media library the form picks from, and the scheduled occurrences for
 * whichever dates are showing.
 */

import type { DateClickArg } from "@fullcalendar/interaction";
import type FullCalendar from "@fullcalendar/react";
import { dateLabel, dayName, isAwakeAt, minuteOfDay, parseDate } from "@shared";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useToast } from "../components/ui/toastContext";
import IconButton from "../components/ui/IconButton";
import InfoPopover from "../components/ui/InfoPopover";
import SegmentedControl from "../components/ui/form/SegmentedControl";
import { FOCUS_RING } from "../components/ui/styles";
import { useAuth } from "../context/AuthContext";
import {
  ApiError,
  deleteScheduledItem,
  deleteScheduledItemOnExit,
  friendlyError,
  getCalendar,
  getScheduledItem,
  getScreenSettings,
  isAbort,
} from "../calendar/api";
import type { DeleteScope, Occurrence, ScheduledItem, ScreenSettings } from "../calendar/api";
import CalendarSidebar from "../calendar/CalendarSidebar";
import ScheduleCalendar from "../calendar/ScheduleCalendar";
import type { CalendarViewName, ClickedItem, VisibleRange } from "../calendar/ScheduleCalendar";
import ScheduleForm from "../calendar/ScheduleForm";
import type { DraftSlot, Removal } from "../calendar/ScheduleForm";
import { firstOpenSlot, valuesForEdit, valuesForNew } from "../calendar/formRules";
import type { EditTarget, FormValues } from "../calendar/formRules";
import { awakeHoursOf, dateText, minuteLabel } from "../calendar/time";
import { listMedia } from "../media/api";
import type { Media } from "../media/api";
import { describeScheduledItem } from "../media/schedule";

const PAGE_INFO =
  "Schedule videos and audio to play on the carepartner's tablet at set times. Shaded hours are when the carepartner is asleep, and nothing plays then.";

// Deletes whose Undo toast is still open, keyed "item|series" or "item|date". The server is
// told only when the toast closes, so until then they stay hidden from the calendar.
// Module level, because the toast outlives this page. If the tab is closed or
// reloaded first, the pagehide listener sends the deletes then. pagehide also fires when
// the browser keeps the page to come back to, so the sent deletes are dropped here: an
// Undo after coming back finds its delete gone and says it is too late.
const pendingRemovals = new Map<string, { id: number; target: DeleteScope }>();

window.addEventListener("pagehide", () => {
  pendingRemovals.forEach(({ id, target }) => deleteScheduledItemOnExit(id, target));
  pendingRemovals.clear();
});

const removalKey = (id: number, target: DeleteScope) => `${id}|${target.scope === "series" ? "series" : target.date}`;

const isHidden = (occurrence: Occurrence) =>
  pendingRemovals.has(removalKey(occurrence.scheduled_item_id, { scope: "series" })) ||
  pendingRemovals.has(removalKey(occurrence.scheduled_item_id, { scope: "occurrence", date: occurrence.date }));

// The Undo toast closed without Undo: delete for real. A failure brings it back, with a toast saying so.
async function commitRemoval(
  key: string,
  name: string,
  showToast: (text: string) => void,
  reload: () => void
) {
  const removal = pendingRemovals.get(key);
  if (!removal) return;
  try {
    await deleteScheduledItem(removal.id, removal.target);
  } catch (err) {
    // Already gone (deleted somewhere else) is the outcome the caregiver wanted
    const alreadyGone = err instanceof ApiError && (err.status === 404 || err.code === "NOT_AN_OCCURRENCE");
    if (!alreadyGone) showToast(`Couldn't remove "${name}". ${friendlyError(err)}`);
  } finally {
    pendingRemovals.delete(key);
    // Show the server's answer: it is back if the delete failed
    reload();
  }
}

type ViewChoice = "day" | "week" | "month";

const VIEW_NAMES: Record<ViewChoice, CalendarViewName> = {
  day: "timeGridDay",
  week: "timeGridWeek",
  month: "dayGridMonth",
};

const VIEW_SEGMENTS: { value: ViewChoice; label: string }[] = [
  { value: "day", label: "Day" },
  { value: "week", label: "Week" },
  { value: "month", label: "Month" },
];

const choiceOf = (view: CalendarViewName): ViewChoice =>
  view === "timeGridDay" ? "day" : view === "dayGridMonth" ? "month" : "week";

// The form on screen: a new item or an existing one, and what it starts with. `key` restarts the form when another is opened
interface OpenForm {
  key: number;
  initial: FormValues;
  editing: EditTarget | null;
}

export default function Calendar() {
  // RequireCarePartner only renders this page once a carepartner is chosen
  const { activeCarePartnerId } = useAuth();
  const carepartnerId = activeCarePartnerId as number;

  const toast = useToast();
  const calendarRef = useRef<FullCalendar>(null);
  // The "asleep" toast on screen, so clicking around doesn't stack them
  const asleepToastId = useRef<number | null>(null);
  // Which dates the calendar is showing; null until it has drawn the first time
  const [range, setRange] = useState<VisibleRange | null>(null);
  const [occurrences, setOccurrences] = useState<Occurrence[]>([]);
  const [settings, setSettings] = useState<ScreenSettings | null>(null);
  const [loading, setLoading] = useState(true);
  // Kept apart so loading the dates can't clear a failure to load the awake hours
  const [settingsError, setSettingsError] = useState<string | null>(null);
  const [calendarError, setCalendarError] = useState<string | null>(null);
  const [reloadKey, setReloadKey] = useState(0);
  // Bumped to ask the server for the dates showing again: after an item is saved or deleted,
  // and after new awake hours change which items are outside them
  const [calendarKey, setCalendarKey] = useState(0);
  const [editingAwake, setEditingAwake] = useState(false);
  // Photos, videos and audio the form picks from; null until loaded
  const [library, setLibrary] = useState<Media[] | null>(null);
  const [libraryKey, setLibraryKey] = useState(0);
  const [form, setForm] = useState<OpenForm | null>(null);
  // Where a new item would land, drawn on the calendar as "New"
  const [draft, setDraft] = useState<DraftSlot | null>(null);
  const formCount = useRef(0);
  // Counts clicks that open a form, so an item that loads after a later click doesn't open over it
  const openRequest = useRef(0);
  // Changes when a delete is hidden or undone, so the calendar draws again
  const [removalsVersion, setRemovalsVersion] = useState(0);
  // The carepartner whose occurrences are on screen
  const loadedFor = useRef<number | null>(null);

  const awakeHours = useMemo(() => (settings ? awakeHoursOf(settings) : null), [settings]);

  const handleRangeChange = (next: VisibleRange) => {
    // The calendar reports again after every redraw; only new dates need new data
    setRange((current) =>
      current && current.from === next.from && current.to === next.to && current.view === next.view && current.title === next.title && current.current === next.current
        ? current
        : next
    );
  };

  useEffect(() => {
    let cancelled = false;
    setLibrary(null);
    listMedia(carepartnerId)
      .then((list) => {
        if (!cancelled) setLibrary(list);
      })
      .catch((err) => {
        if (cancelled) return;
        toast.show(`Couldn't load the photos, videos and audio to choose from. ${friendlyError(err)}`, {
          actionLabel: "Try again",
          onAction: () => setLibraryKey((key) => key + 1),
        });
      });
    return () => {
      cancelled = true;
    };
  }, [carepartnerId, libraryKey, toast]);

  useEffect(() => {
    let cancelled = false;
    setSettings(null);
    getScreenSettings(carepartnerId)
      .then((loaded) => {
        if (cancelled) return;
        setSettings(loaded);
        setSettingsError(null);
      })
      .catch((err) => {
        if (!cancelled) setSettingsError(friendlyError(err));
      });
    return () => {
      cancelled = true;
    };
  }, [carepartnerId, reloadKey]);

  // Reopened from the browser's back/forward cache: pagehide already sent the pending deletes, so ask the server again
  useEffect(() => {
    const onShow = (event: PageTransitionEvent) => {
      if (!event.persisted) return;
      setRemovalsVersion((value) => value + 1);
      setCalendarKey((key) => key + 1);
    };
    window.addEventListener("pageshow", onShow);
    return () => window.removeEventListener("pageshow", onShow);
  }, []);

  const openNew = (date: string, minute: number, media?: { kind: "video" | "audio"; id: number }) => {
    openRequest.current += 1;
    formCount.current += 1;
    setEditingAwake(false);
    setForm({ key: formCount.current, initial: valuesForNew(date, minute, media), editing: null });
  };

  const closeForm = () => {
    openRequest.current += 1;
    setForm(null);
    setDraft(null);
  };

  // The Schedule button: the day on screen (today if it's showing), at the next time that's still to come
  const openFromButton = () => {
    if (!awakeHours || !range) return;
    const now = new Date();
    const today = dateText(now);
    const date = range.view === "timeGridDay" ? range.from : today >= range.from && today <= range.to ? today : range.current;
    const slot = firstOpenSlot(awakeHours, 0, date, today, minuteOfDay(now));
    openNew(slot.date, slot.minute);
  };

  // The media page's "Schedule it" arrives as /calendar?media=<id>: open the form with that file chosen,
  // at the next time it fits (tomorrow once today is too late). The address is cleared once the
  // library and awake hours have loaded, so reloading doesn't open it again.
  const [searchParams, setSearchParams] = useSearchParams();
  const wantedMedia = searchParams.get("media");
  useEffect(() => {
    if (!wantedMedia || !library || !awakeHours) return;
    setSearchParams({}, { replace: true });
    const media = library.find((item) => String(item.id) === wantedMedia);
    if (!media) {
      toast.show("That file isn't in this carepartner's library any more.");
      return;
    }
    if (media.kind !== "video" && media.kind !== "audio") {
      toast.show("Only videos and audio can be scheduled here.");
      return;
    }
    const now = new Date();
    const today = dateText(now);
    const slot = firstOpenSlot(awakeHours, media.duration_seconds ?? 0, today, today, minuteOfDay(now));
    openNew(slot.date, slot.minute, { kind: media.kind, id: media.id });
    // openNew and toast only matter when the address asks for a file
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wantedMedia, library, awakeHours, setSearchParams]);

  // Clicking an item opens it to edit; an occurrence doesn't carry the item's start and end dates, so ask for the item
  const openEdit = async ({ itemId, date, shape }: ClickedItem) => {
    // The form needs the awake hours; the error above says when they couldn't be loaded
    if (!awakeHours) return;
    if (shape === "photo") {
      toast.show("A photo on the schedule can't be changed here.");
      return;
    }
    openRequest.current += 1;
    const request = openRequest.current;
    try {
      const item = await getScheduledItem(itemId);
      // Another item or time was clicked while this one loaded
      if (request !== openRequest.current) return;
      const initial = valuesForEdit(item, date);
      if (!initial) {
        toast.show("That can't be changed here.");
        return;
      }
      formCount.current += 1;
      setEditingAwake(false);
      setForm({ key: formCount.current, initial, editing: { item, occurrenceDate: date } });
    } catch (err) {
      if (request === openRequest.current) toast.show(friendlyError(err));
    }
  };

  const handleSaved = (saved: ScheduledItem, wasEdit: boolean) => {
    closeForm();
    setCalendarKey((key) => key + 1);
    toast.show(`"${saved.media.display_name}" ${wasEdit ? "updated" : "scheduled"}. ${describeScheduledItem(saved)}.`);
  };

  // Hide it right away; the server isn't told until the Undo toast closes, so Undo only has to show it again
  const handleRemove = ({ item, target }: Removal) => {
    closeForm();
    const key = removalKey(item.id, target);
    const name = item.media.display_name;
    const reload = () => setCalendarKey((value) => value + 1);

    pendingRemovals.set(key, { id: item.id, target });
    setRemovalsVersion((value) => value + 1);

    let text = `"${name}" deleted.`;
    if (target.scope === "occurrence") {
      const day = parseDate(target.date);
      text = `"${name}" removed from ${dayName(day)}, ${dateLabel(day)} only.`;
    }

    toast.show(text, {
      actionLabel: "Undo",
      onAction: () => {
        // pagehide sent it already (the page was left and come back to), so it can't be undone
        if (!pendingRemovals.delete(key)) {
          toast.show(`"${name}" was already removed, so it can't be undone.`);
          return;
        }
        setRemovalsVersion((value) => value + 1);
      },
      onDismiss: () => void commitRemoval(key, name, toast.show, reload),
    });
  };

  const handleDateClick = ({ date, view }: DateClickArg) => {
    const calendar = calendarRef.current?.getApi();
    // A day in month view opens that day
    if (view.type === "dayGridMonth") {
      calendar?.changeView("timeGridDay", date);
      return;
    }
    if (!awakeHours) return;
    const minute = minuteOfDay(date);
    if (isAwakeAt(minute, awakeHours)) {
      openNew(dateText(date), minute);
      return;
    }
    if (asleepToastId.current !== null) toast.dismiss(asleepToastId.current);
    asleepToastId.current = toast.show(
      `The carepartner is asleep at ${minuteLabel(minute)}. Videos and audio can only play between ${minuteLabel(awakeHours.wakeMinute)} and ${minuteLabel(awakeHours.bedMinute)}.`,
      {
        actionLabel: "Change awake hours",
        onAction: () => {
          closeForm();
          setEditingAwake(true);
        },
      }
    );
  };

  const handleAwakeSaved = (saved: ScreenSettings) => {
    setSettings(saved);
    setEditingAwake(false);
    setCalendarKey((key) => key + 1);
    toast.show("Awake hours saved.");
  };

  const from = range?.from;
  const to = range?.to;
  useEffect(() => {
    if (from === undefined || to === undefined) return;

    // Another carepartner's schedule must not show while this one loads
    if (loadedFor.current !== carepartnerId) setOccurrences([]);

    const controller = new AbortController();
    setLoading(true);
    getCalendar(carepartnerId, from, to, controller.signal)
      .then((list) => {
        loadedFor.current = carepartnerId;
        setOccurrences(list);
        setCalendarError(null);
        setLoading(false);
      })
      .catch((err) => {
        // Moving to other dates cancels this request; the next one reports for itself
        if (isAbort(err)) return;
        setCalendarError(friendlyError(err));
        setLoading(false);
      });
    return () => controller.abort();
  }, [carepartnerId, from, to, reloadKey, calendarKey]);

  // pendingRemovals isn't state, so removalsVersion says when a delete was hidden or undone
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const visibleOccurrences = useMemo(() => occurrences.filter((occurrence) => !isHidden(occurrence)), [occurrences, removalsVersion]);

  const choice = range ? choiceOf(range.view) : "week";
  // The awake hours first: without them nothing can be scheduled
  const loadError = settingsError ?? calendarError;

  return (
    <div className="lg:h-screen flex flex-col gap-3 px-7 pt-5 pb-5">
      <div className="flex flex-wrap items-center gap-x-2.5 gap-y-2 min-h-11">
        <h1 className="m-0 w-[150px] text-3xl font-extrabold">Calendar</h1>
        <IconButton
          label={`Previous ${choice}`}
          icon={<ChevronLeft size={20} strokeWidth={2.4} aria-hidden="true" />}
          onClick={() => calendarRef.current?.getApi().prev()}
        />
        <IconButton
          label={`Next ${choice}`}
          icon={<ChevronRight size={20} strokeWidth={2.4} aria-hidden="true" />}
          onClick={() => calendarRef.current?.getApi().next()}
        />
        <h2 className="m-0 text-xl font-bold" aria-live="polite">
          {range?.title}
        </h2>
        {loading && (
          <span role="status" className="text-sm text-ink-muted">
            Loading…
          </span>
        )}
        <div className="grow" />
        <InfoPopover label="About the calendar">{PAGE_INFO}</InfoPopover>
        <div className="w-[270px]">
          <SegmentedControl
            label="View"
            segments={VIEW_SEGMENTS}
            value={choice}
            onChange={(next) => calendarRef.current?.getApi().changeView(VIEW_NAMES[next])}
          />
        </div>
      </div>

      {loadError && (
        <p role="alert" className="m-0 flex items-center gap-3 text-base text-link">
          {loadError}
          <button
            type="button"
            onClick={() => setReloadKey((key) => key + 1)}
            className={`h-10 px-2 bg-transparent font-bold text-link underline cursor-pointer ${FOCUS_RING}`}
          >
            Try again
          </button>
        </p>
      )}

      <div className="flex flex-col lg:flex-row gap-4 grow min-h-0">
        {form && awakeHours ? (
          <ScheduleForm
            key={form.key}
            carepartnerId={carepartnerId}
            hours={awakeHours}
            library={library}
            initial={form.initial}
            editing={form.editing}
            onDraftChange={setDraft}
            onClose={closeForm}
            onSaved={handleSaved}
            onRemove={handleRemove}
            isHidden={isHidden}
          />
        ) : (
          <CalendarSidebar
            carepartnerId={carepartnerId}
            range={range}
            settings={settings}
            editingAwake={editingAwake}
            onEditAwake={() => setEditingAwake(true)}
            onCloseAwake={() => setEditingAwake(false)}
            onAwakeSaved={handleAwakeSaved}
            onGoToDate={(date) => calendarRef.current?.getApi().gotoDate(date)}
            onSchedule={awakeHours ? openFromButton : undefined}
          />
        )}
        <div className="grow min-w-0 min-h-[600px] lg:min-h-0">
          <ScheduleCalendar
            calendarRef={calendarRef}
            occurrences={visibleOccurrences}
            awakeHours={awakeHours}
            onRangeChange={handleRangeChange}
            onDateClick={handleDateClick}
            onItemClick={openEdit}
            draft={draft}
            selected={form?.editing ? { itemId: form.editing.item.id, date: form.editing.occurrenceDate } : null}
          />
        </div>
      </div>
    </div>
  );
}
