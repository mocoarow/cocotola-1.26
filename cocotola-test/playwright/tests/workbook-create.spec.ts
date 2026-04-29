import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticatePassword, createUser } from "./helpers/auth";
import { createWorkbook, listWorkbooks, waitForSpaceByType } from "./helpers/workbook";

test.describe("workbook: create in private space", () => {
  test("user creates workbook and finds it in listing", async ({ request }) => {
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });

    const suffix = randomUUID().slice(0, 8);
    const userLoginId = `wb-${suffix}@example.com`;

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

    const privateSpaceId = await waitForSpaceByType(request, userToken, "private");

    const created = await createWorkbook(request, {
      userToken,
      spaceId: privateSpaceId,
      title: `テスト用ワークブック ${suffix}`,
      visibility: "private",
      language: "ja",
    });

    const workbooks = await listWorkbooks(request, userToken, privateSpaceId);
    expect(workbooks.length).toBeGreaterThan(0);
    expect(workbooks.some((w) => w.workbookId === created.workbookId)).toBe(true);
  });
});
