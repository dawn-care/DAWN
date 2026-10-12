/**
 * NavBar.tsx
 *
 * User: Caregiver
 *
 * Left navigation bar shown on every page inside AppShell.
 *
 * Responsibilities:
 *  - Show the D.A.W.N. branding
 *  - Show the active carepartner's initials avatar and name
 *  - Link to Carepartner Screen, Calendar, and Photos, Videos & Audio
 *  - Link to account settings, labeled with the signed-in caregiver's name
 *
 * /family exists as a stub route but is intentionally not linked here yet.
 */

import { useEffect } from "react";
import type { ReactNode } from "react";
import { Link, NavLink } from "react-router-dom";
import { Calendar, Image, Monitor, User } from "lucide-react";
import { useAuth } from "../context/AuthContext";

/* ============================
   Static Data
============================ */

// TODO: hardcoded until the carepartner's facility is stored.
const FACILITY = "St. Ann's";

/**
 * Nav items, in display order
 */
const NAV_ITEMS: { to: string; label: string; end?: boolean; icon: ReactNode }[] = [
  {
    to: "/",
    label: "Carepartner Screen",
    end: true, // otherwise "/" would match every route
    icon: <Monitor size={22} aria-hidden="true" />,
  },
  {
    to: "/calendar",
    label: "Calendar",
    icon: <Calendar size={22} aria-hidden="true" />,
  },
  {
    to: "/media",
    label: "Photos, Videos & Audio",
    icon: <Image size={22} aria-hidden="true" />,
  },
];

/* ============================
   Helpers
============================ */

/**
 * Builds avatar initials from a name: the first letter of up to
 * the first two words ("Jane Doe" → "JD")
 */
function getInitials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0].toUpperCase())
    .join("");
}

/* ============================
   Component
============================ */

/**
 * NavBar Component
 *
 * Displays the branding, active carepartner, page links,
 * and signed-in caregiver for the app shell.
 */
export default function NavBar() {

  /* ============================
     Context
  ============================ */

  const API_URL = import.meta.env.VITE_API_URL;

  const { user, setUser, activeCarePartnerId } = useAuth();

  /* ============================
     Effects
  ============================ */

  /**
   * Reload the caregiver after a page refresh, since the auth context
   * only holds the user until the page reloads
   */
  useEffect(() => {
    if (user) return;

    const fetchCaregiver = async () => {
      try {
        const res = await fetch(`${API_URL}/me`, { credentials: "include" });
        if (res.ok) setUser(await res.json());
      } catch (err) {
        console.error("Error fetching caregiver:", err);
      }
    };

    fetchCaregiver();
  }, [user, setUser, API_URL]);

  /* ============================
     Derived Values
  ============================ */

  const carepartner = user?.carepartners?.find((cp) => cp.id === activeCarePartnerId);
  const carepartnerName = carepartner?.name ?? "";

  /* ============================
     Render
  ============================ */

  return (
    <nav
      aria-label="Main"
      className="w-1/5 min-w-[220px] max-w-[320px] shrink-0 flex flex-col gap-2.5 px-4 py-7 bg-dawn-sidebar border-r border-dawn-line"
    >
      <div>
        <div className="text-2xl font-extrabold">D.A.W.N.</div>
        <div className="text-base text-dawn-muted">Caregiver</div>
      </div>

      <div className="flex items-center gap-3 mt-3.5 mb-2.5">
        <div
          aria-hidden="true"
          className="w-12 h-12 shrink-0 rounded-full bg-dawn-avatar text-dawn-avatar-ink flex items-center justify-center text-lg font-bold"
        >
          {carepartnerName ? getInitials(carepartnerName) : <User size={24} />}
        </div>
        <div className="flex flex-col min-w-0">
          <div className="text-[17px] font-bold truncate">{carepartnerName}</div>
          <div className="text-base text-dawn-muted">{FACILITY}</div>
        </div>
      </div>

      {NAV_ITEMS.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          end={item.end}
          className={({ isActive }) =>
            `flex items-center gap-2.5 h-12 px-2.5 rounded-[10px] whitespace-nowrap text-base border-[1.5px] ${
              isActive
                ? "bg-white border-dawn-line-strong font-bold"
                : "border-transparent hover:bg-white/60"
            }`
          }
        >
          {item.icon}
          {item.label}
        </NavLink>
      ))}

      <div className="grow" />

      {/* Account settings, labeled with the signed-in caregiver's name */}
      <Link
        to="/settings"
        className="flex items-center gap-2.5 h-12 px-2.5 rounded-[10px] text-base hover:bg-white/60"
      >
        <User size={22} aria-hidden="true" />
        <span className="truncate">{user?.name}</span>
      </Link>
    </nav>
  );
}
