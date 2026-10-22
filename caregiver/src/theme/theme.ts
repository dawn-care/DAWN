/**
 * theme.ts
 *
 * User: Caregiver
 *
 * The caregiver's light/dark choice. "system" follows the operating system and
 * leaves <html data-theme> unset, so the prefers-color-scheme rule in index.css
 * decides; "light" and "dark" set it, which wins over the system setting.
 * The inline script in index.html applies the saved choice before the first paint.
 */

export type ThemeChoice = "light" | "dark" | "system";

const STORAGE_KEY = "theme";

export function isThemeChoice(value: unknown): value is ThemeChoice {
  return value === "light" || value === "dark" || value === "system";
}

export function loadTheme(): ThemeChoice {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    return isThemeChoice(saved) ? saved : "system";
  } catch {
    return "system";
  }
}

export function applyTheme(choice: ThemeChoice) {
  if (choice === "system") {
    document.documentElement.removeAttribute("data-theme");
  } else {
    document.documentElement.setAttribute("data-theme", choice);
  }
  try {
    localStorage.setItem(STORAGE_KEY, choice);
  } catch {
    // Storage can be blocked; the choice then lasts until the page closes
  }
}
