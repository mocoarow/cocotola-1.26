import { isRedirectResponse } from "~/lib/route-error";

import type { BackendClient } from "./client.server";

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
 * Centralized here so the loaders in routes/workbooks.tsx and
 * routes/dashboard.tsx (and any future surface that needs the same fields)
 * cannot drift apart on the response shape.
 */
export async function getUserPreferences(client: BackendClient): Promise<UserPreferences> {
  return client.get("auth", "/api/v1/auth/me", {
    source: "user-setting:getUserPreferences",
    errorMessage: "Failed to load user preferences",
    parse: parseUserPreferencesResponse,
  });
}

/**
 * Like getUserPreferences, but returns null when /auth/me fails so the page
 * can still render. The 3xx redirect to /login (fired on 401) is rethrown.
 */
export async function getUserPreferencesOrNull(
  client: BackendClient,
): Promise<UserPreferences | null> {
  try {
    return await getUserPreferences(client);
  } catch (e) {
    if (isRedirectResponse(e)) throw e;
    console.error("[user-setting] /auth/me failed (degraded render):", e);
    return null;
  }
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

async function putUserSetting(
  client: BackendClient,
  path: string,
  body: unknown,
  source: string,
): Promise<void> {
  await client.put("auth", path, {
    source,
    errorMessage: `Failed to ${source}`,
    body,
  });
}

/** Updates the authenticated user's preferred language. */
export async function updateUserLanguage(client: BackendClient, language: string): Promise<void> {
  console.info(`[user-setting] updateUserLanguage called: languageLength=${language.length}`);
  await putUserSetting(
    client,
    "/api/v1/auth/user-setting/language",
    { language },
    "user-setting:updateUserLanguage",
  );
  console.info("[user-setting] updateUserLanguage succeeded");
}

/** Updates the authenticated user's daily problem goal. */
export async function updateUserDailyGoal(client: BackendClient, dailyGoal: number): Promise<void> {
  console.info(`[user-setting] updateUserDailyGoal called: dailyGoal=${dailyGoal}`);
  await putUserSetting(
    client,
    "/api/v1/auth/user-setting/daily-goal",
    { dailyGoal },
    "user-setting:updateUserDailyGoal",
  );
  console.info("[user-setting] updateUserDailyGoal succeeded");
}

/** Updates the authenticated user's preferred IANA timezone. */
export async function updateUserTimezone(client: BackendClient, timezone: string): Promise<void> {
  console.info(`[user-setting] updateUserTimezone called: timezoneLength=${timezone.length}`);
  await putUserSetting(
    client,
    "/api/v1/auth/user-setting/timezone",
    { timezone },
    "user-setting:updateUserTimezone",
  );
  console.info("[user-setting] updateUserTimezone succeeded");
}
