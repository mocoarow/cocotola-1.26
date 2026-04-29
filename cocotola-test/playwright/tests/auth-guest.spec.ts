import { expect, test } from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticateGuest, bearer } from "./helpers/auth";

test.describe("auth: guest", () => {
  test("guest can authenticate and fetch own profile", async ({ request }) => {
    const guestToken = await authenticateGuest(request, {
      organizationName: testEnv.organizationName,
    });

    const profile = await request.get("/api/v1/auth/me", {
      headers: bearer(guestToken),
    });
    expect(profile.status()).toBe(200);
    const body = (await profile.json()) as { loginId: string; organizationName: string };
    expect(body.loginId).toBe(`guest@@${testEnv.organizationName}`);
    expect(body.organizationName).toBe(testEnv.organizationName);
  });
});
