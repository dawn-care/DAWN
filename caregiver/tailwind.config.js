/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      // Palette and font from the caregiver mockups
      colors: {
        dawn: {
          page: "#FAF8F4",
          preview: "#F2EADB",
          sidebar: "#F1ECE3",
          line: "#E2DCD2",
          "line-strong": "#D6CEC1",
          ink: "#1E1A15",
          muted: "#5A5147",
          avatar: "#E8DDCB",
          "avatar-ink": "#6B5E4E",
          accent: "#9A3412",
          "accent-hover": "#7C2D12",
          // Brighter than the mockup's #3F6B3A, which read as near-black; 5.3:1 on dawn-page
          success: "#25772B",
          control: "#8A7F70",
          // The screen editor's grid
          dash: "#B5A68C",
          "tile-line": "#E2D6C2",
        },
      },
      fontFamily: {
        dawn: ["Atkinson Hyperlegible Next", "Atkinson Hyperlegible", "system-ui", "sans-serif"],
      },
      boxShadow: {
        toast: "0 8px 24px rgba(30, 26, 21, 0.3)",
        popover: "0 8px 24px rgba(30, 26, 21, 0.18)",
        dialog: "0 12px 32px rgba(30, 26, 21, 0.25)",
      },
    },
  },
  plugins: [],
}
