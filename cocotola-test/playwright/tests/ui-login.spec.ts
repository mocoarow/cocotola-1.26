import { expect, test } from "@playwright/test";
import { authenticatePassword } from "./helpers/auth";
import { buildSessionCookieValue, sessionCookieName } from "./helpers/session-cookie";
import { testEnv } from "./fixtures/env";

const WEB_BASE_URL = process.env.WEB_BASE_URL ?? "http://localhost:5173";

test.describe("UI login (cookie injection)", () => {
  test("owner can access /workbooks via injected session cookie", async ({ browser, request }) => {
    // given: owner's accessToken from the password login API
    const accessToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const cookieValue = await buildSessionCookieValue({
      accessToken,
      refreshToken: accessToken,
    });
    const url = new URL(WEB_BASE_URL);
    const context = await browser.newContext();
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
    ]);

    // when: visiting /login while authenticated; loader should redirect back to /
    const page = await context.newPage();
    await page.goto(`${WEB_BASE_URL}/login`);

    // then: redirected away from /login (proves the session cookie was accepted)
    await expect(page).not.toHaveURL(/\/login(\/|$)/);
    await page.screenshot({ path: "test-results/ui-login-after-cookie.png", fullPage: true });

    await context.close();
  });
});
