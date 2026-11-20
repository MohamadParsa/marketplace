import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/backoffice": "http://localhost:8080",
      "/catalog/icons": "http://localhost:8080",
    },
  },
  build: {
    outDir: "build",
  },
});
