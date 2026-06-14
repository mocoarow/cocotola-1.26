import { randomUUID } from "node:crypto";
import { type APIRequestContext, type Browser, type Page, expect, test } from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticatePassword, bearer, createUser, updateUserLanguage } from "./helpers/auth";
import { buildSessionCookieValue, sessionCookieName } from "./helpers/session-cookie";
import { getStudyQuestions, listPublicWorkbooks } from "./helpers/workbook";

const WEB_BASE_URL = process.env.WEB_BASE_URL ?? "http://localhost:5173";

// Matches the loader's default preference (cocotola-web getUserPreferences default).
// Sent on record-answer calls so the seeded activity lands in the same bucket
// the dashboard loader will read for "today".
const DEFAULT_TIMEZONE = "Asia/Tokyo";

function todayInTimezone(timezone: string): string {
  // "en-CA" formats as YYYY-MM-DD natively, which is exactly the wire format
  // the dashboard and record-answer endpoints accept (cocotola-question
  // validates with IsValidDateKey).
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}

type AuthenticatedUser = {
  loginId: string;
  token: string;
};

async function provisionUser(
  request: APIRequestContext,
  ownerToken: string,
  loginIdPrefix: string,
): Promise<AuthenticatedUser> {
  const suffix = randomUUID().slice(0, 8);
  const loginId = `${loginIdPrefix}-${suffix}@example.com`;
  await createUser(request, {
    ownerToken,
    loginId,
    password: testEnv.newUserPassword,
  });
  const token = await authenticatePassword(request, {
    loginId,
    password: testEnv.newUserPassword,
    organizationName: testEnv.organizationName,
  });
  return { loginId, token };
}

async function openAuthenticatedDashboard(
  browser: Browser,
  accessToken: string,
): Promise<{ page: Page; close: () => Promise<void> }> {
  const cookieValue = await buildSessionCookieValue({
    accessToken,
    refreshToken: accessToken,
  });
  const url = new URL(WEB_BASE_URL);
  // Drop the JSON Content-Type that playwright.config.ts sets globally for
  // API tests so the UI navigation runs with browser-default headers
  // (matching what a real user agent would send).
  const context = await browser.newContext({ extraHTTPHeaders: {} });
  await context.addCookies([
    {
      name: sessionCookieName,
      value: cookieValue,
      domain: url.hostname,
      path: "/",
      httpOnly: true,
      sameSite: "Lax",
      secure: false,
    },
    // Pin locale to English so role/label selectors (e.g. the Save button)
    // stay stable even though our primary assertions are data-testid based.
    {
      name: "i18nextLng",
      value: "en",
      domain: url.hostname,
      path: "/",
      sameSite: "Lax",
      secure: false,
    },
  ]);
  const page = await context.newPage();
  await page.goto(`${WEB_BASE_URL}/dashboard`);
  return { page, close: () => context.close() };
}

