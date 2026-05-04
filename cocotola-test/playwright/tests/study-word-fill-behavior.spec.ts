import { randomUUID } from "node:crypto";
import {
  type APIRequestContext,
  type Browser,
  type Page,
  expect,
  test,
} from "@playwright/test";
import { testEnv } from "./fixtures/env";
import { authenticatePassword, createUser, updateUserLanguage } from "./helpers/auth";
import { buildSessionCookieValue, sessionCookieName } from "./helpers/session-cookie";
import { addQuestion, createWorkbook, waitForSpaceByType } from "./helpers/workbook";

const WEB_BASE_URL = process.env.WEB_BASE_URL ?? "http://localhost:5173";

type StudyContext = {
  page: Page;
  workbookId: string;
  cleanup: () => Promise<void>;
};

type QuestionSpec =
  | { type: "word_fill"; sourceText: string; targetText: string }
  | {
      type: "multiple_choice";
      questionText: string;
      choices: { id: string; text: string; isCorrect: boolean }[];
      displayCount?: number;
    };

function buildQuestionContent(spec: QuestionSpec): string {
  if (spec.type === "word_fill") {
    return JSON.stringify({
      source: { text: spec.sourceText, lang: "ja" },
      target: { text: spec.targetText, lang: "en" },
    });
  }
  return JSON.stringify({
    questionText: spec.questionText,
    choices: spec.choices,
    displayCount: spec.displayCount ?? spec.choices.length,
    showCorrectCount: false,
    shuffleChoices: false,
    allowPartialCredit: false,
  });
}

async function setupStudy(
  request: APIRequestContext,
  browser: Browser,
  questions: QuestionSpec[],
): Promise<StudyContext> {
  const ownerToken = await authenticatePassword(request, {
    loginId: testEnv.ownerLoginId,
    password: testEnv.ownerPassword,
    organizationName: testEnv.organizationName,
  });

  const suffix = randomUUID().slice(0, 8);
  const studyUserLoginId = `wf-${suffix}@example.com`;
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
  await updateUserLanguage(request, userToken, "en");

  const privateSpaceId = await waitForSpaceByType(request, userToken, "private");

  const workbook = await createWorkbook(request, {
    userToken,
    spaceId: privateSpaceId,
    title: `wf-test-${suffix}`,
    visibility: "private",
    language: "en",
  });

  for (const [index, spec] of questions.entries()) {
    await addQuestion(request, {
      userToken,
      workbookId: workbook.workbookId,
      questionType: spec.type,
      content: buildQuestionContent(spec),
      orderIndex: index,
    });
  }

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

  await page.goto(`${WEB_BASE_URL}/workbooks/${workbook.workbookId}/study`);
  await expect(page.getByRole("heading", { name: "Study Session" })).toBeVisible();

  return {
    page,
    workbookId: workbook.workbookId,
    cleanup: () => context.close(),
  };
}

async function setupWordFillStudy(
  request: APIRequestContext,
  browser: Browser,
  targetText: string,
  sourceText = "テスト用ソース",
): Promise<StudyContext> {
  const ctx = await setupStudy(request, browser, [
    { type: "word_fill", sourceText, targetText },
  ]);
  // Wait for the first blank to render before any test-specific assertions.
  await expect(ctx.page.getByLabel("Blank 1")).toBeVisible();
  return ctx;
}

