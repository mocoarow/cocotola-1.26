import { redirectOnUnauthorized } from "~/lib/auth/session.server";

import { fetchWithIdToken, getQuestionUrl } from "./fetch.server";

export type DashboardDailyItem = {
  date: string;
  answeredCount: number;
  correctCount: number;
};

export type DashboardResponse = {
  from: string;
  to: string;
  days: DashboardDailyItem[];
  currentStreak: number;
  longestStreak: number;
  todayCount: number;
  todayCorrect: number;
  activeDays: number;
  totalAnswered: number;
  totalCorrect: number;
};

// Min / max mirror studyservice.{Min,Max}DashboardDays on the Go side.
// DEFAULT is purely an HTTP-client convention (the backend has its own
// 365 default for the days query parameter) so it lives only on this
// boundary. Centralising it as a constant keeps the clamp fallback and
// the function default in sync — a previous version used the literal
// 365 in two places, which would have drifted on the next bump.
const MIN_DASHBOARD_DAYS = 7;
const MAX_DASHBOARD_DAYS = 730;
const DEFAULT_DASHBOARD_DAYS = 365;

function clampDays(days: number): number {
  if (!Number.isFinite(days)) return DEFAULT_DASHBOARD_DAYS;
  const rounded = Math.floor(days);
  if (rounded < MIN_DASHBOARD_DAYS) return MIN_DASHBOARD_DAYS;
  if (rounded > MAX_DASHBOARD_DAYS) return MAX_DASHBOARD_DAYS;
  return rounded;
}

/**
 * Fetches the user-scoped study dashboard: per-day contribution buckets,
 * streak counters, and today's progress. The caller must supply both the
 * user's local "today" (YYYY-MM-DD) and IANA timezone so the backend
 * windows the response in the user's calendar regardless of where the
 * request is served from.
 *
 * Routes through redirectOnUnauthorized so a 401 from the dashboard
 * endpoint converts to destroySession + redirect("/login") — symmetric
 * with getUserPreferences. Without this funnel a stale tab whose token
 * expired between the /auth/me and the /study/dashboard call would hit
 * the React Router error boundary on the second request with no path
 * back to the login screen.
 */
export async function getDashboard(
  request: Request,
  accessToken: string,
  localDateKey: string,
  timezone: string,
  days: number = DEFAULT_DASHBOARD_DAYS,
): Promise<DashboardResponse> {
  const baseUrl = getQuestionUrl();
  const params = new URLSearchParams({ days: String(clampDays(days)) });
  const url = `${baseUrl}/api/v1/study/dashboard?${params.toString()}`;

  const response = await redirectOnUnauthorized(
    request,
    await fetchWithIdToken(baseUrl, url, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
        "X-Local-Date": localDateKey,
        "X-Local-Timezone": timezone,
      },
    }),
    "dashboard:getDashboard",
  );

  if (!response.ok) {
    console.error(`[dashboard] getDashboard failed: status=${response.status}, url=${url}`);
    throw new Response("Failed to fetch dashboard", { status: response.status });
  }

  return (await response.json()) as DashboardResponse;
}
