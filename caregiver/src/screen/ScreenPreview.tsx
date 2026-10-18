/**
 * ScreenPreview.tsx
 *
 * User: Caregiver
 *
 * The carepartner's saved screen on the home page, at the tablet's 16:10
 * shape, with an info button and a Full screen button (to /preview) below its
 * right corner.
 *
 * The screen is the carepartner web app itself (CarepartnerScreen), so it
 * looks exactly like the tablet. When the carepartner app's address isn't set
 * (e.g. a Docker build), ScreenUnavailable says so instead.
 */

import { Maximize } from "lucide-react";
import { Link } from "react-router-dom";
import InfoPopover from "../components/ui/InfoPopover";
import { FOCUS_RING } from "../components/ui/styles";
import CarepartnerScreen from "./CarepartnerScreen";
import { previewUrl } from "./previewUrl";
import ScreenUnavailable from "./ScreenUnavailable";

const SCREEN_INFO =
  "This is what the carepartner's idle screen looks like. Click Edit Screen to change it, or Full Screen to enlarge it.";

interface ScreenPreviewProps {
  carepartnerId: number;
  /** Change it (e.g. after a save) to load the iframe again */
  reloadKey?: number;
}

export default function ScreenPreview({ carepartnerId, reloadKey = 0 }: ScreenPreviewProps) {
  const src = previewUrl(carepartnerId);

  return (
    <div className="w-full max-w-[880px] mx-auto flex flex-col gap-1">
      <div className="aspect-[16/10] rounded-xl overflow-hidden bg-dawn-preview">
        {src ? <CarepartnerScreen src={src} reloadKey={reloadKey} /> : <ScreenUnavailable />}
      </div>
      <div className="flex items-center justify-end gap-1">
        <InfoPopover label="About her screen">{SCREEN_INFO}</InfoPopover>
        <Link
          to="/preview"
          className={`flex items-center gap-1.5 h-8 px-3 rounded-lg bg-dawn-accent hover:bg-dawn-accent-hover text-white text-sm font-bold ${FOCUS_RING}`}
        >
          <Maximize size={16} strokeWidth={2.2} aria-hidden="true" />
          Full screen
        </Link>
      </div>
    </div>
  );
}
