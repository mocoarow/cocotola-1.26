import { describe, expect, it, vi } from "vitest";

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

import { QUESTION_INVALID_ERROR_KEY } from "~/components/workbook/question-form-error";
import {
  addQuestion,
  deleteQuestion,
  QuestionValidationError,
  updateQuestion,
} from "~/lib/api/question.server";
import { deleteStudyHistory } from "~/lib/api/study.server";
import { updateWorkbook } from "~/lib/api/workbook.server";
import { action } from "./workbooks.$workbookId";
import {
  type ActionArgs,
  buildActionArgs,
  setupActionMocks,
} from "./workbooks.$workbookId.action.test-helpers";

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

const WORD_FILL_REQUIRED_FIELDS = Object.keys(WORD_FILL_ENTRIES);

const MULTIPLE_CHOICE_ENTRIES = {
  questionText: "Which is a fruit?",
  choices: JSON.stringify([
    { id: "1", text: "apple", isCorrect: true },
    { id: "2", text: "chair", isCorrect: false },
  ]),
};

const MULTIPLE_CHOICE_OPTION_ENTRIES = {
  ...MULTIPLE_CHOICE_ENTRIES,
  explanation: "An apple grows on a tree.",
  shuffleChoices: "true",
  showCorrectCount: "true",
  tags: "level:beginner",
};

const EXPECTED_MULTIPLE_CHOICE_CONTENT = JSON.stringify({
  questionText: "Which is a fruit?",
  choices: [
    { id: "1", text: "apple", isCorrect: true },
    { id: "2", text: "chair", isCorrect: false },
  ],
  explanation: "An apple grows on a tree.",
  displayCount: 2,
  showCorrectCount: true,
  shuffleChoices: true,
  allowPartialCredit: false,
});

const UPDATE_ENTRIES = { intent: "updateQuestion", questionId: "q-1" };