test.describe("study: word_fill UI behavior", () => {
  test("focuses the first blank when the question card mounts", async ({ browser, request }) => {
    // given: a study session for a word_fill question with two blanks
    const { page, cleanup } = await setupWordFillStudy(
      request,
      browser,
      "{{hello}} {{world}}",
    );

    try {
      // when: the page has rendered the question card
      // then: the first blank holds focus, the second one does not
      await expect(page.getByLabel("Blank 1")).toBeFocused();
      await expect(page.getByLabel("Blank 2")).not.toBeFocused();
    } finally {
      await cleanup();
    }
  });

  test("moves focus to the next blank when the current answer is correct", async ({
    browser,
    request,
  }) => {
    // given: a study session for a word_fill question with two blanks
    const { page, cleanup } = await setupWordFillStudy(
      request,
      browser,
      "{{hello}} {{world}}",
    );

    try {
      // when: the user types the correct answer in the first blank
      await page.getByLabel("Blank 1").fill("hello");

      // then: focus moves to the second blank
      await expect(page.getByLabel("Blank 2")).toBeFocused();
    } finally {
      await cleanup();
    }
  });

  test("locks a blank as read-only once the correct answer is entered", async ({
    browser,
    request,
  }) => {
    // given: a study session for a word_fill question with two blanks
    const { page, cleanup } = await setupWordFillStudy(
      request,
      browser,
      "{{hello}} {{world}}",
    );

    try {
      // when: the user types the correct answer in the first blank
      await page.getByLabel("Blank 1").fill("hello");

      // then: the first blank is locked — read-only, disabled, and still
      // showing the entered value
      const firstBlank = page.getByLabel("Blank 1");
      await expect(firstBlank).toHaveValue("hello");
      await expect(firstBlank).toHaveAttribute("readonly", "");
      await expect(firstBlank).toBeDisabled();
    } finally {
      await cleanup();
    }
  });

  test("skips locked blanks when wrapping focus from the last blank", async ({
    browser,
    request,
  }) => {
    // given: three blanks where Blank 1 is already locked
    const { page, cleanup } = await setupWordFillStudy(
      request,
      browser,
      "{{first}} {{second}} {{third}}",
    );

    try {
      await page.getByLabel("Blank 1").fill("first");
      await expect(page.getByLabel("Blank 1")).toBeDisabled();

      // when: the user clicks into Blank 3 (skipping Blank 2) and types the
      // correct answer
      await page.getByLabel("Blank 3").click();
      await page.getByLabel("Blank 3").fill("third");

      // then: focus wraps past the locked Blank 1 and lands on the still-empty
      // Blank 2
      await expect(page.getByLabel("Blank 2")).toBeFocused();
    } finally {
      await cleanup();
    }
  });

  test("keeps focus on the current blank when the answer is wrong", async ({
    browser,
    request,
  }) => {
    // given: a study session for a word_fill question with two blanks
    const { page, cleanup } = await setupWordFillStudy(
      request,
      browser,
      "{{hello}} {{world}}",
    );

    try {
      // when: the user types a wrong answer in the first blank
      await page.getByLabel("Blank 1").fill("nope");

      // then: focus remains on the first blank
      await expect(page.getByLabel("Blank 1")).toBeFocused();
      await expect(page.getByLabel("Blank 2")).not.toBeFocused();
    } finally {
      await cleanup();
    }
  });

  test("wraps focus to the first blank after the last blank is correctly answered when others remain unfilled", async ({
    browser,
    request,
  }) => {
    // given: a study session for a word_fill question with two blanks
    const { page, cleanup } = await setupWordFillStudy(
      request,
      browser,
      "{{hello}} {{world}}",
    );

    try {
      // and: the user clicks into the last blank first, leaving the first blank empty
      await page.getByLabel("Blank 2").click();
      await expect(page.getByLabel("Blank 2")).toBeFocused();

      // when: the user types the correct answer in the last blank
      await page.getByLabel("Blank 2").fill("world");

      // then: focus wraps back to the first (still-unfilled) blank
      await expect(page.getByLabel("Blank 1")).toBeFocused();
    } finally {
      await cleanup();
    }
  });

  test("shows the correct result screen but does not advance until Next is clicked when all blanks are correctly answered", async ({
    browser,
    request,
  }) => {
    // given: a study session containing a single word_fill question
    const { page, cleanup } = await setupWordFillStudy(
      request,
      browser,
      "{{hello}} {{world}}",
    );

    try {
      // when: the user fills both blanks correctly
      await page.getByLabel("Blank 1").fill("hello");
      await page.getByLabel("Blank 2").fill("world");

      // then: the in-card result screen appears with Correct! feedback and a
      // Next button — the session must not auto-advance to the result page.
      await expect(page.getByText("Correct!")).toBeVisible();
      await expect(page.getByRole("button", { name: "Next" })).toBeVisible();
      await expect(
        page.getByRole("heading", { name: "Session Complete!" }),
      ).not.toBeVisible();

      // and: clicking Next advances to the session result page
      await page.getByRole("button", { name: "Next" }).click();
      await expect(page.getByRole("heading", { name: "Session Complete!" })).toBeVisible();
      await expect(page.getByText("You scored 100%")).toBeVisible();
    } finally {
      await cleanup();
    }
  });

  test("renders a Show answer button instead of a Check button while the question is unanswered", async ({
    browser,
    request,
  }) => {
    // given: a study session for a word_fill question
    const { page, cleanup } = await setupWordFillStudy(request, browser, "{{hello}}");

    try {
      // when: the question card is on screen and untouched
      // then: the Show answer button is shown and the legacy Check button is not
      await expect(page.getByRole("button", { name: "Show answer" })).toBeVisible();
      await expect(page.getByRole("button", { name: "Check", exact: true })).toHaveCount(0);
    } finally {
      await cleanup();
    }
  });

  test("reveals the correct answer and exposes a Next button after Show answer is clicked", async ({
    browser,
    request,
  }) => {
    // given: a study session for a word_fill question with one blank
    const { page, cleanup } = await setupWordFillStudy(request, browser, "{{hello}}");

    try {
      // when: the user gives up and clicks Show answer without typing anything
      await page.getByRole("button", { name: "Show answer" }).click();

      // then: the correct answer is shown and a Next button appears
      await expect(page.getByText("hello")).toBeVisible();
      await expect(page.getByRole("button", { name: "Next" })).toBeVisible();
    } finally {
      await cleanup();
    }
  });
});