test.describe("dashboard (UI)", () => {
  test("redirects to /login when no session cookie is set", async ({ browser }) => {
    // given: a fresh browser context with no session cookie
    const context = await browser.newContext({ extraHTTPHeaders: {} });
    const page = await context.newPage();

    // when: navigating directly to /dashboard
    await page.goto(`${WEB_BASE_URL}/dashboard`);

    // then: requireAuth redirects to /login
    await expect(page).toHaveURL(/\/login(\/|$|\?)/);

    await context.close();
  });

  test("renders all dashboard cards for a fresh user with no activity", async ({
    browser,
    request,
  }) => {
    // given: a freshly provisioned user authenticated and cookie-injected
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const user = await provisionUser(request, ownerToken, "dashboard-fresh");

    // when: opening /dashboard
    const { page, close } = await openAuthenticatedDashboard(browser, user.token);

    // then: the page header and all three cards render
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    await expect(page.getByTestId("streak-card")).toBeVisible();
    await expect(page.getByTestId("streak-longest")).toBeVisible();
    await expect(page.getByTestId("daily-goal-card")).toBeVisible();
    await expect(page.getByTestId("daily-goal-progress")).toHaveAttribute("data-reached", "false");
    // The dashboard loader fetches a full 365-day window so the contribution
    // graph itself renders even when every bucket is empty. The empty-state
    // variant only fires when the API returns zero days, which never happens
    // for an authenticated user.
    await expect(page.getByTestId("contribution-graph")).toBeVisible();

    await close();
  });

  test("reflects today's activity after recording an answer", async ({ browser, request }) => {
    // given: a fresh user and one correctly-answered question from a public workbook
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const user = await provisionUser(request, ownerToken, "dashboard-activity");
    // The seeded public workbooks are tagged language="ja" (cocotola-init/seed/seeds/
    // public_workbooks.yaml). The list endpoint filters by the requester's
    // language preference, so a default-language ("en") user sees an empty list.
    // Mirror the same `ja` flip used in study-public-workbook-ui.spec.ts.
    await updateUserLanguage(request, user.token, "ja");

    const publicWorkbooks = await listPublicWorkbooks(request, user.token);
    expect(
      publicWorkbooks.length,
      "expected at least one seeded public workbook (run cocotola-init)",
    ).toBeGreaterThan(0);
    const target = publicWorkbooks[0];
    if (!target) throw new Error("no public workbook available");

    const study = await getStudyQuestions(request, user.token, target.workbookId, 1);
    expect(study.questions.length).toBeGreaterThan(0);
    const question = study.questions[0];
    if (!question) throw new Error("no study question available");

    // Build the answer body matching the question type. The dashboard cares
    // only that an answer was recorded today; correctness does not affect
    // todayCount, so we send a minimally valid payload for either type.
    const answerData: Record<string, unknown> =
      question.questionType === "multiple_choice"
        ? { selectedChoiceIds: [] }
        : { correct: true };

    const today = todayInTimezone(DEFAULT_TIMEZONE);
    // Send X-Local-Date / X-Local-Timezone so the daily-stats increment runs.
    // Without these headers the answer is recorded but the dashboard bucket
    // is not updated (RecordAnswerInput contract in cocotola-question/service).
    const recordResponse = await request.post(
      `/api/v1/workbook/${encodeURIComponent(target.workbookId)}/study/${encodeURIComponent(question.questionId)}/answer`,
      {
        headers: {
          ...bearer(user.token),
          "X-Local-Date": today,
          "X-Local-Timezone": DEFAULT_TIMEZONE,
        },
        data: answerData,
      },
    );
    expect(recordResponse.status()).toBe(200);

    // when: opening /dashboard
    const { page, close } = await openAuthenticatedDashboard(browser, user.token);

    // then: today's contribution cell shows at least one answered question
    const todayCell = page.locator(
      `[data-testid="contribution-cell"][data-date="${today}"]`,
    );
    await expect(todayCell).toBeVisible();
    const countAttr = await todayCell.getAttribute("data-count");
    expect(countAttr).not.toBeNull();
    expect(Number(countAttr)).toBeGreaterThanOrEqual(1);

    // and: the daily-goal progress moved off zero (data-reached stays false
    // because one answer cannot satisfy the default goal of 10).
    const progress = page.getByTestId("daily-goal-progress");
    await expect(progress).toBeVisible();
    await expect(progress).toHaveAttribute("data-reached", "false");
    // aria-valuenow is the percentage (0-100). One answer against the default
    // goal of 10 yields 10, so any leading non-zero digit proves the bar moved.
    const progressBar = page.getByTestId("daily-goal-card").getByRole("progressbar");
    await expect(progressBar).toHaveAttribute("aria-valuenow", /^[1-9]\d*$/);

    await close();
  });

  test("does_not_render_preferences_form_after_settings_page_moved", async ({
    browser,
    request,
  }) => {
    // given: a fresh user on /dashboard
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const user = await provisionUser(request, ownerToken, "dashboard-no-prefs");
    const { page, close } = await openAuthenticatedDashboard(browser, user.token);

    // then: the daily-goal preference input lives on /settings now,
    // not on the dashboard. A regression that re-introduces the form
    // here would catch the user out by saving twice (dashboard + settings).
    await expect(page.locator("#dailyGoal")).toHaveCount(0);
    await expect(page.locator("#timezone")).toHaveCount(0);

    await close();
  });
});
