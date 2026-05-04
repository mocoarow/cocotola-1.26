import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticatePassword, createUser } from "./helpers/auth";
import {
  addQuestion,
  createWorkbook,
  recordAnswerForWordFill,
  waitForSpaceByType,
} from "./helpers/workbook";

const FIVE_MINUTES_MS = 5 * 60 * 1000;
// Allow a generous wall-clock skew between the request being sent and the
// server timestamp it persists. CI runners and the local dev box are routinely
// off by a few seconds; the test is asserting the SRS scheduling rule
// (~5 minutes), not clock-precision parity.
const TOLERANCE_MS = 30 * 1000;

test.describe("study: incorrect-answer retry scheduling", () => {
  test("a wrong word_fill answer is rescheduled ~5 minutes later, not a full day later", async ({
    request,
  }) => {
    // given: a fresh user with a private workbook and a single word_fill question
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const suffix = randomUUID().slice(0, 8);
    const userLoginId = `retry-${suffix}@example.com`;
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
    const wb = await createWorkbook(request, {
      userToken,
      spaceId: privateSpaceId,
      title: `retry-${suffix}`,
      visibility: "private",
      language: "en",
    });
    const questionId = await addQuestion(request, {
      userToken,
      workbookId: wb.workbookId,
      questionType: "word_fill",
      content: JSON.stringify({
        source: { text: "ハロー", lang: "ja" },
        target: { text: "{{hello}}", lang: "en" },
      }),
      orderIndex: 0,
    });

    // when: the user records an incorrect answer
    const before = Date.now();
    const result = await recordAnswerForWordFill(
      request,
      userToken,
      wb.workbookId,
      questionId,
      false,
    );
    const after = Date.now();

    // then: the nextDueAt sits roughly 5 minutes after `now`, not a full day
    const nextDueAtMs = Date.parse(result.nextDueAt);
    const expectedMin = before + FIVE_MINUTES_MS - TOLERANCE_MS;
    const expectedMax = after + FIVE_MINUTES_MS + TOLERANCE_MS;
    expect(nextDueAtMs).toBeGreaterThanOrEqual(expectedMin);
    expect(nextDueAtMs).toBeLessThanOrEqual(expectedMax);

    // and: the per-record counters reflect a single incorrect attempt with no
    // accumulated streak
    expect(result.consecutiveCorrect).toBe(0);
    expect(result.totalCorrect).toBe(0);
    expect(result.totalIncorrect).toBe(1);
  });
});
