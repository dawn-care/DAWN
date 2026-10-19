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
 * icon and the view switch) and loads what the calendar draws: the awake
 * hours once, and the scheduled occurrences for whichever dates are showing.
 */

import type { DateClickArg } from "@fullcalendar/interaction";
import type FullCalendar from "@fullcalendar/react";
import { isAwakeAt, minuteOfDay } from "@shared";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useToast } from "../components/ui/toastContext";
import IconButton from "../components/ui/IconButton";
import InfoPopover from "../components/ui/InfoPopover";
import SegmentedControl from "../components/ui/form/SegmentedControl";
import { FOCUS_RING } from "../components/ui/styles";
import { useAuth } from "../context/AuthContext";
import { friendlyError, getCalendar, getScheduledItem, getScreenSettings, isAbort } from "../calendar/api";
import type { Occurrence, ScheduledItem, ScreenSettings } from "../calendar/api";
import CalendarSidebar from "../calendar/CalendarSidebar";
import ScheduleCalendar from "../calendar/ScheduleCalendar";
import type { CalendarViewName, ClickedItem, VisibleRange } from "../calendar/ScheduleCalendar";
import ScheduleForm from "../calendar/ScheduleForm";
import type { DraftSlot } from "../calendar/ScheduleForm";
import { nextOpenMinute, valuesForEdit, valuesForNew } from "../calendar/formRules";
import type { FormValues } from "../calendar/formRules";
import { awakeHoursOf, dateText, minuteLabel } from "../calendar/time";
import { listMedia } from "../media/api";
import type { Media } from "../media/api";
import { describeScheduledItem } from "../media/schedule";

const PAGE_INFO =
  "Schedule videos and audio to play on the carepartner's tablet at set times. Shaded hours are when the carepartner is asleep, and nothing plays then.";

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
  editing: { item: ScheduledItem; occurrenceDate: string } | null;
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
  const [loadError, setLoadError] = useState<string | null>(null);
  const [reloadKey, setReloadKey] = useState(0);
  // Changes after new awake hours are saved, so the calendar asks again which items are outside them
  const [calendarKey, setCalendarKey] = useState(0);
  const [editingAwake, setEditingAwake] = useState(false);
  // Photos, videos and audio the form picks from; null until loaded
  const [library, setLibrary] = useState<Media[] | null>(null);
  const [libraryKey, setLibraryKey] = useState(0);
  const [form, setForm] = useState<OpenForm | null>(null);
  // Where a new item would land, drawn on the calendar as "New"
  const [draft, setDraft] = useState<DraftSlot | null>(null);
  const formCount = useRef(0);
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
        if (!cancelled) setSettings(loaded);
      })
      .catch((err) => {
        if (!cancelled) setLoadError(friendlyError(err));
      });
    return () => {
      cancelled = true;
    };
  }, [carepartnerId, reloadKey]);

  const openNew = (date: string, minute: number) => {
    formCount.current += 1;
    setEditingAwake(false);
    setForm({ key: formCount.current, initial: valuesForNew(date, minute), editing: null });
  };

  const closeForm = () => {
    setForm(null);
    setDraft(null);
  };

  // The Schedule button: the day on screen (today if it's showing), at the next time that's open
  const openFromButton = () => {
    if (!awakeHours || !range) return;
    const today = dateText(new Date());
    const date = range.view === "timeGridDay" ? range.from : today >= range.from && today <= range.to ? today : range.current;
    openNew(date, nextOpenMinute(awakeHours, date === today ? minuteOfDay(new Date()) : null));
  };

  // Clicking an item opens it to edit; an occurrence doesn't carry the item's start and end dates, so ask for the item
  const openEdit = async ({ itemId, date, shape }: ClickedItem) => {
    if (shape === "photo") {
      toast.show("A photo on the schedule can't be changed here.");
      return;
    }
    try {
      const item = await getScheduledItem(itemId);
      const initial = valuesForEdit(item, date);
      if (!initial) {
        toast.show("That can't be changed here.");
        return;
      }
      formCount.current += 1;
      setEditingAwake(false);
      setForm({ key: formCount.current, initial, editing: { item, occurrenceDate: date } });
    } catch (err) {
      toast.show(friendlyError(err));
    }
  };

  const handleSaved = (saved: ScheduledItem, wasEdit: boolean) => {
    closeForm();
    setCalendarKey((key) => key + 1);
    toast.show(`"${saved.media.display_name}" ${wasEdit ? "updated" : "scheduled"}. ${describeScheduledItem(saved)}.`);
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
        setLoadError(null);
        setLoading(false);
      })
      .catch((err) => {
        // Moving to other dates cancels this request; the next one reports for itself
        if (isAbort(err)) return;
        setLoadError(friendlyError(err));
        setLoading(false);
      });
    return () => controller.abort();
  }, [carepartnerId, from, to, reloadKey, calendarKey]);

  const choice = range ? choiceOf(range.view) : "week";

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
          <span role="status" className="text-sm text-dawn-muted">
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
        <p role="alert" className="m-0 flex items-center gap-3 text-base text-dawn-accent">
          {loadError}
          <button
            type="button"
            onClick={() => setReloadKey((key) => key + 1)}
            className={`h-10 px-2 bg-transparent font-bold text-dawn-accent underline cursor-pointer ${FOCUS_RING}`}
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
            occurrences={occurrences}
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
