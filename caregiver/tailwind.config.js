const STATUSES = ["gray", "blue", "green", "yellow", "purple", "orange", "red", "pink"];
const tw = (name) => `rgb(var(--color-${name}) / <alpha-value>)`;

/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      // Semantic tokens only: their light and dark values are CSS variables in src/index.css,
      // stored as space-separated RGB channels so opacity modifiers (bg-accent/10) work
      colors: {
        page: tw("page"),
        surface: tw("surface"),
        "surface-raised": tw("surface-raised"),
        nav: tw("nav"),
        "on-nav": tw("on-nav"),
        ink: tw("ink"),
        "ink-muted": tw("ink-muted"),
        link: tw("link"),
        accent: tw("accent"),
        "on-accent": tw("on-accent"),
        border: tw("border"),
        focus: tw("focus"),
        status: Object.fromEntries(
          STATUSES.flatMap((name) => [
            [`${name}-bg`, tw(`status-${name}-bg`)],
            [`${name}-text`, tw(`status-${name}-text`)],
          ]),
        ),
      },
      borderColor: { DEFAULT: tw("border") },
      fontFamily: {
        dawn: ["Atkinson Hyperlegible Next", "Atkinson Hyperlegible", "system-ui", "sans-serif"],
      },
      boxShadow: {
        toast: "0 8px 24px rgba(30, 33, 64, 0.3)",
        popover: "0 8px 24px rgba(30, 33, 64, 0.18)",
        dialog: "0 12px 32px rgba(30, 33, 64, 0.25)",
      },
    },
  },
  plugins: [],
}
