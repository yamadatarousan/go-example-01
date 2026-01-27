import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "e2e",
  timeout: 30_000,
  expect: {
    timeout: 5_000,
  },
  use: {
    baseURL: "http://localhost:3010",
    trace: "on-first-retry",
  },
  webServer: {
    command: "npm run dev -- --port 3010",
    url: "http://localhost:3010",
    reuseExistingServer: true,
    env: {
      USE_MOCK: "true",
      API_BASE_URL: "http://localhost:8080",
    },
  },
});
