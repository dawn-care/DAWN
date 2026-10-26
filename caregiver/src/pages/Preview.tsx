/**
 * Preview.tsx
 *
 * User: Caregiver
 *
 * Full-screen preview of the carepartner's screen at `/preview`, opened by the
 * home page's Full screen button. Renders outside AppShell, so there is no nav
 * bar: the carepartner's screen (CarepartnerScreen) fills the window, with an
 * Exit full screen button over its top right corner.
 */

import { Link } from "react-router-dom";
import { Minimize } from "lucide-react";
import { FOCUS_RING } from "../components/ui/styles";
import { useAuth } from "../context/AuthContext";
import CarepartnerScreen from "../screen/CarepartnerScreen";
import { previewUrl } from "../screen/previewUrl";
import ScreenUnavailable from "../screen/ScreenUnavailable";

export default function Preview() {
  // RequireCarePartner only renders this page once a carepartner is chosen
  const { activeCarePartnerId } = useAuth();
  const src = previewUrl(activeCarePartnerId as number);

  return (
    <div className="relative h-screen bg-page text-ink font-dawn">
      {src ? <CarepartnerScreen src={src} /> : <ScreenUnavailable />}
      <Link
        to="/"
        className={`absolute right-6 top-6 flex items-center gap-2 h-12 px-[18px] rounded-[10px] bg-accent hover:brightness-95 text-on-accent text-base font-bold ${FOCUS_RING}`}
      >
        <Minimize size={20} strokeWidth={2.2} aria-hidden="true" />
        Exit full screen
      </Link>
    </div>
  );
}
