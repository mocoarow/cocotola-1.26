import { act, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

import { StudyQuestionShell } from "~/components/study/question-shell/study-question-shell";
import type { StudyQuestion } from "~/lib/api/study.server";

const CONTENT = JSON.stringify({
  questionText: "Q?",
  choices: [
    { id: "1", text: "Right", isCorrect: true },
    { id: "2", text: "Wrong", isCorrect: false },
  ],
  shuffleChoices: false,
  showCorrectCount: false,
  explanation: "",
});

const QUESTION: StudyQuestion = {
  questionId: "q-1",
  questionType: "multiple_choice",
  content: CONTENT,
  orderIndex: 0,
};

function renderQuestion() {
  const onAnswer = vi.fn();
  render(<StudyQuestionShell question={QUESTION} onAnswer={onAnswer} onSkip={vi.fn()} />);
  return onAnswer;
}

function choice(name: string) {
  return screen.getByRole("button", { name });
}

function focusChoice(name: string) {
  act(() => choice(name).focus());
}

describe("multiple_choice question keyboard", () => {
  describe("Enter with a selection", () => {
    it("should check the answer", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion();
      await user.click(choice("Right"));

      // when
      await user.keyboard("{Enter}");

      // then
      expect(screen.getByRole("button", { name: "Next" })).toBeInTheDocument();
    });

    it("should send the selection", async () => {
      // given
      const user = userEvent.setup();
      const onAnswer = renderQuestion();
      await user.click(choice("Right"));

      // when
      await user.keyboard("{Enter}");
      await user.click(screen.getByRole("button", { name: "Next" }));

      // then
      expect(onAnswer).toHaveBeenCalledWith({
        correct: true,
        payload: { type: "multiple_choice", selectedChoiceIds: ["1"] },
      });
    });

    it("should not toggle the focused choice", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion();
      await user.click(choice("Right"));
      focusChoice("Wrong");

      // when
      await user.keyboard("{Enter}");

      // then
      expect(choice("Wrong")).toHaveAttribute("aria-pressed", "false");
    });

    it("should prevent the default action", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion();
      await user.click(choice("Right"));

      // when
      const notPrevented = fireEvent.keyDown(choice("Right"), { key: "Enter" });

      // then
      expect(notPrevented).toBe(false);
    });
  });

  describe("Enter without a selection", () => {
    it("should not check the answer", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion();
      focusChoice("Right");

      // when
      await user.keyboard("{Enter}");

      // then
      expect(screen.queryByRole("button", { name: "Next" })).not.toBeInTheDocument();
    });

    it("should not select the focused choice", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion();
      focusChoice("Right");

      // when
      await user.keyboard("{Enter}");

      // then
      expect(choice("Right")).toHaveAttribute("aria-pressed", "false");
    });
  });

  describe("other keys", () => {
    it("should toggle the focused choice with Space", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion();
      focusChoice("Right");

      // when
      await user.keyboard(" ");

      // then
      expect(choice("Right")).toHaveAttribute("aria-pressed", "true");
    });

    const modifierTests = [
      { name: "Shift+Enter", keys: "{Shift>}{Enter}{/Shift}" },
      { name: "Ctrl+Enter", keys: "{Control>}{Enter}{/Control}" },
      { name: "Alt+Enter", keys: "{Alt>}{Enter}{/Alt}" },
      { name: "Meta+Enter", keys: "{Meta>}{Enter}{/Meta}" },
    ];
    for (const tt of modifierTests) {
      it(`should not check the answer with ${tt.name}`, async () => {
        // given
        const user = userEvent.setup();
        renderQuestion();
        await user.click(choice("Right"));

        // when
        await user.keyboard(tt.keys);

        // then
        expect(screen.queryByRole("button", { name: "Next" })).not.toBeInTheDocument();
      });
    }
  });
});
