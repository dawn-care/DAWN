/**
 * WidgetPanel.tsx
 *
 * User: Caregiver
 *
 * The panel to the right of the screen, where a widget's settings are changed.
 * It is always on the page; with no widget to change (not editing, nothing
 * selected, or a widget with no settings) it is an empty card. What the
 * caregiver types changes the draft straight away, so the placeholder tile
 * shows it. Save closes the settings and keeps the widget. There is no close
 * button: a widget is removed only with its bin, and Cancel on the page undoes
 * every change. While `disabled` (the draft is being saved) every field and the
 * Save button are locked.
 */

import { TEXT_LIMITS, widgetName } from "@shared";
import { useCallback, useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import PhotoPicker from "../components/ui/PhotoPicker";
import SelectField from "../components/ui/form/SelectField";
import TextField from "../components/ui/form/TextField";
import { FOCUS_RING } from "../components/ui/styles";
import { fileUrl, friendlyError as mediaError, listMedia } from "../media/api";
import type { Media } from "../media/api";
import { ApiError, friendlyError, loadSportsLeagues } from "./api";
import type { MessageConfig, SportsConfig, SportsLeague, WeatherConfig } from "./api";
import type { EditorWidget } from "./layout";
import { isComplete, isValidZip } from "./widgetSettings";
import type { WidgetChanges } from "./useScreenEditor";

interface WidgetPanelProps {
  /** Whose photos the photo picker lists */
  carepartnerId: number;
  /** The widget whose settings to show, or null for an empty panel */
  widget: EditorWidget | null;
  /** Locks every field and Save, e.g. while the draft is being saved */
  disabled?: boolean;
  /** Change the widget's settings in the draft */
  onChange: (changes: WidgetChanges) => void;
  /** Save: keep the widget and what was typed; `original` is the widget as it was when the panel opened */
  onDone: (original: EditorWidget) => void;
}

export default function WidgetPanel({ carepartnerId, widget, disabled = false, onChange, onDone }: WidgetPanelProps) {
  return (
    <aside
      aria-label="Settings"
      className="h-full flex flex-col gap-3 p-4 bg-surface border-[1.5px] border-border rounded-[14px] text-ink overflow-y-auto"
    >
      {/* Each selected widget gets its own settings, which remember how it was when they opened */}
      {widget && (
        // A disabled fieldset locks every input, select and button inside it
        <fieldset disabled={disabled} className="contents">
          <WidgetSettings
            key={widget.id}
            carepartnerId={carepartnerId}
            widget={widget}
            onChange={onChange}
            onDone={onDone}
          />
        </fieldset>
      )}
    </aside>
  );
}

function WidgetSettings({
  carepartnerId,
  widget,
  onChange,
  onDone,
}: Omit<WidgetPanelProps, "widget" | "disabled"> & { widget: EditorWidget }) {
  const name = widgetName({ type: widget.type });
  // The widget as it was when the panel opened; the parent mounts a new panel for each widget
  const original = useRef(widget);

  return (
    <>
      <div className="flex flex-col gap-0.5">
        <span className="text-[13px] font-bold uppercase tracking-[0.06em] text-ink-muted">{name}</span>
        <span className="text-xl font-extrabold">{name}</span>
      </div>

      {widget.type === "photo" && <PhotoSettings carepartnerId={carepartnerId} widget={widget} onChange={onChange} />}
      {widget.type === "message" && <MessageSettings widget={widget} onChange={onChange} />}
      {widget.type === "weather" && <WeatherSettings widget={widget} onChange={onChange} />}
      {widget.type === "sports" && <SportsSettings widget={widget} onChange={onChange} />}

      <div className="flex justify-end mt-auto">
        <button
          type="button"
          onClick={() => onDone(original.current)}
          // A widget can't be kept without what it needs: text, a zip, a photo, a league and team
          disabled={!isComplete(widget)}
          className={`h-11 px-[18px] rounded-[10px] border-2 border-accent bg-accent hover:brightness-95 text-on-accent text-base font-bold cursor-pointer disabled:opacity-50 disabled:cursor-default disabled:hover:bg-accent disabled:hover:border-accent ${FOCUS_RING}`}
        >
          Save
        </button>
      </div>
    </>
  );
}

interface SettingsProps {
  widget: EditorWidget;
  onChange: (changes: WidgetChanges) => void;
}

function MessageSettings({ widget, onChange }: SettingsProps) {
  const config = widget.config as MessageConfig;
  const text = config.text ?? "";
  return (
    <div className="flex flex-col gap-1">
      <TextField
        label="Text"
        value={text}
        maxLength={TEXT_LIMITS.message}
        autoFocus
        onChange={(event) => onChange({ config: { ...config, text: event.target.value } })}
      />
      <div className="self-end text-sm text-ink-muted" aria-live="polite">
        {text.length} / {TEXT_LIMITS.message}
      </div>
    </div>
  );
}

function WeatherSettings({ widget, onChange }: SettingsProps) {
  const config = widget.config as WeatherConfig;
  const zip = config.zip ?? "";
  const invalid = !isValidZip(zip);
  return (
    <div className="flex flex-col gap-2">
      <TextField
        label="Zip code"
        value={zip}
        inputMode="numeric"
        maxLength={5}
        autoFocus
        aria-invalid={invalid}
        aria-describedby={invalid ? "weather-zip-help" : undefined}
        // Digits only: anything else typed or pasted is dropped
        onChange={(event) => onChange({ config: { ...config, zip: event.target.value.replace(/\D/g, "") } })}
      />
      {invalid && (
        <div id="weather-zip-help" className="text-[15px] leading-snug text-link">
          Enter a 5-digit zip code.
        </div>
      )}
    </div>
  );
}

/** Loads something when the settings open, with a way to try again */
function useLoad<T>(load: () => Promise<T>, describe: (err: unknown) => string) {
  const [state, setState] = useState<
    { status: "loading" } | { status: "error"; message: string } | { status: "ready"; value: T }
  >({ status: "loading" });
  const [attempt, setAttempt] = useState(0);
  // The loader and its wording are the same on every render; only a retry loads again
  const latest = useRef({ load, describe });
  latest.current = { load, describe };

  useEffect(() => {
    let current = true;
    setState({ status: "loading" });
    latest.current.load().then(
      (value) => current && setState({ status: "ready", value }),
      (err: unknown) => {
        if (!current) return;
        console.error("Loading the widget's settings failed", err);
        setState({ status: "error", message: latest.current.describe(err) });
      },
    );
    return () => {
      current = false;
    };
  }, [attempt]);

  return { state, retry: useCallback(() => setAttempt((n) => n + 1), []) };
}

function LoadError({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div role="alert" className="flex flex-col items-start gap-2 text-base text-ink-muted">
      <p className="m-0">{message}</p>
      <button
        type="button"
        onClick={onRetry}
        className={`h-11 px-[18px] rounded-[10px] border-2 border-ink-muted bg-surface hover:bg-page text-ink text-base font-bold cursor-pointer ${FOCUS_RING}`}
      >
        Try again
      </button>
    </div>
  );
}

function PhotoSettings({ carepartnerId, widget, onChange }: SettingsProps & { carepartnerId: number }) {
  const { state, retry } = useLoad<Media[]>(() => listMedia(carepartnerId), mediaError);

  if (state.status === "loading") return <p className="m-0 text-base text-ink-muted">Loading photos…</p>;
  if (state.status === "error") return <LoadError message={state.message} onRetry={retry} />;

  const photos = state.value.filter((media) => media.kind === "photo");
  const chosen = photos.find((media) => media.id === widget.photoMediaId);
  return (
    <div className="flex flex-col gap-2">
      <span className="text-base font-bold">Photo</span>
      {photos.length === 0 ? (
        <div className="flex items-center justify-between gap-2">
          <p className="m-0 text-base text-ink-muted">No photos yet.</p>
          <Link to="/media" className="shrink-0 text-base font-bold text-link hover:underline">
            Add photos
          </Link>
        </div>
      ) : (
        <PhotoPicker
          label="Photo"
          photos={photos.map((media) => ({ id: String(media.id), label: media.display_name, src: fileUrl(media) }))}
          value={widget.photoMediaId === null ? "" : String(widget.photoMediaId)}
          onChange={(id) => {
            const picked = photos.find((media) => String(media.id) === id);
            if (picked) onChange({ photoMediaId: picked.id, photoUrl: picked.url });
          }}
        />
      )}
      {/* The chosen photo, larger, with its name under it */}
      {chosen && (
        <figure className="m-0 flex flex-col gap-1.5">
          <img src={fileUrl(chosen)} alt="" className="w-full aspect-[4/3] object-cover rounded-lg bg-border" />
          <figcaption className="text-base font-bold break-words">{chosen.display_name}</figcaption>
        </figure>
      )}
    </div>
  );
}

// The save wording ("Two widgets overlap"...) doesn't fit a failed load, so only a lost connection keeps its own
function teamsLoadError(err: unknown): string {
  if (err instanceof ApiError && err.code === "NETWORK_ERROR") return friendlyError(err);
  return "The teams couldn't be loaded. Please try again.";
}

function SportsSettings({ widget, onChange }: SettingsProps) {
  const config = widget.config as Partial<SportsConfig>;
  const { state, retry } = useLoad<SportsLeague[]>(loadSportsLeagues, teamsLoadError);

  if (state.status === "loading") return <p className="m-0 text-base text-ink-muted">Loading teams…</p>;
  if (state.status === "error") return <LoadError message={state.message} onRetry={retry} />;

  const league = state.value.find((other) => other.id === config.league);
  return (
    <div className="flex flex-col gap-3">
      <SelectField
        label="League"
        placeholder="Choose a league"
        value={config.league ?? ""}
        options={state.value.map((other) => ({ value: other.id, label: other.name }))}
        // A team belongs to one league, so changing the league clears it
        onChange={(event) => onChange({ config: { league: event.target.value as SportsConfig["league"], team: "" } })}
      />
      <SelectField
        label="Team"
        placeholder="Choose a team"
        value={config.team ?? ""}
        disabled={!league}
        options={(league?.teams ?? []).map((team) => ({ value: team.abbr, label: team.name }))}
        onChange={(event) => league && onChange({ config: { league: league.id, team: event.target.value } })}
      />
    </div>
  );
}
