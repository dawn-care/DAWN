/**
 * ScreenPreview.tsx
 *
 * User: Caregiver
 *
 * The carepartner's screen on the home page, at the tablet's 16:10 shape,
 * with an info button and a Full screen button (to /preview) below its right
 * corner. While editing, `editView` (the placeholder grid) takes the screen's
 * place and Full screen is hidden, since it would show the saved screen.
 *
 * The screen is the carepartner web app itself (CarepartnerScreen), so it
 * looks exactly like the tablet. When the carepartner app's address isn't set
 * (e.g. a Docker build), ScreenUnavailable says so instead.
 */

import { Maximize } from "lucide-react";
import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import InfoPopover from "../components/ui/InfoPopover";
import { FOCUS_RING } from "../components/ui/styles";
import CarepartnerScreen from "./CarepartnerScreen";
import { previewUrl } from "./previewUrl";
import ScreenUnavailable from "./ScreenUnavailable";

const SCREEN_INFO =
  "This is what the carepartner's idle screen looks like. Click Edit screen to change it, or Full screen to enlarge it.";
const EDIT_INFO =
  "Each square shows where something sits on the carepartner's screen. Pick something to add above the screen, then click a lit square to place it. Nothing changes on the tablet until you save. Click Cancel to undo all of your changes.";

interface ScreenPreviewProps {
  carepartnerId: number;
  /** Change it (e.g. after a save) to load the iframe again */
  reloadKey?: number;
  /** Shown in the screen's place while editing */
  editView?: ReactNode;
}

export default function ScreenPreview({ carepartnerId, reloadKey = 0, editView }: ScreenPreviewProps) {
  const src = previewUrl(carepartnerId);
  const editing = editView !== undefined;

  return (
    // Fills the page down to the bottom, left aligned. Its container measures the space left, and the screen
    // takes the largest 16:10 box that fits while leaving 3.5rem below it: the 0.75rem gap and the 2.75rem
    // row of buttons
    <div className="flex-1 min-w-0 min-h-[360px] pt-2 [container-type:size]">
      <div className="flex flex-col gap-3" style={{ width: "min(100cqw, calc((100cqh - 3.5rem) * 1.6))" }}>
        {/* The border shows where the carepartner's screen ends; ink-muted is the border of the tablet's widgets (colors.day.border in carepartner/src/theme/theme.ts) */}
        <div className="aspect-[16/10] rounded-xl overflow-hidden bg-page border-[3px] border-ink-muted">
          {editing ? (
            <div className="h-full p-4">{editView}</div>
          ) : src ? (
            <CarepartnerScreen src={src} reloadKey={reloadKey} fit="cover" />
          ) : (
            <ScreenUnavailable />
          )}
        </div>
        <div className="flex items-center justify-end gap-1 h-11">
          <InfoPopover label="About the carepartner's screen">{editing ? EDIT_INFO : SCREEN_INFO}</InfoPopover>
          {!editing && (
            <Link
              to="/preview"
              className={`flex items-center gap-2 h-11 px-[18px] rounded-[10px] border-2 border-accent bg-accent hover:brightness-95 text-on-accent text-base font-bold ${FOCUS_RING}`}
            >
              <Maximize size={18} strokeWidth={2.2} aria-hidden="true" />
              Full screen
            </Link>
          )}
        </div>
      </div>
    </div>
  );
}
