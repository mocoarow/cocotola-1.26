import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticatePassword, bearer, createUser } from "./helpers/auth";

type Space = {
  spaceId: string;
  spaceType: string;
};

type ListSpacesResponse = {
  spaces: Space[];
};

test.describe("user: private space provisioning", () => {
  test("owner creates user, user logs in and finds own private space", async ({ request }) => {
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });

    const suffix = randomUUID().slice(0, 8);
    const userLoginId = `private-${suffix}@example.com`;

    await createUser(request, {
      ownerToken,
      loginId: userLoginId,
      password: testEnv.newUserPassword,
    });

    const userToken = await authenticatePassword(request, {
      loginId: userLoginId,
      password: testEnv.newUserPassword,
      organizationName: testEnv.organizationName,
    });

    const listSpaces = await request.get("/api/v1/auth/space", {
      headers: bearer(userToken),
    });
    expect(listSpaces.status()).toBe(200);
    const body = (await listSpaces.json()) as ListSpacesResponse;
    expect(body.spaces.length).toBeGreaterThan(0);
    const privateSpace = body.spaces.find((s) => s.spaceType === "private");
    expect(privateSpace, "user should have a private space").toBeDefined();
    expect(privateSpace?.spaceId).toBeTruthy();
  });
});