const UPDATE_WORD_FILL_ENTRIES = {
  ...UPDATE_ENTRIES,
  questionType: "word_fill",
  ...WORD_FILL_ENTRIES,
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

const ACTION_APIS = [
  { name: "updateWorkbook", api: updateWorkbook },
  { name: "updateQuestion", api: updateQuestion },
  { name: "deleteQuestion", api: deleteQuestion },
  { name: "deleteStudyHistory", api: deleteStudyHistory },
  { name: "addQuestion", api: addQuestion },
];

async function catchThrown(args: ActionArgs): Promise<unknown> {
  try {
    await action(args);
  } catch (error) {
    return error;
  }
  return undefined;
}

setupActionMocks();

describe("WorkbookDetail action updateQuestion intent", () => {
  it("should return ok when the question is updated", async () => {
    // given
    const args = buildActionArgs(UPDATE_WORD_FILL_ENTRIES);

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: true });
  });

  it("should send the word_fill content with every explanation to updateQuestion", async () => {
    // given
    const args = buildActionArgs({
      ...UPDATE_WORD_FILL_ENTRIES,
      explanation: "to discard",
      explanation1: "throw = 投げる",
      explanation2: "away = 離れて",
      orderIndex: "3",
    });

    // when
    await action(args);

    // then
    expect(updateQuestion).toHaveBeenCalledWith("test-token", "wb-1", "q-1", {
      content: JSON.stringify({
        source: { text: "ゴミを捨てる", lang: "ja" },
        target: { text: "{{throw}} it {{away}}", lang: "en" },
        explanation: "to discard",
        explanation1: "throw = 投げる",
        explanation2: "away = 離れて",
      }),
      tags: undefined,
      orderIndex: 3,
    });
  });

  it("should update as word_fill when the questionType is missing", async () => {
    // given
    const args = buildActionArgs({ ...UPDATE_ENTRIES, ...WORD_FILL_ENTRIES });

    // when
    await action(args);

    // then
    expect(vi.mocked(updateQuestion).mock.calls[0][3].content).toBe(
      JSON.stringify({
        source: { text: "ゴミを捨てる", lang: "ja" },
        target: { text: "{{throw}} it {{away}}", lang: "en" },
      }),
    );
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

  it.each(
    WORD_FILL_REQUIRED_FIELDS,
  )("should throw a 400 response when the word_fill %s is blank", async (field) => {
    // given
    const args = buildActionArgs({ ...UPDATE_WORD_FILL_ENTRIES, [field]: "  " });

    // when
    const thrown = await catchThrown(args);

    // then
    expect(thrown).toMatchObject({ status: 400 });
  });

  it("should send the multiple_choice content to updateQuestion", async () => {
    // given
    const args = buildActionArgs({
      ...UPDATE_ENTRIES,
      questionType: "multiple_choice",
      orderIndex: "2",
      ...MULTIPLE_CHOICE_OPTION_ENTRIES,
    });

    // when
    await action(args);

    // then
    expect(updateQuestion).toHaveBeenCalledWith("test-token", "wb-1", "q-1", {
      content: EXPECTED_MULTIPLE_CHOICE_CONTENT,
      tags: ["level:beginner"],
      orderIndex: 2,
    });
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

  it("should throw a 400 response when the questionId is missing", async () => {
    // given
    const args = buildActionArgs({
      intent: "updateQuestion",
      questionType: "word_fill",
      ...WORD_FILL_ENTRIES,
    });

    // when
    const thrown = await catchThrown(args);

    // then
    expect(thrown).toMatchObject({ status: 400 });
  });

  it("should throw a 400 response when the questionType is unknown", async () => {
    // given
    const args = buildActionArgs({ ...UPDATE_WORD_FILL_ENTRIES, questionType: "essay" });

    // when
    const thrown = await catchThrown(args);

    // then
    expect(thrown).toMatchObject({ status: 400 });
  });

  it("should not call updateQuestion when the questionType is unknown", async () => {
    // given
    const args = buildActionArgs({ ...UPDATE_WORD_FILL_ENTRIES, questionType: "essay" });

    // when
    await catchThrown(args);

    // then
    expect(updateQuestion).not.toHaveBeenCalled();
  });
});

describe("WorkbookDetail action deleteQuestion intent", () => {
  it("should return ok when the question is deleted", async () => {
    // given
    const args = buildActionArgs({ intent: "deleteQuestion", questionId: "q-1" });

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: true });
  });

  it("should send the questionId to deleteQuestion", async () => {
    // given
    const args = buildActionArgs({ intent: "deleteQuestion", questionId: "q-1" });

    // when
    await action(args);

    // then
    expect(deleteQuestion).toHaveBeenCalledWith("test-token", "wb-1", "q-1");
  });

  it("should throw a 400 response when the questionId is missing", async () => {
    // given
    const args = buildActionArgs({ intent: "deleteQuestion" });

    // when
    const thrown = await catchThrown(args);

    // then
    expect(thrown).toMatchObject({ status: 400 });
  });

  it("should not call deleteQuestion when the questionId is missing", async () => {
    // given
    const args = buildActionArgs({ intent: "deleteQuestion", questionId: "" });

    // when
    await catchThrown(args);

    // then
    expect(deleteQuestion).not.toHaveBeenCalled();
  });
});

describe("WorkbookDetail action clearStudyHistory intent", () => {
  it("should report the clearing when the study history is deleted", async () => {
    // given
    const args = buildActionArgs({ intent: "clearStudyHistory" });

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: true, cleared: true });
  });

  it("should send the workbookId to deleteStudyHistory", async () => {
    // given
    const args = buildActionArgs({ intent: "clearStudyHistory" });

    // when
    await action(args);

    // then
    expect(deleteStudyHistory).toHaveBeenCalledWith("test-token", "wb-1");
  });
});

