/**
 * Home.tsx
 *
 * User: Caregiver
 *
 * Placeholder landing page shown after login, account setup, or
 * carepartner selection. To be replaced by the new app shell.
 */

import { Link } from "react-router-dom";

export default function Home() {
  return (
    <main className="p-8">
      <h1 className="text-2xl font-semibold">StartUp</h1>
      <Link to="/settings" className="underline">
        Settings
      </Link>
    </main>
  );
}
