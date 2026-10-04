import path from "path"
import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"

const apiURL = process.env.ASSIETTE_API_URL ?? "http://localhost:8080"
const proxied = ["/api", "/login", "/signup", "/logout"]

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: Object.fromEntries(proxied.map((p) => [p, apiURL])),
  },
})
