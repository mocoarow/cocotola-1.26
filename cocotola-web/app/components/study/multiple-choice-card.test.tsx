import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

import { MultipleChoiceCard } from "./multiple-choice-card";

function makeContent(questionText: string, choices: { id: string; text: string; isCorrect: boolean }[]) {
  return JSON.stringify({ questionText, choices, shuffleChoices: false });
}

describe("MultipleChoiceCard", () => {
  it("should render question text and choices", () => {
    // given
    const content = makeContent("What is 1+1?", [
      { id: "1", text: "2", isCorrect: true },
      { id: "2", text: "3", isCorrect: false },
    ]);
    const onAnswer = vi.fn();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);

    // then
    expect(screen.getByText("What is 1+1?")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });

  it("should show correct/incorrect feedback after selecting a choice", async () => {
    // given
    const content = makeContent("Q?", [
      { id: "1", text: "Right", isCorrect: true },
      { id: "2", text: "Wrong", isCorrect: false },
    ]);
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("Right"));

    // then
    expect(screen.getByText("Correct!")).toBeInTheDocument();
    expect(screen.getByText("Next")).toBeInTheDocument();
  });

  it("should show incorrect feedback when wrong choice selected", async () => {
    // given
    const content = makeContent("Q?", [
      { id: "1", text: "Right", isCorrect: true },
      { id: "2", text: "Wrong", isCorrect: false },
    ]);
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("Wrong"));

    // then
    expect(screen.getByText("Incorrect")).toBeInTheDocument();
  });

  it("should call onAnswer with true when correct choice selected and next clicked", async () => {
    // given
    const content = makeContent("Q?", [
      { id: "1", text: "Right", isCorrect: true },
      { id: "2", text: "Wrong", isCorrect: false },
    ]);
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("Right"));
    await user.click(screen.getByText("Next"));

    // then
    expect(onAnswer).toHaveBeenCalledWith(true);
  });

  it("should call onAnswer with false when wrong choice selected and next clicked", async () => {
    // given
    const content = makeContent("Q?", [
      { id: "1", text: "Right", isCorrect: true },
      { id: "2", text: "Wrong", isCorrect: false },
    ]);
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("Wrong"));
    await user.click(screen.getByText("Next"));

    // then
    expect(onAnswer).toHaveBeenCalledWith(false);
  });

  it("should disable choices after selection", async () => {
    // given
    const content = makeContent("Q?", [
      { id: "1", text: "A", isCorrect: true },
      { id: "2", text: "B", isCorrect: false },
    ]);
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("A"));

    // then
    expect(screen.getByText("A").closest("button")).toBeDisabled();
    expect(screen.getByText("B").closest("button")).toBeDisabled();
  });

  it("should render fallback when content is invalid", () => {
    // given
    const onAnswer = vi.fn();

    // when
    render(<MultipleChoiceCard content="bad json" onAnswer={onAnswer} />);

    // then
    expect(screen.getByText("bad json")).toBeInTheDocument();
  });

  it("should show explanation after answering", async () => {
    // given
    const content = JSON.stringify({
      questionText: "Q?",
      choices: [{ id: "1", text: "A", isCorrect: true }],
      explanation: "Because reasons",
      shuffleChoices: false,
    });
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("A"));

    // then
    expect(screen.getByText("Because reasons")).toBeInTheDocument();
  });
});
