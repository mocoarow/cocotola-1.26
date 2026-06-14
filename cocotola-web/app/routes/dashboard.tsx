import { useTranslation } from "react-i18next";
import { useLoaderData } from "react-router";

import { ContributionGraph } from "~/components/dashboard/contribution-graph";
import { DailyGoalCard } from "~/components/dashboard/daily-goal-card";
import { StreakCard } from "~/components/dashboard/streak-card";
import { getDashboard } from "~/lib/api/dashboard.server";
import { getUserPreferences } from "~/lib/api/user-setting.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { getLocalDateKey } from "~/lib/format/local-date";

import type { Route } from "./+types/dashboard";

/**
 * Server loader: reads the user's timezone + daily goal from /auth/me and
 * fetches a 365-day dashboard window. "Today" is computed server-side
 * from the persisted preference, so the first paint matches the timezone
 * the user has actively chosen — if the browser later disagrees (the
 * "user traveled" case) they have to update the preference manually,
 * which is the simplest behavior that keeps server-rendered and client
 * views consistent.
 *
 * The /auth/me lookup goes through getUserPreferences so the response
 * shape, default-field fallbacks, and 401-redirect-to-login behavior stay
 * consistent with routes/workbooks.tsx — without this funnel a stale tab
 * pointed at /dashboard would render the React Router error boundary
 * instead of being redirected back to the login screen.
 */
export async function loader({ request }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);
  const prefs = await getUserPreferences(request, accessToken);
  const todayKey = getLocalDateKey(prefs.timezone);
  // Omitting `days` so the function default (DEFAULT_DASHBOARD_DAYS in
  // dashboard.server.ts) is the single source of truth. A `365` literal
  // here would have made the constant in the client library cosmetic.
  const dashboard = await getDashboard(request, accessToken, todayKey, prefs.timezone);
  return {
    dashboard,
    dailyGoal: prefs.dailyGoal,
  };
}

export default function DashboardPage() {
  const { dashboard, dailyGoal } = useLoaderData<typeof loader>();
  const { t, i18n } = useTranslation();

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold">{t("dashboard.title")}</h1>
        <p className="text-sm text-muted-foreground">{t("dashboard.subtitle")}</p>
      </header>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <StreakCard current={dashboard.currentStreak} longest={dashboard.longestStreak} />
        <DailyGoalCard todayCount={dashboard.todayCount} dailyGoal={dailyGoal} />
      </div>

      <ContributionGraph
        items={dashboard.days}
        from={dashboard.from}
        to={dashboard.to}
        locale={i18n.language}
      />

      <p className="text-xs text-muted-foreground">
        {t("dashboard.totals.active", { count: dashboard.activeDays })}
        {" · "}
        {t("dashboard.totals.answered", { count: dashboard.totalAnswered })}
      </p>
    </div>
  );
}
