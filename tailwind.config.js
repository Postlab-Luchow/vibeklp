/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './web/templates/**/*.html',
    './web/static/js/**/*.js',
  ],
  safelist: [
    // Leaflet and MarkerCluster apply these classes at runtime, so they
    // don't appear in our HTML/JS sources. Without the safelist, the
    // tree-shaker would strip our overrides in @layer components.
    { pattern: /^leaflet-/ },
    { pattern: /^marker-cluster/ },
  ],
  darkMode: 'media',
  theme: {
    extend: {
      colors: {
        canvas: 'rgb(var(--canvas) / <alpha-value>)',
        surface: 'rgb(var(--surface) / <alpha-value>)',
        'surface-elevated': 'rgb(var(--surface-elevated) / <alpha-value>)',
        border: 'rgb(var(--border) / <alpha-value>)',
        'border-strong': 'rgb(var(--border-strong) / <alpha-value>)',
        ink: 'rgb(var(--ink) / <alpha-value>)',
        muted: 'rgb(var(--muted) / <alpha-value>)',
        accent: {
          DEFAULT: 'rgb(var(--accent) / <alpha-value>)',
          strong: 'rgb(var(--accent-strong) / <alpha-value>)',
          soft: 'rgb(var(--accent-soft) / <alpha-value>)',
        },
        warn: {
          bg: 'rgb(var(--warn-bg) / <alpha-value>)',
          text: 'rgb(var(--warn-text) / <alpha-value>)',
        },
        danger: {
          bg: 'rgb(var(--error-bg) / <alpha-value>)',
          text: 'rgb(var(--error-text) / <alpha-value>)',
        },
        success: {
          bg: 'rgb(var(--success-bg) / <alpha-value>)',
          text: 'rgb(var(--success-text) / <alpha-value>)',
        },
      },
      fontFamily: {
        sans: [
          '-apple-system',
          'BlinkMacSystemFont',
          '"Segoe UI"',
          'Inter',
          'system-ui',
          'sans-serif',
        ],
      },
      boxShadow: {
        soft: '0 1px 2px rgba(0, 0, 0, 0.04), 0 2px 8px rgba(0, 0, 0, 0.05)',
        elevated: '0 12px 32px rgba(0, 0, 0, 0.12)',
      },
      borderRadius: {
        xl: '0.875rem',
        '2xl': '1.125rem',
      },
    },
  },
  corePlugins: {
    // We disable the preflight border-color reset so plain `.border` keeps using
    // our --border CSS variable via @layer base below.
  },
  plugins: [],
};
