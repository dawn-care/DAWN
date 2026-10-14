/**
 * AppShell.tsx
 *
 * User: Caregiver
 *
 * Layout route that places the left NavBar beside the active page.
 * Pages that must render without the nav (e.g. /preview) sit outside it.
 * The ToastProvider itself is mounted at the app root (main.tsx); the shell
 * only places the viewport that draws toasts beside the nav bar.
 */

import { Outlet } from "react-router-dom";
import NavBar from "./NavBar";
import { ToastViewport } from "./ui/Toast";

export default function AppShell() {
  return (
    <div className="flex min-h-screen bg-dawn-page text-dawn-ink font-dawn">
      <NavBar />
      {/* tabIndex -1 lets a closing toast return focus here, without adding a tab stop */}
      <main tabIndex={-1} className="grow min-w-0 outline-none">
        {/* Toasts show top center of the page content, beside the nav bar */}
        <ToastViewport />
        <Outlet />
      </main>
    </div>
  );
}
