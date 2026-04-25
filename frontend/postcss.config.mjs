// PostCSS processes CSS after Tailwind expansion.
// Here we only activate the Tailwind PostCSS plugin.
const config = {
  plugins: {
    "@tailwindcss/postcss": {},
  },
};

export default config;
