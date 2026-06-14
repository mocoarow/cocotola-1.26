import { redirectOnUnauthorized } from "~/lib/auth/session.server";

import { fetchWithIdToken } from "./fetch.server";

function getAuthUrl(): string {
  const url = process.env.AUTH_BASE_URL;
  if (!url) {
    throw new Error("AUTH_BASE_URL environment variable is required");
  }
  return url;
}

/**
 * Shape returned by GET /api/v1/auth/me. Keeping the type next to the only
 * function that calls the endpoint is intentional — every loader that
 * needs user identity or preferences must funnel through getUserPreferences
 * so that adding a new field to the backend response only ever requires
 * touching one frontend file.
 */
export type UserPreferences = {
  userId: string;
  loginId: string;
  organizationName: string;
  language: string;
  dailyGoal: number;
  timezone: string;
};

/**
 * Loads the authenticated user's identity and preferences from /auth/me.
 *
 * Centralised here so the loaders in routes/workbooks.tsx and
 * routes/dashboard.tsx (and any future surface that needs the same fields)
 * cannot drift apart on the response shape.
 *
 * 401 semantics matter for the layout consistency contract: when the
 * supplied access token is rejected we destroy the React Router session
 * and redirect to /login. Without this branch a stale tab opened against
 * the dashboard or any other layout that reads /auth/me would render
 * raw error markup with no way back to a login form.
 */
export async function getUserPreferences(
  request: Request,
  accessToken: string,
): Promise<UserPreferences> {
  const authUrl = getAuthUrl();
  const url = `${authUrl}/api/v1/auth/me`;

  const response = await redirectOnUnauthorized(
    request,
    await fetchWithIdToken(authUrl, url, {
      headers: { Authorization: `Bearer ${accessToken}` },
    }),
    "user-setting:getUserPreferences",
  );

  if (!response.ok) {
    console.error(
      `[user-setting] getUserPreferences failed: status=${response.status}, url=${url}`,
    );
    throw new Response("Failed to load user preferences", { status: response.status });
  }

  const raw = (await response.json()) as unknown;
  return parseUserPreferencesResponse(raw);
}

/**
 * Validates each field's runtime type before threading it into a
 * UserPreferences value. `as Partial<UserPreferences>` would happily
 * pass `dailyGoal: "10"` (string) through `?? 10`, leaving a
 * stringly-typed number to crash downstream consumers; this guard
 * forces a well-typed fallback instead. The defaults mirror the auth
 * backend's domain.Default* constants — they are repeated here only
 * because TypeScript cannot reach across the language boundary, not
 * because they should diverge.
 */
function parseUserPreferencesResponse(raw: unknown): UserPreferences {
  const obj = (raw === null || typeof raw !== "object" ? {} : raw) as Record<string, unknown>;

  return {
    userId: pickString(obj.userId, ""),
    loginId: pickString(obj.loginId, ""),
    organizationName: pickString(obj.organizationName, ""),
    language: pickString(obj.language, "en"),
    dailyGoal: pickFiniteInt(obj.dailyGoal, 10),
    timezone: pickString(obj.timezone, "Asia/Tokyo"),
  };
}

function pickString(v: unknown, fallback: string): string {
  return typeof v === "string" ? v : fallback;
}

function pickFiniteInt(v: unknown, fallback: number): number {
  return typeof v === "number" && Number.isFinite(v) ? v : fallback;
}

/**
 * Routes every user-setting update through this helper so the 401 -> /login
 * branch lives in exactly one place. The three update endpoints are
 * structurally identical, only differing in path and body, so the
 * boilerplate can collapse without obscuring intent at the call sites.
 */
async function putUserSetting(
  request: Request,
  accessToken: string,
  path: string,
  body: unknown,
  source: string,
): Promise<void> {
  const authUrl = getAuthUrl();
  const url = `${authUrl}${path}`;

  const response = await redirectOnUnauthorized(
    request,
    await fetchWithIdToken(authUrl, url, {
      method: "PUT",
      headers: {
        Authorization: `Bearer ${accessToken}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    }),
    source,
  );

  if (!response.ok) {
    console.error(`[user-setting] ${source} failed: status=${response.status}, url=${url}`);
    throw new Response(`Failed to ${source}`, { status: response.status });
  }
}

/** Updates the authenticated user's preferred language. */
export async function updateUserLanguage(
  request: Request,
  accessToken: string,
  language: string,
): Promise<void> {
  console.info(`[user-setting] updateUserLanguage called: language=${language}`);
  await putUserSetting(
    request,
    accessToken,
    "/api/v1/auth/user-setting/language",
    { language },
    "user-setting:updateUserLanguage",
  );
  console.info("[user-setting] updateUserLanguage succeeded");
}

/** Updates the authenticated user's daily problem goal. */
export async function updateUserDailyGoal(
  request: Request,
  accessToken: string,
  dailyGoal: number,
): Promise<void> {
  console.info(`[user-setting] updateUserDailyGoal called: dailyGoal=${dailyGoal}`);
  await putUserSetting(
    request,
    accessToken,
    "/api/v1/auth/user-setting/daily-goal",
    { dailyGoal },
    "user-setting:updateUserDailyGoal",
  );
  console.info("[user-setting] updateUserDailyGoal succeeded");
}

/** Updates the authenticated user's preferred IANA timezone. */
export async function updateUserTimezone(
  request: Request,
  accessToken: string,
  timezone: string,
): Promise<void> {
  console.info(`[user-setting] updateUserTimezone called: timezone=${timezone}`);
  await putUserSetting(
    request,
    accessToken,
    "/api/v1/auth/user-setting/timezone",
    { timezone },
    "user-setting:updateUserTimezone",
  );
  console.info("[user-setting] updateUserTimezone succeeded");
}