test.describe("study: mixed workbook (word_fill + multiple_choice)", () => {
  test("user studies through a workbook containing both word_fill and multiple_choice questions and finishes with a perfect score", async ({
    browser,
    request,
  }) => {
    // given: a workbook with one word_fill and one multiple_choice question
    const { page, cleanup } = await setupStudy(request, browser, [
      {
        type: "word_fill",
        sourceText: "ミックス用ソース",
        targetText: "{{hello}} {{world}}",
      },
      {
        type: "multiple_choice",
        questionText: "Pick the greeting",
        choices: [
          { id: "c1", text: "Hello", isCorrect: true },
          { id: "c2", text: "Goodbye", isCorrect: false },
          { id: "c3", text: "Maybe", isCorrect: false },
        ],
      },
    ]);

    try {
      // when: a word_fill card appears first; verify the new behavior end-to-end
      await expect(page.getByLabel("Blank 1")).toBeVisible();
      await expect(page.getByLabel("Blank 1")).toBeFocused();

      // and: typing the first correct answer moves focus to the second blank
      await page.getByLabel("Blank 1").fill("hello");
      await expect(page.getByLabel("Blank 2")).toBeFocused();

      // and: completing the second blank shows the in-card correct screen
      await page.getByLabel("Blank 2").fill("world");
      await expect(page.getByText("Correct!")).toBeVisible();
      await expect(page.getByRole("button", { name: "Next" })).toBeVisible();
      // and: the session must not have skipped to the result page
      await expect(
        page.getByRole("heading", { name: "Session Complete!" }),
      ).not.toBeVisible();

      // and: clicking Next advances to the multiple_choice question
      await page.getByRole("button", { name: "Next" }).click();
      await expect(page.getByText("Pick the greeting")).toBeVisible();
      await expect(page.getByRole("button", { name: "Hello" })).toBeVisible();

      // and: multiple_choice still uses the legacy Check → Next flow
      await page.getByRole("button", { name: "Hello", exact: true }).click();
      await page.getByRole("button", { name: "Check" }).click();
      await expect(page.getByText("Correct!")).toBeVisible();
      await page.getByRole("button", { name: "Next" }).click();

      // then: the session result page reports a perfect score
      await expect(page.getByRole("heading", { name: "Session Complete!" })).toBeVisible();
      await expect(page.getByText("You scored 100%")).toBeVisible();
    } finally {
      await cleanup();
    }
  });

  test("word_fill card renders Show answer while multiple_choice card renders Check in the same workbook", async ({
    browser,
    request,
  }) => {
    // given: a workbook with one word_fill followed by one multiple_choice
    const { page, cleanup } = await setupStudy(request, browser, [
      {
        type: "word_fill",
        sourceText: "混在テスト",
        targetText: "{{hello}}",
      },
      {
        type: "multiple_choice",
        questionText: "Pick the greeting",
        choices: [
          { id: "c1", text: "Hello", isCorrect: true },
          { id: "c2", text: "Goodbye", isCorrect: false },
        ],
      },
    ]);

    try {
      // when: the word_fill card is on screen
      await expect(page.getByLabel("Blank 1")).toBeVisible();

      // then: the action button is "Show answer", not "Check"
      await expect(page.getByRole("button", { name: "Show answer" })).toBeVisible();
      await expect(page.getByRole("button", { name: "Check", exact: true })).toHaveCount(0);

      // when: the user gives up and reveals the answer, then advances
      await page.getByRole("button", { name: "Show answer" }).click();
      await page.getByRole("button", { name: "Next" }).click();

      // then: the next card is multiple_choice and uses the legacy Check button
      await expect(page.getByText("Pick the greeting")).toBeVisible();
      await expect(page.getByRole("button", { name: "Check" })).toBeVisible();
      await expect(page.getByRole("button", { name: "Show answer" })).toHaveCount(0);
    } finally {
      await cleanup();
    }
  });
});
