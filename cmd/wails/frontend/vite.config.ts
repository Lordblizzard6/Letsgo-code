import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import wails from "@wailsio/runtime/plugins/vite";

/// <reference types="vitest/config" />

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [react(), wails("./bindings")],
  resolve: {
    conditions: ["browser"],
  },
  ssr: {
    resolve: {
      conditions: ["browser", "module", "import"],
    },
  },
  test: {
    environment: "jsdom",
    include: ["src/**/*.test.ts"],
    server: {
      deps: {
        inline: [/@testing-library/],
      },
    },
  },
});
