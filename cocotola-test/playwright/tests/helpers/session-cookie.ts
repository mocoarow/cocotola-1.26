import { createCookieSessionStorage } from "react-router";

// IMPORTANT: this cookie spec must mirror cocotola-web/app/lib/auth/session.server.ts.
// The web app's loader rejects cookies that don't match (name, sameSite, path, maxAge,
// signing secret), so any change there must be reflected here, and vice versa.
// The `react-router` dependency is pinned to the same version as cocotola-web for
// the same reason — encoding/signing internals must agree.

type SessionData = {
  accessToken: string;
  refreshToken: string;
};

const SESSION_COOKIE_NAME = "__cocotola_session";

function getSessionSecret(): string {
  const secret = process.env.SESSION_SECRET;
  if (!secret) {
    throw new Error("SESSION_SECRET environment variable is required (must match cocotola-web/.env)");
  }
  return secret;
}

const storage = createCookieSessionStorage<SessionData>({
  cookie: {
    name: SESSION_COOKIE_NAME,
    httpOnly: true,
    path: "/",
    sameSite: "lax",
    secrets: [getSessionSecret()],
    // Local-only assumption: Playwright targets the dev server over plain HTTP,
    // so `secure: false` is required for the cookie to be sent. If these tests
    // ever run against an HTTPS staging environment, mirror cocotola-web's rule
    // (`secure: process.env.NODE_ENV === "production"`).
    secure: false,
    maxAge: 60 * 60 * 24 * 7,
  },
});

export async function buildSessionCookieValue(data: SessionData): Promise<string> {
  const session = await storage.getSession();
  session.set("accessToken", data.accessToken);
  session.set("refreshToken", data.refreshToken);
  const setCookieHeader = await storage.commitSession(session);
  const valuePart = setCookieHeader.split(";")[0]?.split("=").slice(1).join("=");
  if (!valuePart) {
    throw new Error("failed to extract cookie value from Set-Cookie header");
  }
  return valuePart;
}

export const sessionCookieName = SESSION_COOKIE_NAME;
