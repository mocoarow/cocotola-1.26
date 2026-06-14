import "@testing-library/jest-dom/vitest";

// SESSION_SECRET is required at import time by lib/auth/session.server.ts
// because createCookieSessionStorage runs at module-load. Tests that
// transitively import session.server (e.g. via lib/api/user-setting.server.ts'
// getUserPreferences 401 branch) need this default before per-test
// vi.stubEnv can take effect.
if (!process.env.SESSION_SECRET) {
  process.env.SESSION_SECRET = "test-session-secret";
}
