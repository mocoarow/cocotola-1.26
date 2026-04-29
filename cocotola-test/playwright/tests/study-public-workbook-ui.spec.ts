import { randomUUID } from "node:crypto";
import { type Page, expect, test } from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticatePassword, createUser, updateUserLanguage } from "./helpers/auth";
import { buildSessionCookieValue, sessionCookieName } from "./helpers/session-cookie";
import {
  type StudyQuestion,
  getStudyQuestions,
  listPublicWorkbooks,
} from "./helpers/workbook";

const WEB_BASE_URL = process.env.WEB_BASE_URL ?? "http://localhost:5173";
const STUDY_LIMIT = 20;

type Choice = { id: string; text: string; isCorrect: boolean };
type MultipleChoiceContent = { questionText: string; choices: Choice[] };
type WordFillContent = { source?: { text: string }; target?: { text: string } };

type AnswerKey =
  | { type: "multiple_choice"; questionText: string; correctTexts: string[] }
  | { type: "word_fill"; sourceText: string; targetText: string; answers: string[] };

function buildAnswerKey(question: StudyQuestion): AnswerKey {
  if (question.questionType === "multiple_choice") {
    const parsed = JSON.parse(question.content) as MultipleChoiceContent;
    if (!Array.isArray(parsed.choices) || parsed.choices.length === 0) {
      throw new Error(`question ${question.questionId} has no choices`);
    }
    const correctTexts = parsed.choices.filter((c) => c.isCorrect).map((c) => c.text);
    if (correctTexts.length === 0) {
      throw new Error(`question ${question.questionId} has no correct choice`);
    }
    return {
      type: "multiple_choice",
      questionText: parsed.questionText,
      correctTexts,
    };
  }
  if (question.questionType === "word_fill") {
    const parsed = JSON.parse(question.content) as WordFillContent;
    const targetText = parsed.target?.text;
    if (!targetText) {
      throw new Error(`question ${question.questionId} has no target.text`);
    }
    const matches = targetText.match(/\{\{([^}]+)\}\}/g);
    if (!matches || matches.length === 0) {
      throw new Error(`question ${question.questionId} has no blanks in target.text`);
    }
    return {
      type: "word_fill",
      sourceText: parsed.source?.text ?? "",
      targetText,
      answers: matches.map((m) => m.slice(2, -2)),
    };
  }
  throw new Error(`unsupported question type: ${question.questionType}`);
}

async function readStablePrompt(page: Page): Promise<string> {
  // The action submit triggers a loader revalidation that re-renders the
  // current question card. Read the prompt twice with a short delay until two
  // consecutive samples agree, so subsequent locators target the settled DOM.
  const promptLocator = page.locator("p.text-lg.font-medium").first();
  let prev = "";
  for (let i = 0; i < 30; i++) {
    await page.waitForLoadState("networkidle");
    await expect(promptLocator).toBeVisible();
    const cur = (await promptLocator.textContent())?.trim() ?? "";
    if (cur && cur === prev) return cur;
    prev = cur;
    await page.waitForTimeout(100);
  }
  throw new Error("question prompt did not stabilize");
}

async function answerCurrentQuestion(page: Page, keys: AnswerKey[]): Promise<void> {
  const promptText = await readStablePrompt(page);
  const mc = keys.find((k) => k.type === "multiple_choice" && k.questionText === promptText);
  if (mc && mc.type === "multiple_choice") {
    for (const text of mc.correctTexts) {
      await page.getByRole("button", { name: text, exact: true }).click();
    }
    await page.getByRole("button", { name: "Check" }).click();
    await page.getByRole("button", { name: "Next" }).click();
    return;
  }
  const wf = keys.find((k) => k.type === "word_fill" && k.sourceText === promptText);
  if (wf && wf.type === "word_fill") {
    for (const [i, answer] of wf.answers.entries()) {
      await page.getByLabel(`Blank ${i + 1}`).fill(answer);
    }
    await page.getByRole("button", { name: "Check" }).click();
    await page.getByRole("button", { name: "Next" }).click();
    return;
  }
  throw new Error(`no matching answer key for prompt: ${promptText}`);
}

test.describe("study: public workbook (UI)", () => {
  test("user logs in, opens a public workbook, and answers all questions correctly", async ({
    browser,
    request,
  }) => {
    // given: a fresh user authenticated via password and language set to "ja"
    const ownerToken = await authenticatePassword(request, {
      loginId: testEnv.ownerLoginId,
      password: testEnv.ownerPassword,
      organizationName: testEnv.organizationName,
    });

    const suffix = randomUUID().slice(0, 8);
    const studyUserLoginId = `study-ui-${suffix}@example.com`;
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
    await updateUserLanguage(request, userToken, "ja");

    // and: a seeded public workbook with at least one question
    const publicWorkbooks = await listPublicWorkbooks(request, userToken);
    expect(
      publicWorkbooks.length,
      "expected at least one seeded public workbook (run cocotola-init)",
    ).toBeGreaterThan(0);
    const target = publicWorkbooks[0];
    if (!target) throw new Error("no public workbook available");

    const study = await getStudyQuestions(request, userToken, target.workbookId, STUDY_LIMIT);
    expect(study.questions.length).toBeGreaterThan(0);

    // and: a browser context pre-authenticated by injecting the cocotola session cookie
    // (and forcing the UI locale to English so role/label selectors are stable).
    const cookieValue = await buildSessionCookieValue({
      accessToken: userToken,
      refreshToken: userToken,
    });
    const url = new URL(WEB_BASE_URL);
    // Drop the global Content-Type override (set in playwright.config.ts for
    // API tests) — it breaks form-encoded action submissions made by the UI.
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

    // when: navigating to the public workbooks page and starting a study session
    await page.goto(`${WEB_BASE_URL}/workbooks/public`);
    await expect(page.getByRole("heading", { name: "Public Workbooks" })).toBeVisible();

    const card = page.locator("div.group").filter({ hasText: target.title }).first();
    await card.getByRole("button", { name: "Study", exact: true }).click();
    await expect(page).toHaveURL(new RegExp(`/workbooks/${target.workbookId}/study$`));
    await expect(page.getByRole("heading", { name: "Study Session" })).toBeVisible();

    // and: answering every served question correctly. The SSR loader may return
    // fewer questions than the test pre-fetched (revalidation between tests
    // and the answer-record action), so loop until the result page renders.
    const answerKeys = study.questions.map(buildAnswerKey);
    const resultHeading = page.getByRole("heading", { name: "Session Complete!" });
    const maxIterations = answerKeys.length + 5;
    for (let i = 0; i < maxIterations; i++) {
      if (await resultHeading.isVisible().catch(() => false)) break;
      await answerCurrentQuestion(page, answerKeys);
    }

    // then: the result page reports a perfect score
    await expect(page.getByRole("heading", { name: "Session Complete!" })).toBeVisible();
    await expect(page.getByText(`You scored 100%`)).toBeVisible();
    await page.screenshot({
      path: "test-results/study-public-workbook-ui-result.png",
      fullPage: true,
    });

    await context.close();
  });
});
