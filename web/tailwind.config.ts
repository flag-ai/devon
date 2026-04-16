import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        // Catppuccin Mocha palette (dark).
        mocha: {
          base: "#1e1e2e",
          mantle: "#181825",
          crust: "#11111b",
          surface0: "#313244",
          surface1: "#45475a",
          text: "#cdd6f4",
          subtext1: "#bac2de",
          subtext0: "#a6adc8",
          overlay1: "#7f849c",
          blue: "#89b4fa",
          mauve: "#cba6f7",
          green: "#a6e3a1",
          red: "#f38ba8",
          peach: "#fab387",
          yellow: "#f9e2af",
          lavender: "#b4befe",
          sky: "#89dcfe",
          teal: "#94e2d5",
        },
      },
    },
  },
  plugins: [],
};

export default config;
