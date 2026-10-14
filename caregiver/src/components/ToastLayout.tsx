/**
 * ToastLayout.tsx
 *
 * User: Caregiver
 *
 * Layout route for pages that sit outside AppShell (login, settings, setup,
 * preview). It adds the toast viewport and renders the page unchanged, so
 * toasts shown from these pages are visible too. Pages inside AppShell get
 * their viewport from AppShell instead.
 */

import { Outlet } from "react-router-dom";
import { ToastViewport } from "./ui/Toast";

export default function ToastLayout() {
  return (
    <>
      <ToastViewport />
      <Outlet />
    </>
  );
}
