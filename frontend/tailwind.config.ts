import type { Config } from "tailwindcss";
import defaultTheme from "tailwindcss/defaultTheme";
import flowbitePlugin from "flowbite/plugin";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./lib/**/*.{js,ts,jsx,tsx,mdx}",
    "./node_modules/flowbite-react/**/*.{js,jsx,ts,tsx}",
    "./node_modules/flowbite/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        coffee: {
          night: "#0f0603",
          bean: "#2d140d",
          crema: "#f4d9b2",
          foam: "#fff9f2",
          bloom: "#c6784f",
        },
      },
      fontFamily: {
        sans: ["var(--font-inter)", ...defaultTheme.fontFamily.sans],
      },
      backgroundImage: {
        "coffee-radial":
          "radial-gradient(circle at 20% 20%, rgba(255, 249, 242, 0.12), transparent 55%), radial-gradient(circle at 80% 0%, rgba(198, 120, 79, 0.25), transparent 60%)",
      },
      boxShadow: {
        oracle: "0 35px 120px rgba(12, 5, 3, 0.55)",
      },
    },
  },
  plugins: [flowbitePlugin],
};

export default config;
