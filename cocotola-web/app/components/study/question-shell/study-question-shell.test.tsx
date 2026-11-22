import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

import type { StudyQuestion } from "~/lib/api/study.server";
import { StudyQuestionShell } from "./study-question-shell";

function question(questionType: string, content: string): StudyQuestion {
  return { questionId: "q-1", questionType, content, orderIndex: 0 };
}

describe("StudyQuestionShell", () => {
  it("should offer Skip when the question type is unknown", () => {
    // given
    const unknown = question("unknown_type", JSON.stringify({ items: ["a", "b"] }));

    // when
    render(<StudyQuestionShell question={unknown} onAnswer={vi.fn()} onSkip={vi.fn()} />);

    // then
    expect(screen.getByRole("button", { name: "Skip" })).toBeInTheDocument();
  });

  it("should not show the raw content when the question cannot be rendered", () => {
    // given
    const unknown = question("unknown_type", "RAW-CONTENT");

    // when
    render(<StudyQuestionShell question={unknown} onAnswer={vi.fn()} onSkip={vi.fn()} />);

    // then
    expect(screen.queryByText("RAW-CONTENT")).not.toBeInTheDocument();
  });

  it("should call onSkip when Skip is clicked", async () => {
    // given
    const user = userEvent.setup();
    const onSkip = vi.fn();
    render(
      <StudyQuestionShell
        question={question("unknown_type", "{}")}
        onAnswer={vi.fn()}
        onSkip={onSkip}
      />,
    );

    // when
    await user.click(screen.getByRole("button", { name: "Skip" }));

    // then
    expect(onSkip).toHaveBeenCalledTimes(1);
  });

  describe("Enter on the Check button", () => {
    const multipleChoice = question(
      "multiple_choice",
      JSON.stringify({
        questionText: "Q?",
        choices: [
          { id: "1", text: "Right", isCorrect: true },
          { id: "2", text: "Wrong", isCorrect: false },
        ],
        shuffleChoices: false,
        showCorrectCount: false,
        explanation: "",
      }),
    );

    async function focusCheckWithSelection(user: ReturnType<typeof userEvent.setup>) {
      await user.click(screen.getByRole("button", { name: "Right" }));
      act(() => screen.getByRole("button", { name: "Check" }).focus());
    }

    it("should not move to the next question when Enter is held", async () => {
      // given
      const user = userEvent.setup();
      const onAnswer = vi.fn();
      render(<StudyQuestionShell question={multipleChoice} onAnswer={onAnswer} onSkip={vi.fn()} />);
      await focusCheckWithSelection(user);

      // when
      await user.keyboard("{Enter>3}");

      // then
      expect(onAnswer).not.toHaveBeenCalled();
    });

    it("should move to the next question when Enter is pressed again after the held Enter", async () => {
      // given
      const user = userEvent.setup();
      const onAnswer = vi.fn();
      render(<StudyQuestionShell question={multipleChoice} onAnswer={onAnswer} onSkip={vi.fn()} />);
      await focusCheckWithSelection(user);

      // when
      await user.keyboard("{Enter>3/}{Enter}");

      // then
      expect(onAnswer).toHaveBeenCalledTimes(1);
    });
  });
});
