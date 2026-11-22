import { z } from "zod";

import type { BackendClient } from "./client.server";

const dashboardDailyItemSchema = z.object({
  date: z.string(),
  answeredCount: z.number().int(),
  correctCount: z.number().int(),
});

export type DashboardDailyItem = z.infer<typeof dashboardDailyItemSchema>;

const dashboardResponseSchema = z.object({
  from: z.string(),
  to: z.string(),
  days: z.array(dashboardDailyItemSchema),
  currentStreak: z.number().int().nonnegative(),
  longestStreak: z.number().int().nonnegative(),
  todayCount: z.number().int().nonnegative(),
  todayCorrect: z.number().int().nonnegative(),
  activeDays: z.number().int().nonnegative(),
  totalAnswered: z.number().int().nonnegative(),
  totalCorrect: z.number().int().nonnegative(),
});

export type DashboardResponse = z.infer<typeof dashboardResponseSchema>;

const parseDashboard = (raw: unknown): DashboardResponse => dashboardResponseSchema.parse(raw);

// Min / max mirror studyservice.{Min,Max}DashboardDays on the Go side.
// DEFAULT is purely an HTTP-client convention (the backend has its own
// 365 default for the days query parameter) so it lives only on this
// boundary. Centralizing it as a constant keeps the clamp fallback and
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
 */
export async function getDashboard(
  client: BackendClient,
  localDateKey: string,
  timezone: string,
  days: number = DEFAULT_DASHBOARD_DAYS,
): Promise<DashboardResponse> {
  const params = new URLSearchParams({ days: String(clampDays(days)) });

  return client.get("question", `/api/v1/study/dashboard?${params.toString()}`, {
    source: "dashboard:getDashboard",
    errorMessage: "Failed to fetch dashboard",
    headers: { "X-Local-Date": localDateKey, "X-Local-Timezone": timezone },
    parse: parseDashboard,
  });
}
