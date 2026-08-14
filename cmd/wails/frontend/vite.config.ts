import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import wails from "@wailsio/runtime/plugins/vite";

/// <reference types="vitest/config" />

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [svelte(), wails("./bindings")],
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
        inline: [/svelte/, /@testing-library/],
      },
    },
  },
});
