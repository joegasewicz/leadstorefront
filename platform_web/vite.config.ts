import { defineConfig } from "vite";

export default defineConfig({
  build: {
    emptyOutDir: true,
    manifest: false,
    outDir: "static/assets",
    rollupOptions: {
      input: "src/main.ts",
      output: {
        entryFileNames: "app.js",
        assetFileNames: "app.css"
      }
    }
  }
});
