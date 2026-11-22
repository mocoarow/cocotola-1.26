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

import { updateWorkbook, WorkbookValidationError } from "~/lib/api/workbook.server";
import { action } from "./workbooks.$workbookId";
import { buildActionArgs, setupActionMocks } from "./workbooks.$workbookId.action.test-helpers";

const MAX_TITLE_LENGTH = 200;

const UPDATE_TITLE_ENTRIES = {
  intent: "updateTitle",
  title: "Phrasal verbs",
  description: "",
  visibility: "private",
};

const SUBMITTED_TITLE_ENTRIES = {
  ...UPDATE_TITLE_ENTRIES,
  title: "  Phrasal verbs  ",
  description: "Common phrasal verbs",
  visibility: "public",
  language: "fr",
};

setupActionMocks();

describe("WorkbookDetail action updateTitle intent", () => {
  it("should return ok when the workbook is updated", async () => {
    // given
    const args = buildActionArgs(SUBMITTED_TITLE_ENTRIES);

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: true });
  });

  it("should return ok when the title is exactly at the limit", async () => {
    // given
    const args = buildActionArgs({
      ...SUBMITTED_TITLE_ENTRIES,
      title: "a".repeat(MAX_TITLE_LENGTH),
    });

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: true });
  });

  it("should send the trimmed title and submitted fields to updateWorkbook", async () => {
    // given
    const args = buildActionArgs(SUBMITTED_TITLE_ENTRIES);

    // when
    await action(args);

    // then
    expect(updateWorkbook).toHaveBeenCalledWith("test-token", "wb-1", {
      title: "Phrasal verbs",
      description: "Common phrasal verbs",
      visibility: "public",
      language: "fr",
    });
  });

  it("should send private visibility when the submitted visibility is unknown", async () => {
    // given
    const args = buildActionArgs({ ...SUBMITTED_TITLE_ENTRIES, visibility: "everyone" });

    // when
    await action(args);

    // then
    expect(vi.mocked(updateWorkbook).mock.calls[0][2].visibility).toBe("private");
  });

  it("should return the title required error key when the title is blank", async () => {
    // given
    const args = buildActionArgs({ ...SUBMITTED_TITLE_ENTRIES, title: "   " });

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: false, errorKey: "workbooks.detail.errors.titleRequired" });
  });

  it("should return the title too long error key when the title exceeds the limit", async () => {
    // given
    const args = buildActionArgs({
      ...SUBMITTED_TITLE_ENTRIES,
      title: "a".repeat(MAX_TITLE_LENGTH + 1),
    });

    // when
    const result = await action(args);

    // then
    expect(result).toEqual({ ok: false, errorKey: "workbooks.detail.errors.titleTooLong" });
  });

  it("should not call updateWorkbook when the title is blank", async () => {
    // given
    const args = buildActionArgs({ ...SUBMITTED_TITLE_ENTRIES, title: "" });

    // when
    await action(args);

    // then
    expect(updateWorkbook).not.toHaveBeenCalled();
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
