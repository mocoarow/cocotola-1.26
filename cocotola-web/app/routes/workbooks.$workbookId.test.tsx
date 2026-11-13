import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/question.server", async (importOriginal) => {
  const actual = await importOriginal<typeof import("~/lib/api/question.server")>();
  return {
    listQuestions: vi.fn(),
    updateQuestion: vi.fn(),
    deleteQuestion: vi.fn(),
    addQuestion: vi.fn(),
    QuestionValidationError: actual.QuestionValidationError,
  };
});

vi.mock("~/lib/api/study.server", () => ({
  getStudyQuestions: vi.fn(),
  getStudySummary: vi.fn(),
  recordAnswer: vi.fn(),
  deleteStudyHistory: vi.fn(),
  listStudyRecords: vi.fn(),
}));

vi.mock("~/lib/api/workbook.server", async (importOriginal) => {
  const actual = await importOriginal<typeof import("~/lib/api/workbook.server")>();
  return {
    listWorkbooks: vi.fn(),
    createWorkbook: vi.fn(),
    getWorkbook: vi.fn(),
    updateWorkbook: vi.fn(),
    deleteWorkbook: vi.fn(),
    WorkbookValidationError: actual.WorkbookValidationError,
  };
});

import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRoutesStub, data } from "react-router";
import { ConfirmDialogProvider } from "~/components/confirm-dialog-provider";
import { QUESTION_INVALID_ERROR_KEY } from "~/components/workbook/question-form-error";
import i18n from "~/i18n/config";
import type { Question } from "~/lib/api/question.server";
import { addQuestion, QuestionValidationError, updateQuestion } from "~/lib/api/question.server";
import type { Workbook } from "~/lib/api/workbook.server";
import { updateWorkbook, WorkbookValidationError } from "~/lib/api/workbook.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import WorkbookDetail, { action } from "./workbooks.$workbookId";

type ActionArgs = Parameters<typeof action>[0];

const ORDERING_ENTRIES = {
  sourceText: "私は学生です。",
  sourceLang: "ja",
  targetText: "I / am / a student [[.]]",
  targetLang: "en",
  tags: "level:beginner",
};

const EXPECTED_CONTENT = JSON.stringify({
  source: { text: "私は学生です。", lang: "ja" },
  target: {
    lang: "en",
    segments: [
      { type: "chip", text: "I" },
      { type: "chip", text: "am" },
      { type: "chip", text: "a student" },
      { type: "fixed", text: "." },
    ],
  },
});

const WORD_FILL_ENTRIES = {
  sourceText: "ゴミを捨てる",
  sourceLang: "ja",
  targetText: "{{throw}} it {{away}}",
  targetLang: "en",
};

const MULTIPLE_CHOICE_ENTRIES = {
  questionText: "Which is a fruit?",
  choices: JSON.stringify([
    { id: "1", text: "apple", isCorrect: true },
    { id: "2", text: "chair", isCorrect: false },
  ]),
};

const UPDATE_ENTRIES = { intent: "updateQuestion", questionId: "q-1" };

const UPDATE_TITLE_ENTRIES = {
  intent: "updateTitle",
  title: "Phrasal verbs",
  description: "",
  visibility: "private",
};

type QuestionBranch = {
  name: string;
  entries: Record<string, string>;
  api: typeof addQuestion | typeof updateQuestion;
};

const QUESTION_BRANCHES: QuestionBranch[] = [
  {
    name: "addWordFill",
    entries: { intent: "addWordFill", ...WORD_FILL_ENTRIES },
    api: addQuestion,
  },
  {
    name: "addMultipleChoice",
    entries: { intent: "addMultipleChoice", ...MULTIPLE_CHOICE_ENTRIES },
    api: addQuestion,
  },
  {
    name: "addOrdering",
    entries: { intent: "addOrdering", ...ORDERING_ENTRIES },
    api: addQuestion,
  },
  {
    name: "updateQuestion for word_fill",
    entries: { ...UPDATE_ENTRIES, questionType: "word_fill", ...WORD_FILL_ENTRIES },
    api: updateQuestion,
  },
  {
    name: "updateQuestion for multiple_choice",
    entries: { ...UPDATE_ENTRIES, questionType: "multiple_choice", ...MULTIPLE_CHOICE_ENTRIES },
    api: updateQuestion,
  },
  {
    name: "updateQuestion for ordering",
    entries: { ...UPDATE_ENTRIES, questionType: "ordering", ...ORDERING_ENTRIES },
    api: updateQuestion,
  },
];

