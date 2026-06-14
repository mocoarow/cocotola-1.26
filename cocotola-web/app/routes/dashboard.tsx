import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Form, useActionData, useLoaderData, useNavigation } from "react-router";

import { ContributionGraph } from "~/components/dashboard/contribution-graph";
import { DailyGoalCard } from "~/components/dashboard/daily-goal-card";
import { StreakCard } from "~/components/dashboard/streak-card";
import { Button } from "~/components/ui/button";
import { getDashboard } from "~/lib/api/dashboard.server";
import {
  getUserPreferences,
  updateUserDailyGoal,
  updateUserTimezone,
} from "~/lib/api/user-setting.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import {
  detectBrowserTimezone,
  getLocalDateKey,
  isValidIanaTimezoneShape,
} from "~/lib/format/local-date";
import { MAX_DAILY_GOAL, MIN_DAILY_GOAL } from "~/lib/format/user-setting-bounds";

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
    timezone: prefs.timezone,
  };
}

/**
 * Saves a single preference field. Catches per-field failures so the
 * caller can report partial success, but propagates Response thrown
 * with a 3xx redirect status — the user-setting update path converts
 * a backend 401 into `throw redirect("/login", ...)` via
 * redirectOnUnauthorized, and that redirect must reach React Router
 * unchanged. Without this funnel a session-expired tab would either
 * silently retry the now-stale token or land on a generic error page
 * instead of the login screen.
 *
 * Promise.all over the two updates would have surfaced the same generic
 * 500 even when one field had already been persisted, leaving the user
 * with no way to tell what was saved; sequential await + a list of
 * failed fields is the minimum fix that keeps the UI honest. A future
 * combined PATCH /user-setting endpoint would make this atomic at the
 * backend; until then we report partial state instead of hiding it.
 */
async function savePreferenceField(
  field: "dailyGoal" | "timezone",
  saveFn: () => Promise<void>,
): Promise<{ field: typeof field; ok: boolean }> {
  try {
    await saveFn();
    return { field, ok: true };
  } catch (err) {
    if (err instanceof Response && err.status >= 300 && err.status < 400) {
      // Redirect (e.g. session-expiry redirect to /login). React Router
      // recognizes thrown Responses with Location headers — leave them
      // alone so the navigation happens instead of being recorded as a
      // field-level save failure.
      throw err;
    }
    console.error(`[dashboard] save ${field} failed`, err);
    return { field, ok: false };
  }
}

// Generic body sent to the client for any action-level validation
// failure. The specific cause (intent mismatch, dailyGoal out of range,
// timezone shape) is recorded in the server log instead — the user's UI
// shape already prevents the legitimate paths from triggering these
// branches, so the only callers are tampered submits where revealing
// the exact validation rule offers no UX benefit.
const REJECTED_REQUEST_BODY = "Invalid request parameters";

export async function action({ request }: Route.ActionArgs) {
  const { accessToken } = await requireAuth(request);
  const formData = await request.formData();
  const intent = String(formData.get("intent") ?? "");
  if (intent !== "savePreferences") {
    console.warn(`[dashboard] action rejected: unknown intent (got "${intent}")`);
    throw new Response(REJECTED_REQUEST_BODY, { status: 400 });
  }

  const dailyGoalRaw = formData.get("dailyGoal");
  const timezone = String(formData.get("timezone") ?? "");

  // Require a digit-only string. Number.parseInt accepts partial parses
  // ("30abc" => 30, "0x1f" => 0), letting tampered form bodies sneak past
  // the range check below; restricting the shape with /^\d+$/ first
  // converts every malformed input into NaN.
  const dailyGoal =
    typeof dailyGoalRaw === "string" && /^\d+$/.test(dailyGoalRaw)
      ? Number.parseInt(dailyGoalRaw, 10)
      : Number.NaN;
  if (!Number.isFinite(dailyGoal) || dailyGoal < MIN_DAILY_GOAL || dailyGoal > MAX_DAILY_GOAL) {
    console.warn(`[dashboard] action rejected: dailyGoal out of range (got ${dailyGoal})`);
    throw new Response(REJECTED_REQUEST_BODY, { status: 400 });
  }
  if (!isValidIanaTimezoneShape(timezone)) {
    console.warn(
      `[dashboard] action rejected: timezone shape invalid (length=${timezone.length})`,
    );
    throw new Response(REJECTED_REQUEST_BODY, { status: 400 });
  }

  const goalResult = await savePreferenceField("dailyGoal", () =>
    updateUserDailyGoal(request, accessToken, dailyGoal),
  );
  const tzResult = await savePreferenceField("timezone", () =>
    updateUserTimezone(request, accessToken, timezone),
  );

  const failedFields = [goalResult, tzResult].filter((r) => !r.ok).map((r) => r.field);
  if (failedFields.length > 0) {
    return { ok: false as const, failedFields };
  }
  return { ok: true as const, savedAt: new Date().toISOString() };
}

