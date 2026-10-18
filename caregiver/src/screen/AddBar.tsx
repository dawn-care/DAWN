/**
 * AddBar.tsx
 *
 * User: Caregiver
 *
 * The "Add to her screen" bar above the screen: one button per kind of widget
 * the caregiver can add, and a line of hint text. The buttons are greyed out
 * until Edit screen is pressed. Picking one lights the squares on the screen
 * where it fits (ScreenGrid); picking it again, or pressing Escape, stops
 * placing. Picking one that fits nowhere places nothing and says the screen is
 * full instead.
 */

import { widgetName } from "@shared";
import { useEffect } from "react";
import InfoPopover from "../components/ui/InfoPopover";
import { FOCUS_RING } from "../components/ui/styles";
import { ADDABLE_WIDGET_TYPES } from "./useScreenEditor";
import type { AddableWidgetType } from "./useScreenEditor";
import WidgetIcon from "./WidgetIcon";

const ADD_INFO =
  "After selecting Edit screen, click an icon, then on one of the lit squares to place it there. Only squares where the icon will fit light up. Videos and audio are scheduled on the Calendar.";

interface AddBarProps {
  editing: boolean;
  /** The type being placed, or null */
  pendingType: AddableWidgetType | null;
  /** The type the caregiver picked that fits nowhere, or null */
  fullType: AddableWidgetType | null;
  onPick: (type: AddableWidgetType | null) => void;
}

const lower = (type: AddableWidgetType) => widgetName({ type }).toLowerCase();

function hintFor({
  editing,
  pendingType,
  fullType,
}: Pick<AddBarProps, "editing" | "pendingType" | "fullType">) {
  if (!editing) return "Press Edit screen to add or change icons";
  if (fullType)
    return `No room for a ${lower(fullType)}. Remove something on the screen first.`;
  if (pendingType)
    return `Click a lit square to put the ${lower(pendingType)} there.`;
  return "Select an icon, then click a lit square.";
}

export default function AddBar({
  editing,
  pendingType,
  fullType,
  onPick,
}: AddBarProps) {
  // Escape puts the picked type back down
  useEffect(() => {
    if (!pendingType) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") onPick(null);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [pendingType, onPick]);

  return (
    <section
      aria-label="Add to the screen"
      className="flex flex-wrap items-center gap-2 min-h-[76px] px-3.5 py-2 bg-white border-[1.5px] border-dawn-line rounded-[14px]"
    >
      <div className="w-[250px] shrink-0 flex flex-col gap-0.5">
        <div className="flex items-center gap-0.5 h-7">
          <span className="text-[17px] font-extrabold">Add to the screen</span>
          <InfoPopover label="About adding things">{ADD_INFO}</InfoPopover>
        </div>
        <div
          className={`text-sm leading-snug ${fullType ? "text-dawn-accent" : "text-dawn-muted"}`}
        >
          {hintFor({ editing, pendingType, fullType })}
        </div>
      </div>
      {ADDABLE_WIDGET_TYPES.map((type) => {
        const picked = pendingType === type;
        return (
          <button
            key={type}
            type="button"
            disabled={!editing}
            aria-pressed={picked}
            onClick={() => onPick(type)}
            className={`w-[100px] h-14 flex flex-col items-center justify-center gap-[3px] p-0 rounded-[10px] text-sm font-bold ${FOCUS_RING} ${
              !editing
                ? "bg-dawn-sidebar border-2 border-dawn-line-strong text-dawn-control cursor-default"
                : picked
                  ? "bg-[#FFF1E6] border-[2.5px] border-dawn-accent text-dawn-ink cursor-pointer"
                  : "bg-white border-2 border-dawn-control hover:bg-dawn-page text-dawn-ink cursor-pointer"
            }`}
          >
            <WidgetIcon type={type} size={20} />
            {widgetName({ type })}
          </button>
        );
      })}
    </section>
  );
}
