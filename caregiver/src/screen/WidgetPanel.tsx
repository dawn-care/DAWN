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
 * every change.
 */

import { TEXT_LIMITS, widgetName } from "@shared";
import { useRef } from "react";
import TextField from "../components/ui/form/TextField";
import { FOCUS_RING } from "../components/ui/styles";
import type { MessageConfig, WeatherConfig } from "./api";
import type { EditorWidget } from "./layout";
import { isValidZip } from "./widgetSettings";
import type { WidgetChanges } from "./useScreenEditor";

interface WidgetPanelProps {
  /** The widget whose settings to show, or null for an empty panel */
  widget: EditorWidget | null;
  /** Change the widget's settings in the draft */
  onChange: (changes: WidgetChanges) => void;
  /** Save: keep the widget and what was typed; `original` is the widget as it was when the panel opened */
  onDone: (original: EditorWidget) => void;
}

export default function WidgetPanel({ widget, onChange, onDone }: WidgetPanelProps) {
  return (
    <aside
      aria-label="Settings"
      className="h-full flex flex-col gap-3 p-4 bg-white border-[1.5px] border-dawn-line rounded-[14px] text-dawn-ink overflow-y-auto"
    >
      {/* A new widget gets a new set of settings, remembering how it was when they opened */}
      {widget && <WidgetSettings key={widget.id} widget={widget} onChange={onChange} onDone={onDone} />}
    </aside>
  );
}

function WidgetSettings({ widget, onChange, onDone }: Omit<WidgetPanelProps, "widget"> & { widget: EditorWidget }) {
  const name = widgetName({ type: widget.type });
  // The widget as it was when the panel opened; the parent mounts a new panel for each widget
  const original = useRef(widget);

  return (
    <>
      <div className="flex flex-col gap-0.5">
        <span className="text-[13px] font-bold uppercase tracking-[0.06em] text-dawn-muted">{name}</span>
        <span className="text-xl font-extrabold">{name}</span>
      </div>

      {widget.type === "message" && <MessageSettings widget={widget} onChange={onChange} />}
      {widget.type === "weather" && <WeatherSettings widget={widget} onChange={onChange} />}

      <div className="flex justify-end mt-auto">
        <button
          type="button"
          onClick={() => onDone(original.current)}
          className={`h-11 px-[18px] rounded-[10px] border-2 border-dawn-accent bg-dawn-accent hover:bg-dawn-accent-hover hover:border-dawn-accent-hover text-white text-base font-bold cursor-pointer ${FOCUS_RING}`}
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
      <div className="self-end text-sm text-dawn-muted" aria-live="polite">
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
        <div id="weather-zip-help" className="text-[15px] leading-snug text-dawn-accent">
          Enter a 5-digit zip code.
        </div>
      )}
    </div>
  );
}
