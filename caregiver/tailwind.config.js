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
        },
      },
      fontFamily: {
        dawn: ["Atkinson Hyperlegible Next", "Atkinson Hyperlegible", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
}