function buildActionArgs(formEntries: Record<string, string>): ActionArgs {
  const request = new Request("http://localhost/workbooks/wb-1", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams(formEntries).toString(),
  });
  return { request, params: { workbookId: "wb-1" }, context: {} } as unknown as ActionArgs;
}

describe("WorkbookDetail action", () => {
  beforeEach(() => {
    vi.mocked(requireAuth).mockResolvedValue({
      accessToken: "test-token",
      refreshToken: undefined,
    });
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it("should add an ordering question on addOrdering intent", async () => {
    // given
    const args = buildActionArgs({ intent: "addOrdering", ...ORDERING_ENTRIES });

    // when
    await action(args);

    // then
    expect(addQuestion).toHaveBeenCalledWith("test-token", "wb-1", {
      questionType: "ordering",
      content: EXPECTED_CONTENT,
      tags: ["level:beginner"],
      orderIndex: 0,
    });
  });

  it("should report the addition on addOrdering intent", async () => {
    // given
    const args = buildActionArgs({ intent: "addOrdering", ...ORDERING_ENTRIES });

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: true, added: true });
  });

  it("should reject addOrdering intent with an invalid notation", async () => {
    // given
    const args = buildActionArgs({
      intent: "addOrdering",
      ...ORDERING_ENTRIES,
      targetText: "I [[.]]",
    });

    // when / then
    await expect(action(args)).rejects.toMatchObject({ status: 400 });
  });

  it("should update an ordering question on updateQuestion intent", async () => {
    // given
    const args = buildActionArgs({
      intent: "updateQuestion",
      questionId: "q-1",
      questionType: "ordering",
      orderIndex: "5",
      ...ORDERING_ENTRIES,
    });

    // when
    await action(args);

    // then
    expect(updateQuestion).toHaveBeenCalledWith("test-token", "wb-1", "q-1", {
      content: EXPECTED_CONTENT,
      tags: ["level:beginner"],
      orderIndex: 5,
    });
  });

  it("should split the tags on addWordFill intent", async () => {
    // given
    const args = buildActionArgs({
      intent: "addWordFill",
      sourceText: "ゴミを捨てる",
      sourceLang: "ja",
      targetText: "{{throw}} it {{away}}",
      targetLang: "en",
      tags: " level:beginner , ,topic:phrasal-verbs ",
    });

    // when
    await action(args);

    // then
    expect(vi.mocked(addQuestion).mock.calls[0][2].tags).toEqual([
      "level:beginner",
      "topic:phrasal-verbs",
    ]);
  });

  it("should split the tags on updateQuestion intent for word_fill", async () => {
    // given
    const args = buildActionArgs({
      intent: "updateQuestion",
      questionId: "q-1",
      questionType: "word_fill",
      sourceText: "ゴミを捨てる",
      sourceLang: "ja",
      targetText: "{{throw}} it {{away}}",
      targetLang: "en",
      tags: " level:beginner , ,topic:phrasal-verbs ",
    });

    // when
    await action(args);

    // then
    expect(vi.mocked(updateQuestion).mock.calls[0][3].tags).toEqual([
      "level:beginner",
      "topic:phrasal-verbs",
    ]);
  });

  it("should reject updateQuestion intent for ordering with an invalid notation", async () => {
    // given
    const args = buildActionArgs({
      intent: "updateQuestion",
      questionId: "q-1",
      questionType: "ordering",
      ...ORDERING_ENTRIES,
      targetText: "I / / am",
    });

    // when / then
    await expect(action(args)).rejects.toMatchObject({ status: 400 });
  });

  it.each(
    QUESTION_BRANCHES,
  )("should return the invalid error key with status 400 when $name is rejected as invalid", async ({
    entries,
    api,
  }) => {
    // given
    vi.mocked(api).mockRejectedValue(new QuestionValidationError());
    const args = buildActionArgs(entries);

    // when
    const result = await action(args);

    // then
    expect(result).toMatchObject({
      data: { ok: false, errorKey: QUESTION_INVALID_ERROR_KEY },
      init: { status: 400 },
    });
  });

  it.each(
    QUESTION_BRANCHES,
  )("should rethrow the error when $name fails for another reason", async ({ entries, api }) => {
    // given
    const failure = new Response("Failed to save question", { status: 500 });
    vi.mocked(api).mockRejectedValue(failure);
    const args = buildActionArgs(entries);

    // when / then
    await expect(action(args)).rejects.toBe(failure);
  });

  it("should send the submitted language on updateTitle intent", async () => {
    // given
    const args = buildActionArgs({ ...UPDATE_TITLE_ENTRIES, language: "fr" });

    // when
    await action(args);

    // then
    expect(vi.mocked(updateWorkbook).mock.calls[0][2].language).toBe("fr");
  });

  it("should return the invalid error key when updateTitle intent has no language", async () => {
    // given
    const args = buildActionArgs(UPDATE_TITLE_ENTRIES);

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: false, errorKey: "workbooks.detail.errors.invalid" });
  });

  it("should return the invalid error key when updateTitle is rejected as invalid", async () => {
    // given
    vi.mocked(updateWorkbook).mockRejectedValue(new WorkbookValidationError());
    const args = buildActionArgs({ ...UPDATE_TITLE_ENTRIES, language: "fr" });

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: false, errorKey: "workbooks.detail.errors.invalid" });
  });

  it("should rethrow the error when updateTitle fails for another reason", async () => {
    // given
    const failure = new Response("Failed to update workbook", { status: 500 });
    vi.mocked(updateWorkbook).mockRejectedValue(failure);
    const args = buildActionArgs({ ...UPDATE_TITLE_ENTRIES, language: "fr" });

    // when / then
    await expect(action(args)).rejects.toBe(failure);
  });
});

