import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/question.server", () => ({
  listQuestions: vi.fn(),
  updateQuestion: vi.fn(),
  deleteQuestion: vi.fn(),
  addQuestion: vi.fn(),
}));

vi.mock("~/lib/api/study.server", () => ({
  getStudyQuestions: vi.fn(),
  getStudySummary: vi.fn(),
  recordAnswer: vi.fn(),
  deleteStudyHistory: vi.fn(),
  listStudyRecords: vi.fn(),
}));

vi.mock("~/lib/api/workbook.server", () => ({
  listWorkbooks: vi.fn(),
  createWorkbook: vi.fn(),
  getWorkbook: vi.fn(),
  updateWorkbook: vi.fn(),
  deleteWorkbook: vi.fn(),
}));

import { addQuestion, updateQuestion } from "~/lib/api/question.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { action } from "./workbooks.$workbookId";

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
    vi.clearAllMocks();
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
});
