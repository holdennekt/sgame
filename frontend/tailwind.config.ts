import type { Config } from "tailwindcss";
import plugin from "tailwindcss/plugin";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "var(--background)",
        "on-background": "var(--on-background)",
        surface: "var(--surface)",
        "surface-raised": "var(--surface-raised)",
        "on-surface": "var(--on-surface)",
        "on-surface-muted": "var(--on-surface-muted)",
        border: "var(--border)",
        primary: "var(--primary)",
        "primary-hover": "var(--primary-hover)",
        "on-primary": "var(--on-primary)",
        secondary: "var(--secondary)",
        "on-secondary": "var(--on-secondary)",
        danger: "var(--danger)",
      },
      boxShadow: {
        sm: "var(--shadow-sm)",
        DEFAULT: "var(--shadow)",
        md: "var(--shadow-md)",
        lg: "var(--shadow-lg)",
      },
      backgroundImage: {
        "gradient-radial": "radial-gradient(var(--tw-gradient-stops))",
        "gradient-conic":
          "conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))",
      },
    },
  },
  plugins: [
    plugin(({ addVariant }) => {
      addVariant("hover", "@media (hover: hover) { &:hover }");
      addVariant(
        "group-hover",
        "@media (hover: hover) { :merge(.group):hover & }"
      );
    }),
  ],
};
export default config;
