import { render, screen } from "@testing-library/react";
import type { useFetcher } from "react-router";
import { describe, expect, it, vi } from "vitest";
import i18n from "~/i18n/config";
import type { Question } from "~/lib/api/question.server";
import { MultipleChoiceEditForm } from "./multiple-choice-edit-form";

const fetcher = { state: "idle", Form: "form" } as unknown as ReturnType<typeof useFetcher>;

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

describe("MultipleChoiceEditForm", () => {
  it("should describe the tags with the hint", () => {
    // given / when
    render(
      <MultipleChoiceEditForm
        question={QUESTION}
        parsed={PARSED}
        fetcher={fetcher}
        onCancel={vi.fn()}
      />,
    );

    // then
    expect(screen.getByLabelText("Tags (optional)")).toHaveAccessibleDescription(
      i18n.t("workbooks.addQuestion.tagsHint"),
    );
  });
});
