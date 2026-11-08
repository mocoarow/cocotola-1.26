import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import i18n from "~/i18n/config";
import type { Question } from "~/lib/api/question.server";
import { MultipleChoiceEditForm } from "./multiple-choice-edit-form";
import { QUESTION_INVALID_ERROR_KEY, type QuestionEditFetcher } from "./question-form-error";

const fetcher = { state: "idle", Form: "form" } as unknown as QuestionEditFetcher;

const QUESTION: Question = {
  questionId: "q-1",
  questionType: "multiple_choice",
  content: "",
  tags: ["level:beginner"],
  orderIndex: 1,
  createdAt: "2026-11-01T00:00:00Z",
  updatedAt: "2026-11-01T00:00:00Z",
};

const PARSED = {
  questionText: "Which is a fruit?",
  choices: [
    { id: "1", text: "apple", isCorrect: true },
    { id: "2", text: "chair", isCorrect: false },
  ],
};

function renderForm(formFetcher = fetcher) {
  return render(
    <MultipleChoiceEditForm
      question={QUESTION}
      parsed={PARSED}
      fetcher={formFetcher}
      onCancel={vi.fn()}
    />,
  );
}

describe("MultipleChoiceEditForm", () => {
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

  it("should describe the tags with the hint", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Tags (optional)")).toHaveAccessibleDescription(
      i18n.t("workbooks.addQuestion.tagsHint"),
    );
  });
});