const BOUNDARY_TEXT = "route error boundary";

const WORKBOOK: Workbook = {
  workbookId: "wb-1",
  spaceId: "space-1",
  ownerId: "user-1",
  organizationId: "org-1",
  title: "Phrasal verbs",
  description: "",
  visibility: "private",
  language: "fr",
  createdAt: "2026-11-01T00:00:00Z",
  updatedAt: "2026-11-01T00:00:00Z",
};

function wordFillQuestion(questionId: string, sourceText: string): Question {
  return {
    questionId,
    questionType: "word_fill",
    content: JSON.stringify({
      source: { text: sourceText, lang: "ja" },
      target: { text: "{{throw}} it {{away}}", lang: "en" },
    }),
    orderIndex: 0,
    createdAt: "2026-11-01T00:00:00Z",
    updatedAt: "2026-11-01T00:00:00Z",
  };
}

const QUESTIONS = [wordFillQuestion("q-1", "ゴミを捨てる"), wordFillQuestion("q-2", "電気を消す")];

function rejectAsInvalid() {
  return data({ ok: false, errorKey: QUESTION_INVALID_ERROR_KEY }, { status: 400 });
}

function revalidatedTitle(loadCount: number) {
  return `${WORKBOOK.title} ${loadCount}`;
}

function renderDetail(editAction: (args: { request: Request }) => unknown) {
  let loadCount = 0;
  const Stub = createRoutesStub([
    {
      path: "/workbooks/:workbookId",
      Component: WorkbookDetail,
      HydrateFallback: () => null,
      ErrorBoundary: () => <p>{BOUNDARY_TEXT}</p>,
      loader: () => {
        loadCount += 1;
        return {
          workbook: { ...WORKBOOK, title: revalidatedTitle(loadCount) },
          questions: QUESTIONS,
        };
      },
      action: editAction,
      shouldRevalidate: () => true,
    },
  ]);
  render(
    <ConfirmDialogProvider>
      <Stub initialEntries={["/workbooks/wb-1"]} />
    </ConfirmDialogProvider>,
  );
  return userEvent.setup();
}

async function openEditor(user: ReturnType<typeof userEvent.setup>, index: number) {
  const editButtons = await screen.findAllByRole("button", {
    name: i18n.t("workbooks.detail.editQuestion"),
  });
  await user.click(editButtons[index]);
}