export default function DashboardPage() {
  const { dashboard, dailyGoal, timezone } = useLoaderData<typeof loader>();
  const actionData = useActionData<typeof action>();
  const navigation = useNavigation();
  const { t, i18n } = useTranslation();
  // Draft state is seeded from loader once on mount and stays under the
  // user's control after that. Re-syncing from the loader on every
  // revalidation (the original implementation used a useEffect) wiped
  // mid-edit input whenever React Router refreshed the route — including
  // after a partial-failure submit, which is exactly when the user
  // needs to retry without re-typing the half that didn't save.
  const [draftGoal, setDraftGoal] = useState<number>(dailyGoal);
  const [draftTimezone, setDraftTimezone] = useState(timezone);

  const isSaving =
    navigation.state === "submitting" && navigation.formData?.get("intent") === "savePreferences";

  // The browser-detected timezone never changes for the lifetime of the
  // mounted component — memoise so we are not constructing a new
  // Intl.DateTimeFormat on every render. The value is wrong during SSR
  // (it would report the server's TZ); the hint paragraph is purely
  // advisory, so the post-hydration value taking over on the client is
  // good enough — no need to plumb this through the loader.
  const detectedBrowserTimezone = useMemo(() => detectBrowserTimezone(), []);

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

      <section
        aria-labelledby="dashboard-settings-heading"
        className="space-y-4 rounded-md border bg-card p-4"
      >
        <h2 id="dashboard-settings-heading" className="text-sm font-semibold">
          {t("dashboard.settings.heading")}
        </h2>
        <Form method="post" className="space-y-3">
          <input type="hidden" name="intent" value="savePreferences" />
          <div>
            <label htmlFor="dailyGoal" className="block text-sm font-medium">
              {t("dashboard.settings.dailyGoalLabel")}
            </label>
            {/*
              Treating empty as NaN (instead of letting Number("") === 0
              slip through) keeps the field visually empty when the user
              clears it and prevents the "0" submission that the server
              would reject with a 400. The `required` attribute blocks
              submit while the value is empty so the user gets browser-
              native feedback rather than discovering the rejection only
              after a network round trip.
            */}
            <input
              id="dailyGoal"
              name="dailyGoal"
              type="number"
              min={MIN_DAILY_GOAL}
              max={MAX_DAILY_GOAL}
              required
              value={Number.isFinite(draftGoal) ? draftGoal : ""}
              onChange={(e) => {
                const raw = e.target.value;
                setDraftGoal(raw === "" ? Number.NaN : Number(raw));
              }}
              className="mt-1 h-9 w-32 rounded-md border border-input bg-transparent px-2 text-sm"
            />
            <p className="mt-1 text-xs text-muted-foreground">
              {t("dashboard.settings.dailyGoalHint")}
            </p>
          </div>
          <div>
            <label htmlFor="timezone" className="block text-sm font-medium">
              {t("dashboard.settings.timezoneLabel")}
            </label>
            <input
              id="timezone"
              name="timezone"
              type="text"
              value={draftTimezone}
              onChange={(e) => setDraftTimezone(e.target.value)}
              maxLength={64}
              pattern="[A-Za-z_/+\-0-9]+"
              className="mt-1 h-9 w-72 rounded-md border border-input bg-transparent px-2 text-sm"
            />
            <p className="mt-1 text-xs text-muted-foreground">
              {/*
                Show the browser-detected zone as a hint so the user can
                spot a mismatch with their saved preference. The value
                comes from a useMemo above so we are not re-running
                Intl resolution on every render. Passing the persisted
                timezone to detectBrowserTimezone() would have made the
                hint indistinguishable from the saved value and defeated
                the point of showing it at all.
              */}
              {t("dashboard.settings.timezoneHint")} ({detectedBrowserTimezone})
            </p>
          </div>
          <Button type="submit" disabled={isSaving}>
            {isSaving ? t("common.saving") : t("dashboard.settings.save")}
          </Button>
          {actionData?.ok === true && (
            <p className="text-xs text-emerald-600" data-testid="settings-saved">
              {t("dashboard.settings.saved")}
            </p>
          )}
          {actionData?.ok === false && (
            <p className="text-xs text-amber-700" data-testid="settings-partial-failure">
              {t("dashboard.settings.partialFailure", {
                fields: actionData.failedFields.join(", "),
              })}
            </p>
          )}
        </Form>
      </section>

    </div>
  );
}
