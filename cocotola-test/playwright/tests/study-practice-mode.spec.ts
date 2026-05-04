import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticatePassword, createUser, updateUserLanguage } from "./helpers/auth";
import { buildSessionCookieValue, sessionCookieName } from "./helpers/session-cookie";
import {
  addQuestion,
  createWorkbook,
  getStudyQuestions,
  recordAnswerForWordFill,
  waitForSpaceByType,
} from "./helpers/workbook";

const WEB_BASE_URL = process.env.WEB_BASE_URL ?? "http://localhost:5173";

test.describe("study: practice mode", () => {
  test("API: practice mode returns not-yet-due questions and answering in practice mode does not change SRS state", async ({
    request,
  }) => {
    // given: a fresh user with a private workbook containing one word_fill question
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const suffix = randomUUID().slice(0, 8);
    const userLoginId = `prac-${suffix}@example.com`;
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
      title: `practice-${suffix}`,
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

    // when: the user answers the question correctly so its NextDueAt jumps to tomorrow
    await recordAnswerForWordFill(request, userToken, wb.workbookId, questionId, true);

    // then: a normal study fetch returns no questions (the day's queue is exhausted)
    const normal = await getStudyQuestions(request, userToken, wb.workbookId, 20);
    expect(normal.questions).toHaveLength(0);
    expect(normal.totalDue).toBe(0);

    // and: a practice-mode fetch surfaces the same question regardless of NextDueAt
    const practice = await getStudyQuestions(request, userToken, wb.workbookId, 20, {
      practice: true,
    });
    expect(practice.questions).toHaveLength(1);
    expect(practice.questions[0]?.questionId).toBe(questionId);
    expect(practice.totalDue).toBe(1);
  });

  test("API: practice mode also surfaces questions whose record was just rescheduled by an incorrect answer", async ({
    request,
  }) => {
    // given: a fresh user with a private workbook containing one word_fill
    // question. Incorrect answers reschedule NextDueAt to ~5 minutes later,
    // so a normal fetch within those five minutes returns 0 — practice mode
    // must still return the question because the user has not yet mastered it.
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const suffix = randomUUID().slice(0, 8);
    const userLoginId = `prac-wrong-${suffix}@example.com`;
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
      title: `practice-wrong-${suffix}`,
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

    // when: the user records an incorrect answer, pushing NextDueAt ~5 min out
    await recordAnswerForWordFill(request, userToken, wb.workbookId, questionId, false);

    // then: a normal fetch returns nothing because the record's NextDueAt is
    // still in the future
    const normal = await getStudyQuestions(request, userToken, wb.workbookId, 20);
    expect(normal.questions).toHaveLength(0);

    // and: practice mode returns the not-yet-mastered question regardless
    const practice = await getStudyQuestions(request, userToken, wb.workbookId, 20, {
      practice: true,
    });
    expect(practice.questions).toHaveLength(1);
    expect(practice.questions[0]?.questionId).toBe(questionId);
    expect(practice.totalDue).toBe(1);
  });

  test("UI: noQuestions empty state offers a Continue-practicing button that loads the same questions in practice mode without recording answers", async ({
    browser,
    request,
  }) => {
    // given: a fresh user with a private workbook of one word_fill question
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const suffix = randomUUID().slice(0, 8);
    const userLoginId = `prac-ui-${suffix}@example.com`;
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
    await updateUserLanguage(request, userToken, "en");
    const privateSpaceId = await waitForSpaceByType(request, userToken, "private");
    const wb = await createWorkbook(request, {
      userToken,
      spaceId: privateSpaceId,
      title: `practice-ui-${suffix}`,
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

    // and: the user has already answered every question correctly today, so the
    // normal SRS queue is empty.
    await recordAnswerForWordFill(request, userToken, wb.workbookId, questionId, true);
    const before = await getStudyQuestions(request, userToken, wb.workbookId, 20);
    expect(before.questions).toHaveLength(0);

    const cookieValue = await buildSessionCookieValue({
      accessToken: userToken,
      refreshToken: userToken,
    });
    const url = new URL(WEB_BASE_URL);
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

    try {
      // when: the user lands on the study page with the day's queue empty
      await page.goto(`${WEB_BASE_URL}/workbooks/${wb.workbookId}/study`);

      // then: the empty-state message and the Continue-practicing CTA are shown
      await expect(page.getByText("No questions due for study right now.")).toBeVisible();
      const cta = page.getByRole("link", { name: "Continue practicing (no progress saved)" });
      await expect(cta).toBeVisible();

      // when: the user clicks the CTA
      await cta.click();

      // then: the URL flips to practice mode and the practice banner is shown
      await expect(page).toHaveURL(
        new RegExp(`/workbooks/${wb.workbookId}/study\\?practice=true$`),
      );
      await expect(
        page.getByText("Practice mode — your progress is not saved."),
      ).toBeVisible();
      await expect(page.getByRole("heading", { name: "Practice Session" })).toBeVisible();

      // and: the previously-answered question is presented again
      await expect(page.getByLabel("Blank 1")).toBeVisible();

      // when: the user answers correctly in practice mode and clicks Next
      await page.getByLabel("Blank 1").fill("hello");
      await expect(page.getByText("Correct!")).toBeVisible();
      await page.getByRole("button", { name: "Next" }).click();
      await expect(page.getByRole("heading", { name: "Session Complete!" })).toBeVisible();

      // then: the SRS state is unchanged — the next regular fetch is still empty
      const after = await getStudyQuestions(request, userToken, wb.workbookId, 20);
      expect(after.questions).toHaveLength(0);
    } finally {
      await context.close();
    }
  });

  test("UI: clicking Continue-practicing remounts the study session so questions render instead of an immediate Session Complete", async ({
    browser,
    request,
  }) => {
    // Regression: StudySession seeds its `queue` from `questions` via
    // useState which only runs on mount. When the user navigates from
    // /study (questions=[], queue=[]) to /study?practice=true via the
    // Continue-practicing CTA without remounting, the loader returns the
    // ten practice-mode questions but the stale `queue=[]` triggers the
    // Done phase, showing "Session Complete! 0%" until a full reload.
    // The fix is a `key` based on the practice flag forcing a remount.

    // given: a fresh user studying a public workbook so the study auth path
    // does not hit the private-space ACL — orthogonal to this regression.
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });
    const suffix = randomUUID().slice(0, 8);
    const userLoginId = `prac-cta-${suffix}@example.com`;
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
    await updateUserLanguage(request, userToken, "ja");

    const publicResp = await request.get("/api/v1/workbook/public", {
      headers: { authorization: `Bearer ${userToken}` },
    });
    expect(publicResp.status()).toBe(200);
    const publicBody = (await publicResp.json()) as {
      workbooks: { workbookId: string }[];
    };
    expect(publicBody.workbooks.length).toBeGreaterThan(0);
    const wb = publicBody.workbooks[0];
    if (!wb) throw new Error("expected at least one seeded public workbook");

    // and: every question already answered correctly so the normal-mode
    // queue is empty (NextDueAt jumps to tomorrow for correct answers)
    const sq = await request.get(
      `/api/v1/workbook/${wb.workbookId}/study?limit=20`,
      { headers: { authorization: `Bearer ${userToken}` } },
    );
    const sqBody = (await sq.json()) as {
      questions: { questionId: string; questionType: string; content: string }[];
    };
    expect(sqBody.questions.length).toBeGreaterThan(0);
    for (const q of sqBody.questions) {
      const data: Record<string, unknown> =
        q.questionType === "word_fill"
          ? { correct: true }
          : (() => {
              const c = JSON.parse(q.content) as {
                choices: { id: string; isCorrect: boolean }[];
              };
              return {
                selectedChoiceIds: c.choices.filter((x) => x.isCorrect).map((x) => x.id),
              };
            })();
      const r = await request.post(
        `/api/v1/workbook/${wb.workbookId}/study/${q.questionId}/answer`,
        { headers: { authorization: `Bearer ${userToken}` }, data },
      );
      expect(r.status()).toBe(200);
    }

    const cookieValue = await buildSessionCookieValue({
      accessToken: userToken,
      refreshToken: userToken,
    });
    const url = new URL(WEB_BASE_URL);
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
      {
        name: "i18nextLng",
        value: "ja",
        domain: url.hostname,
        path: "/",
        sameSite: "Lax",
        secure: false,
      },
    ]);
    const page = await context.newPage();

    try {
      // when: the user opens the study page in normal mode (queue empty)
      await page.goto(`${WEB_BASE_URL}/workbooks/${wb.workbookId}/study`);
      await expect(page.getByText("現在、学習する問題はありません。")).toBeVisible();
      const cta = page.locator('a:has-text("練習を続ける")');
      await expect(cta).toBeVisible();

      // and: the user clicks Continue practicing — a client-side navigation
      await cta.first().click();
      await page.waitForURL(/practice=true/);

      // then: the practice screen renders with the question card, not the
      // "Session Complete! 0%" result page
      await expect(page.getByRole("heading", { name: "練習セッション" })).toBeVisible();
      await expect(
        page.getByRole("heading", { name: "セッション完了！" }),
      ).not.toBeVisible();
      // The first practice question is a real card, not the empty state
      await expect(page.getByText("0 / ", { exact: false })).toBeVisible();
    } finally {
      await context.close();
    }
  });
});
