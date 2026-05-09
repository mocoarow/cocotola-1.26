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
  // Casbin v3's policy enforcer is not safe under concurrent AddPolicyForUser calls
  // (concurrent map iteration and map write panics). On CI we serialize workers to
  // avoid the race; locally we keep the default for speed.
  workers: process.env.CI ? 1 : undefined,
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