async function failOnce(user: ReturnType<typeof userEvent.setup>, index: number) {
  await openEditor(user, index);
  await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));
  await screen.findByText(i18n.t(QUESTION_INVALID_ERROR_KEY));
  await user.click(screen.getByRole("button", { name: i18n.t("common.cancel") }));
}

function deferredAction() {
  let settle: (outcome: { value?: unknown; error?: unknown }) => void = () => {};
  const editAction = vi.fn(
    () =>
      new Promise((resolve, reject) => {
        settle = ({ value, error }) => (error === undefined ? resolve(value) : reject(error));
      }),
  );
  return { editAction, settle: (outcome: { value?: unknown; error?: unknown }) => settle(outcome) };
}

async function closeWhileSaving(
  user: ReturnType<typeof userEvent.setup>,
  editAction: ReturnType<typeof deferredAction>["editAction"],
) {
  await openEditor(user, 0);
  await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));
  await vi.waitFor(() => expect(editAction).toHaveBeenCalled());
  await user.click(screen.getByRole("button", { name: i18n.t("common.cancel") }));
}

async function saveSuccessfully(user: ReturnType<typeof userEvent.setup>) {
  await openEditor(user, 0);
  await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));
  await screen.findByText(revalidatedTitle(2));
}

describe("WorkbookDetail edit sheet", () => {
  it("should show the error key in the edit form when the input is rejected", async () => {
    // given
    const user = renderDetail(rejectAsInvalid);
    await openEditor(user, 0);

    // when
    await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));

    // then
    expect(await screen.findByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).toBeInTheDocument();
  });

  it("should not show the previous error when another question is opened", async () => {
    // given
    const user = renderDetail(rejectAsInvalid);
    await failOnce(user, 0);

    // when
    await openEditor(user, 1);

    // then
    expect(screen.queryByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).not.toBeInTheDocument();
  });

  it("should not show the previous error when the same question is reopened", async () => {
    // given
    const user = renderDetail(rejectAsInvalid);
    await failOnce(user, 0);

    // when
    await openEditor(user, 0);

    // then
    expect(screen.queryByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).not.toBeInTheDocument();
  });

  it("should not show the error that arrives after the sheet was closed while submitting", async () => {
    // given
    const { editAction, settle } = deferredAction();
    const user = renderDetail(editAction);
    await closeWhileSaving(user, editAction);
    await openEditor(user, 0);

    // when
    act(() => settle({ value: rejectAsInvalid() }));
    await screen.findByText(revalidatedTitle(2));

    // then
    expect(screen.queryByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).not.toBeInTheDocument();
  });

  it("should show the route error boundary when the save fails after the sheet was closed", async () => {
    // given
    const { editAction, settle } = deferredAction();
    const user = renderDetail(editAction);
    await closeWhileSaving(user, editAction);

    // when
    act(() => settle({ error: new Response("Failed to update question", { status: 500 }) }));

    // then
    expect(await screen.findByText(BOUNDARY_TEXT)).toBeInTheDocument();
  });

  it("should close the edit sheet when the save succeeds", async () => {
    // given
    const user = renderDetail(() => ({ ok: true }));

    // when
    await saveSuccessfully(user);

    // then
    expect(screen.queryByRole("button", { name: i18n.t("common.save") })).not.toBeInTheDocument();
  });

  it("should keep the edit sheet open when the same question is reopened after a save", async () => {
    // given
    const user = renderDetail(() => ({ ok: true }));
    await saveSuccessfully(user);

    // when
    await openEditor(user, 0);

    // then
    expect(screen.getByRole("button", { name: i18n.t("common.save") })).toBeInTheDocument();
  });
});

describe("WorkbookDetail header", () => {
  it("should submit the workbook's language when the title is saved", async () => {
    // given
    let submitted: FormData | undefined;
    const user = renderDetail(async ({ request }) => {
      submitted = await request.formData();
      return { ok: true };
    });
    await user.click(
      await screen.findByRole("button", { name: i18n.t("workbooks.detail.editWorkbook") }),
    );

    // when
    await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));

    // then
    await vi.waitFor(() => expect(submitted?.get("language")).toBe(WORKBOOK.language));
  });
});
