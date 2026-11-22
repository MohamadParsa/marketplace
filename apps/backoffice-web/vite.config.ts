import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
export default defineConfig(({ mode }) => {
  const target =
    loadEnv(mode, ".", "CHERRIVA_").CHERRIVA_API_PROXY ||
    "http://localhost:8080";
  return {
    plugins: [react()],
    server: {
      proxy: {
        "/backoffice": target,
        "/seller": target,
        "/catalog/icons": target,
        "/content": target,
      },
    },
    build: { outDir: "build" },
  };
});
