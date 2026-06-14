import { createCookieSessionStorage, redirect } from "react-router";

// IMPORTANT: the cookie spec below is mirrored in
// cocotola-test/playwright/tests/helpers/session-cookie.ts so that Playwright
// can mint cookies the loaders here will accept. Any change to the cookie name,
// sameSite, path, or maxAge must be reflected in the test helper as well.

type SessionData = {
  accessToken: string;
  refreshToken: string;
};

function getSessionSecret(): string {
  const secret = process.env.SESSION_SECRET;
  if (!secret) {
    throw new Error("SESSION_SECRET environment variable is required");
  }
  return secret;
}

const sessionStorage = createCookieSessionStorage<SessionData>({
  cookie: {
    name: "__cocotola_session",
    httpOnly: true,
    path: "/",
    sameSite: "lax",
    secrets: [getSessionSecret()],
    secure: process.env.NODE_ENV === "production",
    maxAge: 60 * 60 * 24 * 7, // 7 days
  },
});

export async function getSession(request: Request) {
  return sessionStorage.getSession(request.headers.get("Cookie"));
}

export async function commitSession(session: Awaited<ReturnType<typeof getSession>>) {
  return sessionStorage.commitSession(session);
}

export async function destroySession(session: Awaited<ReturnType<typeof getSession>>) {
  return sessionStorage.destroySession(session);
}

/**
 * If `response` carries HTTP 401, destroy the React Router session and
 * throw a `redirect("/login")` Response. Otherwise resolves with `response`
 * unchanged so the caller can fall through to its normal handling.
 *
 * Every server-side fetch that authenticates against an internal service
 * (i.e. anything going through fetchWithIdToken with a Bearer token) must
 * funnel its 401 branch through this helper. Without the funnel a stale
 * tab whose token expired between loaders would produce a generic React
 * Router error boundary on whichever endpoint replied first — instead of
 * the user being sent to /login with the cookie cleared. The "source" tag
 * appears in the structured log so it is clear which call site triggered
 * the redirect during incident review.
 */
export async function redirectOnUnauthorized(
  request: Request,
  response: Response,
  source: string,
): Promise<Response> {
  if (response.status !== 401) return response;

  console.info(`[${source}] backend returned 401, destroying session`);
  const session = await getSession(request);
  throw redirect("/login", {
    headers: { "Set-Cookie": await destroySession(session) },
  });
}
