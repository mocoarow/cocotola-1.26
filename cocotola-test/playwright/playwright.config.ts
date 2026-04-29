import { defineConfig } from "@playwright/test";
import { config as loadDotenv } from "dotenv";
import { existsSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
for (const candidate of [".env.local", "env.local", "env"]) {
  const path = resolve(here, candidate);
  if (existsSync(path)) {
    loadDotenv({ path });
    break;
  }
}

const baseURL = process.env.AUTH_BASE_URL ?? "http://localhost:8000";

export default defineConfig({
  testDir: "tests",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 0,
  reporter: [["list"]],
  use: {
    baseURL,
    extraHTTPHeaders: {
      "Content-Type": "application/json",
    },
    trace: "retain-on-failure",
  },
});
