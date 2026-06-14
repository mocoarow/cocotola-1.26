import { randomUUID } from "node:crypto";
import { type APIRequestContext, type Browser, type Page, expect, test } from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticatePassword, createUser } from "./helpers/auth";
import { buildSessionCookieValue, sessionCookieName } from "./helpers/session-cookie";

const WEB_BASE_URL = process.env.WEB_BASE_URL ?? "http://localhost:5173";

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

async function openAuthenticatedSettings(
  browser: Browser,
  accessToken: string,
): Promise<{ page: Page; close: () => Promise<void> }> {
  const cookieValue = await buildSessionCookieValue({
    accessToken,
    refreshToken: accessToken,
  });
  const url = new URL(WEB_BASE_URL);
  // Drop the JSON Content-Type that playwright.config.ts sets globally for
  // API tests — the settings page posts as application/x-www-form-urlencoded.
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
    // Pin locale to English so role/label selectors stay stable.
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
  await page.goto(`${WEB_BASE_URL}/settings`);
  return { page, close: () => context.close() };
}

test.describe("settings (UI)", () => {
  test("redirects to /login when no session cookie is set", async ({ browser }) => {
    // given: a fresh browser context with no session cookie
    const context = await browser.newContext({ extraHTTPHeaders: {} });
    const page = await context.newPage();

    // when: navigating directly to /settings
    await page.goto(`${WEB_BASE_URL}/settings`);

    // then: requireAuth redirects to /login
    await expect(page).toHaveURL(/\/login(\/|$|\?)/);

    await context.close();
  });

  test("renders account info, language selector, and study preferences", async ({
    browser,
    request,
  }) => {
    // given: a freshly provisioned user authenticated and cookie-injected
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const user = await provisionUser(request, ownerToken, "settings-render");

    // when: opening /settings
    const { page, close } = await openAuthenticatedSettings(browser, user.token);

    // then: all three sections render
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    await expect(page.getByTestId("settings-login-id")).toBeVisible();
    await expect(page.getByTestId("settings-organization")).toBeVisible();
    await expect(page.getByTestId("settings-language-select")).toBeVisible();
    await expect(page.locator("#dailyGoal")).toBeVisible();
    await expect(page.locator("#timezone")).toBeVisible();

    await close();
  });

  test("saves the daily goal preference and persists across reload", async ({
    browser,
    request,
  }) => {
    // given: a fresh user on /settings with the form pre-filled by the loader
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const user = await provisionUser(request, ownerToken, "settings-prefs");
    const { page, close } = await openAuthenticatedSettings(browser, user.token);

    const dailyGoalInput = page.locator("#dailyGoal");
    await expect(dailyGoalInput).toBeVisible();

    // when: changing the daily goal and submitting the study form
    await dailyGoalInput.fill("25");
    // The form has a single submit button; identifying by type is locale-safe.
    // Scope to the form section to avoid colliding with any header buttons.
    const studySection = page.locator('section:has(#dailyGoal)');
    await studySection.locator('button[type="submit"]').click();

    // then: the saved confirmation appears
    await expect(page.getByTestId("settings-saved")).toBeVisible();

    // and: the change survives a full reload (loader rehydrates from /auth/me)
    await page.reload();
    await expect(page.locator("#dailyGoal")).toHaveValue("25");

    await close();
  });
});
