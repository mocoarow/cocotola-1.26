import { expect, test } from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticatePassword, bearer } from "./helpers/auth";

test.describe("auth: owner", () => {
  test("owner can authenticate and fetch own profile", async ({ request }) => {
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });

    const profile = await request.get("/api/v1/auth/me", {
      headers: bearer(ownerToken),
    });
    expect(profile.status()).toBe(200);
    const body = (await profile.json()) as { loginId: string; organizationName: string };
    expect(body.loginId).toBe(testEnv.ownerLoginId);
    expect(body.organizationName).toBe(testEnv.organizationName);
  });
});
