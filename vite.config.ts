import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  // In development the Go API runs separately; in production it serves this
  // build from its own binary.
  server: {
    proxy: {
      "/api": "http://localhost:3000",
    },
  },
});
