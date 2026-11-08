import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import i18n from "~/i18n/config";
import type { Question } from "~/lib/api/question.server";
import { QUESTION_INVALID_ERROR_KEY, type QuestionEditFetcher } from "./question-form-error";
import { WordFillEditForm } from "./word-fill-edit-form";

const fetcher = { state: "idle", Form: "form" } as unknown as QuestionEditFetcher;

const QUESTION: Question = {
  questionId: "q-1",
  questionType: "word_fill",
  content: "",
  tags: ["level:beginner"],
  orderIndex: 1,
  createdAt: "2026-11-01T00:00:00Z",
  updatedAt: "2026-11-01T00:00:00Z",
};

const PARSED = {
  source: { text: "ゴミを捨てる", lang: "ja" },
  target: { text: "{{throw}} it {{away}}", lang: "en" },
};

function renderForm(formFetcher = fetcher) {
  return render(
    <WordFillEditForm
      question={QUESTION}
      parsed={PARSED}
      fetcher={formFetcher}
      onCancel={vi.fn()}
    />,
  );
}

describe("WordFillEditForm", () => {
  it("should show the error key returned by the action", () => {
    // given
    const rejected = {
      ...fetcher,
      data: { ok: false, errorKey: QUESTION_INVALID_ERROR_KEY },
    } as unknown as QuestionEditFetcher;

    // when
    renderForm(rejected);

    // then
    expect(screen.getByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).toBeInTheDocument();
  });

  it("should describe the target text with the hint", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Target Text")).toHaveAccessibleDescription(
      i18n.t("workbooks.addQuestion.targetHint"),
    );
  });

  it("should describe the tags with the hint", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Tags (optional)")).toHaveAccessibleDescription(
      i18n.t("workbooks.addQuestion.tagsHint"),
    );
  });
});
