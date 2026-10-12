/**
 * Home.tsx
 *
 * User: Caregiver
 *
 * Carepartner Screen page at `/`, and the redirect target after login,
 * account setup, and carepartner selection. Placeholder until the page
 * is built.
 */

import { Link } from "react-router-dom";

export default function Home() {
  return (
    <div className="p-7">
      <h1 className="text-3xl font-extrabold">Carepartner Screen</h1>
      <Link to="/settings" className="underline">
        Settings
      </Link>
    </div>
  );
}
