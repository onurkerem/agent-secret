import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  site: "https://agent-secret.dev",
  vite: {
    plugins: [tailwindcss()],
  },
});
