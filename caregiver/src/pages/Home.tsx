/**
 * Home.tsx
 *
 * User: Caregiver
 *
 * Carepartner Screen page at `/`, and the redirect target after login,
 * account setup, and carepartner selection. Shows the carepartner's saved
 * screen as the tablet draws it (ScreenPreview's iframe), and loads the saved
 * layout that editing starts from.
 *
 * FUTURE IMPLEMENTATION (#360): Edit screen with the add bar, the pencil and
 * bin on each widget, their popups, and Save changes are added in later steps.
 */

import { Check } from "lucide-react";
import InfoPopover from "../components/ui/InfoPopover";
import { FOCUS_RING } from "../components/ui/styles";
import { useAuth } from "../context/AuthContext";
import ScreenPreview from "../screen/ScreenPreview";
import { useSavedLayout } from "../screen/useSavedLayout";

const PAGE_INFO =
  "This is the screen she sees all day, whenever no video or audio is playing. Press Edit screen to change it. Nothing changes on her tablet until you press Save changes.";

export default function Home() {
  // RequireCarePartner only renders this page once a carepartner is chosen
  const { activeCarePartnerId } = useAuth();
  const carepartnerId = activeCarePartnerId as number;
  // The saved layout is what editing starts from; the screen itself doesn't need it
  const { state, reload } = useSavedLayout(carepartnerId);

  return (
    <div className="flex flex-col gap-3 px-7 py-5">
      <div className="flex flex-wrap items-center justify-between gap-4 min-h-11">
        <div className="flex items-center gap-1">
          <h1 className="m-0 text-3xl font-extrabold">Carepartner Screen</h1>
          <InfoPopover label="About this page">{PAGE_INFO}</InfoPopover>
        </div>
        {state.status === "ready" && (
          <div className="flex items-center gap-1.5 text-base text-dawn-success">
            <Check size={18} strokeWidth={2.4} aria-hidden="true" />
            Saved
          </div>
        )}
      </div>

      {state.status === "error" && (
        <div role="alert" className="flex flex-wrap items-center gap-3 text-base text-dawn-muted">
          <p className="m-0">{state.message}</p>
          <button
            type="button"
            onClick={reload}
            className={`h-11 px-[18px] rounded-[10px] bg-dawn-accent hover:bg-dawn-accent-hover text-white text-base font-bold cursor-pointer ${FOCUS_RING}`}
          >
            Try again
          </button>
        </div>
      )}

      <ScreenPreview carepartnerId={carepartnerId} />

      {state.status === "ready" && state.layout.dropped.length > 0 && (
        <p className="m-0 text-center text-[15px] text-dawn-muted">
          {state.layout.dropped.length === 1
            ? "1 thing saved on her screen can't be shown on her tablet. It's removed the next time you save."
            : `${state.layout.dropped.length} things saved on her screen can't be shown on her tablet. They're removed the next time you save.`}
        </p>
      )}
    </div>
  );
}
