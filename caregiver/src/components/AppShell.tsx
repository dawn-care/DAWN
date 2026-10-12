/**
 * AppShell.tsx
 *
 * User: Caregiver
 *
 * Layout route that places the left NavBar beside the active page.
 * Pages that must render without the nav (e.g. /preview) sit outside it.
 */

import { Outlet } from "react-router-dom";
import NavBar from "./NavBar";

export default function AppShell() {
  return (
    <div className="flex min-h-screen bg-[#FAF8F4] text-[#1E1A15] font-['Atkinson_Hyperlegible_Next','Atkinson_Hyperlegible',system-ui,sans-serif]">
      <NavBar />
      <main className="grow min-w-0">
        <Outlet />
      </main>
    </div>
  );
}