describe("WorkbookDetail action addWordFill intent", () => {
  it("should report the addition when the word_fill question is added", async () => {
    // given
    const args = buildActionArgs({ intent: "addWordFill", ...WORD_FILL_ENTRIES });

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: true, added: true });
  });

  it("should keep only the explanation in the word_fill content on addWordFill", async () => {
    // given
    const args = buildActionArgs({
      intent: "addWordFill",
      ...WORD_FILL_ENTRIES,
      explanation: "to discard",
      explanation1: "throw = 投げる",
    });

    // when
    await action(args);

    // then
    expect(addQuestion).toHaveBeenCalledWith("test-token", "wb-1", {
      questionType: "word_fill",
      content: JSON.stringify({
        source: { text: "ゴミを捨てる", lang: "ja" },
        target: { text: "{{throw}} it {{away}}", lang: "en" },
        explanation: "to discard",
      }),
      tags: undefined,
      orderIndex: 0,
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

  it.each(
    WORD_FILL_REQUIRED_FIELDS,
  )("should throw a 400 response when the %s is blank", async (field) => {
    // given
    const args = buildActionArgs({ intent: "addWordFill", ...WORD_FILL_ENTRIES, [field]: "  " });

    // when
    const thrown = await catchThrown(args);

    // then
    expect(thrown).toMatchObject({ status: 400 });
  });

  it("should not call addQuestion when a required field is blank", async () => {
    // given
    const args = buildActionArgs({ intent: "addWordFill", ...WORD_FILL_ENTRIES, sourceText: "" });

    // when
    await catchThrown(args);

    // then
    expect(addQuestion).not.toHaveBeenCalled();
  });
});

describe("WorkbookDetail action addMultipleChoice intent", () => {
  it("should report the addition when the multiple_choice question is added", async () => {
    // given
    const args = buildActionArgs({
      intent: "addMultipleChoice",
      ...MULTIPLE_CHOICE_OPTION_ENTRIES,
    });

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: true, added: true });
  });

  it("should send the multiple_choice content to addQuestion", async () => {
    // given
    const args = buildActionArgs({
      intent: "addMultipleChoice",
      ...MULTIPLE_CHOICE_OPTION_ENTRIES,
    });

    // when
    await action(args);

    // then
    expect(addQuestion).toHaveBeenCalledWith("test-token", "wb-1", {
      questionType: "multiple_choice",
      content: EXPECTED_MULTIPLE_CHOICE_CONTENT,
      tags: ["level:beginner"],
      orderIndex: 0,
    });
  });

  it("should throw a 400 response when the questionText is blank", async () => {
    // given
    const args = buildActionArgs({
      intent: "addMultipleChoice",
      ...MULTIPLE_CHOICE_OPTION_ENTRIES,
      questionText: " ",
    });

    // when
    const thrown = await catchThrown(args);

    // then
    expect(thrown).toMatchObject({ status: 400 });
  });

  it("should not call addQuestion when the questionText is blank", async () => {
    // given
    const args = buildActionArgs({
      intent: "addMultipleChoice",
      ...MULTIPLE_CHOICE_OPTION_ENTRIES,
      questionText: "",
    });

    // when
    await catchThrown(args);

    // then
    expect(addQuestion).not.toHaveBeenCalled();
  });
});

describe("WorkbookDetail action addOrdering intent", () => {
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
});

describe("WorkbookDetail action question save failures", () => {
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
});

describe("WorkbookDetail action unknown intent", () => {
  it("should throw a 400 response when the intent is unknown", async () => {
    // given
    const args = buildActionArgs({ intent: "deleteEverything" });

    // when
    const thrown = await catchThrown(args);

    // then
    expect(thrown).toMatchObject({ status: 400 });
  });

  it("should throw a 400 response when the intent is missing", async () => {
    // given
    const args = buildActionArgs({});

    // when
    const thrown = await catchThrown(args);

    // then
    expect(thrown).toMatchObject({ status: 400 });
  });

  it.each(ACTION_APIS)("should not call $name when the intent is unknown", async ({ api }) => {
    // given
    const args = buildActionArgs({ intent: "deleteEverything" });

    // when
    await catchThrown(args);

    // then
    expect(api).not.toHaveBeenCalled();
  });
});
