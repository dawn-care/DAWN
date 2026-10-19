/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./app/**/*.{js,jsx,ts,tsx}', './components/**/*.{js,jsx,ts,tsx}'],
  presets: [require('nativewind/preset')],
  // Class-based so the app is light by default and dark is opt-in. Media-based
  // dark mode conflicts with Expo setting the color scheme itself.
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Semantic tokens only — see global.css for the brand values behind these.
        page: 'rgb(var(--color-page) / <alpha-value>)',
        surface: 'rgb(var(--color-surface) / <alpha-value>)',
        'surface-raised': 'rgb(var(--color-surface-raised) / <alpha-value>)',
        nav: 'rgb(var(--color-nav) / <alpha-value>)',
        'on-nav': 'rgb(var(--color-on-nav) / <alpha-value>)',
        ink: 'rgb(var(--color-ink) / <alpha-value>)',
        'ink-muted': 'rgb(var(--color-ink-muted) / <alpha-value>)',
        link: 'rgb(var(--color-link) / <alpha-value>)',
        accent: 'rgb(var(--color-accent) / <alpha-value>)',
        'on-accent': 'rgb(var(--color-on-accent) / <alpha-value>)',
        border: 'rgb(var(--color-border) / <alpha-value>)',
        focus: 'rgb(var(--color-focus) / <alpha-value>)',
        status: {
          'gray-bg': 'rgb(var(--color-status-gray-bg) / <alpha-value>)',
          'gray-text': 'rgb(var(--color-status-gray-text) / <alpha-value>)',
          'blue-bg': 'rgb(var(--color-status-blue-bg) / <alpha-value>)',
          'blue-text': 'rgb(var(--color-status-blue-text) / <alpha-value>)',
          'green-bg': 'rgb(var(--color-status-green-bg) / <alpha-value>)',
          'green-text': 'rgb(var(--color-status-green-text) / <alpha-value>)',
          'yellow-bg': 'rgb(var(--color-status-yellow-bg) / <alpha-value>)',
          'yellow-text': 'rgb(var(--color-status-yellow-text) / <alpha-value>)',
          'purple-bg': 'rgb(var(--color-status-purple-bg) / <alpha-value>)',
          'purple-text': 'rgb(var(--color-status-purple-text) / <alpha-value>)',
          'orange-bg': 'rgb(var(--color-status-orange-bg) / <alpha-value>)',
          'orange-text': 'rgb(var(--color-status-orange-text) / <alpha-value>)',
          'red-bg': 'rgb(var(--color-status-red-bg) / <alpha-value>)',
          'red-text': 'rgb(var(--color-status-red-text) / <alpha-value>)',
          'pink-bg': 'rgb(var(--color-status-pink-bg) / <alpha-value>)',
          'pink-text': 'rgb(var(--color-status-pink-text) / <alpha-value>)',
        },
      },
    },
  },
  plugins: [],
};
