import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticatePassword, createUser, updateUserLanguage } from "./helpers/auth";
import {
  getStudyQuestions,
  listPublicWorkbooks,
  recordAnswer,
} from "./helpers/workbook";

// Public workbooks are seeded by `cocotola-init` (see
// cocotola-init/seed/seeds/public_workbooks.yaml). Regular users cannot create
// workbooks in the public space — only the SystemOwner (used by the seeder)
// can. So this scenario verifies the realistic flow: a freshly created user
// logs in, browses pre-seeded public workbooks, and studies a question.
test.describe("study: public workbook", () => {
  test("user logs in and studies a question from a public workbook", async ({ request }) => {
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });

    const suffix = randomUUID().slice(0, 8);
    const studyUserLoginId = `study-${suffix}@example.com`;

    await createUser(request, {
      ownerToken,
      loginId: studyUserLoginId,
      password: testEnv.newUserPassword,
    });

    const userToken = await authenticatePassword(request, {
      loginId: studyUserLoginId,
      password: testEnv.newUserPassword,
      organizationName: testEnv.organizationName,
    });

    // Public workbooks are filtered by the caller's preferred language. The
    // seeded fixtures live under "ja" (see public_workbooks.yaml), so switch
    // the user's language before listing.
    await updateUserLanguage(request, userToken, "ja");

    const publicWorkbooks = await listPublicWorkbooks(request, userToken);
    expect(
      publicWorkbooks.length,
      "expected at least one seeded public workbook (run cocotola-init)",
    ).toBeGreaterThan(0);
    const target = publicWorkbooks[0];
    expect(target).toBeDefined();
    if (!target) return;

    const study = await getStudyQuestions(request, userToken, target.workbookId, 10);
    expect(study.totalDue).toBeGreaterThan(0);
    expect(study.questions.length).toBeGreaterThan(0);
    const firstQuestion = study.questions[0];
    expect(firstQuestion).toBeDefined();
    if (!firstQuestion) return;

    const answer = await recordAnswer(
      request,
      userToken,
      target.workbookId,
      firstQuestion.questionId,
      true,
    );
    expect(answer.totalCorrect).toBe(1);
    expect(answer.totalIncorrect).toBe(0);
    expect(answer.consecutiveCorrect).toBe(1);
    expect(new Date(answer.nextDueAt).getTime()).toBeGreaterThan(Date.now());
  });
});
