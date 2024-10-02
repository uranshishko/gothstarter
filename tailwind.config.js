/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./**/*.{html,templ,go}"],
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      {
        custom: {
          primary: "#00b7b5",
          "primary-content": "#000c0c",
          secondary: "#00b800",
          "secondary-content": "#000c00",
          accent: "#316000",
          "accent-content": "#d3decd",
          neutral: "#000e09",
          "neutral-content": "#c4c8c6",
          "base-100": "#fefcff",
          "base-200": "#dddbde",
          "base-300": "#bdbbbe",
          "base-content": "#161616",
          info: "#009cff",
          "info-content": "#000916",
          success: "#11bb2b",
          "success-content": "#000d01",
          warning: "#da5500",
          "warning-content": "#110200",
          error: "#ff7386",
          "error-content": "#160506",
        },
      },
    ],
  },
};
