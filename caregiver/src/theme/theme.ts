/**
 * theme.ts
 *
 * User: Caregiver
 *
 * The caregiver's light/dark choice. <html data-theme> is always "light" or
 * "dark", and index.css keys the dark values off it. "system" resolves to
 * whichever the operating system prefers and follows it when it changes.
 * The inline script in index.html sets it the same way before the first paint;
 * watchTheme (called once in main.tsx) keeps it current afterwards, including
 * when another tab changes the choice.
 */

export type ThemeChoice = "light" | "dark" | "system";

const STORAGE_KEY = "theme";

const darkQuery = window.matchMedia("(prefers-color-scheme: dark)");
const listeners = new Set<() => void>();

let current: ThemeChoice = readSaved();

export function isThemeChoice(value: unknown): value is ThemeChoice {
  return value === "light" || value === "dark" || value === "system";
}

function readSaved(): ThemeChoice {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    return isThemeChoice(saved) ? saved : "system";
  } catch {
    return "system";
  }
}

function apply() {
  const dark = current === "dark" || (current === "system" && darkQuery.matches);
  document.documentElement.setAttribute("data-theme", dark ? "dark" : "light");
}

function update(choice: ThemeChoice) {
  current = choice;
  apply();
  listeners.forEach((listener) => listener());
}

/** The current choice */
export function getTheme(): ThemeChoice {
  return current;
}

/** Applies the choice at once and saves it for next time */
export function setTheme(choice: ThemeChoice) {
  update(choice);
  try {
    localStorage.setItem(STORAGE_KEY, choice);
  } catch {
    // Storage can be blocked; the choice then lasts until the page closes
  }
}

/** Calls listener whenever the choice changes; returns a function that stops it */
export function subscribeTheme(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

/** Follows the operating system on System, and choices made in other tabs */
export function watchTheme() {
  apply();
  darkQuery.addEventListener("change", apply);
  window.addEventListener("storage", (event) => {
    if (event.key === STORAGE_KEY) update(isThemeChoice(event.newValue) ? event.newValue : "system");
  });
}
