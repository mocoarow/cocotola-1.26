import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

import { MultipleChoiceCard } from "./multiple-choice-card";

type ChoiceInput = { id: string; text: string; isCorrect: boolean };

function makeContent(
  questionText: string,
  choices: ChoiceInput[],
  options: { showCorrectCount?: boolean; explanation?: string } = {},
) {
  return JSON.stringify({
    questionText,
    choices,
    shuffleChoices: false,
    showCorrectCount: options.showCorrectCount ?? false,
    explanation: options.explanation ?? "",
  });
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

  it("should reveal Correct! after checking the only correct choice", async () => {
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
    await user.click(screen.getByRole("button", { name: "Check" }));

    // then
    expect(screen.getByText("Correct!")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Next" })).toBeInTheDocument();
  });

  it("should reveal Incorrect when only a wrong choice is checked", async () => {
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
    await user.click(screen.getByRole("button", { name: "Check" }));

    // then
    expect(screen.getByText("Incorrect")).toBeInTheDocument();
  });

  it("should call onAnswer with the selection and true when only the correct choice is selected then Next", async () => {
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
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then
    expect(onAnswer).toHaveBeenCalledTimes(1);
    const [ids, correct] = onAnswer.mock.calls[0];
    expect([...ids].sort()).toEqual(["1"]);
    expect(correct).toBe(true);
  });

  it("should call onAnswer with the selection and false when the wrong choice is selected then Next", async () => {
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
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then
    expect(onAnswer).toHaveBeenCalledTimes(1);
    const [ids, correct] = onAnswer.mock.calls[0];
    expect([...ids].sort()).toEqual(["2"]);
    expect(correct).toBe(false);
  });

  it("should call onAnswer with all correct ids and true for a multi-correct question", async () => {
    // given
    const content = makeContent("Pick the oceans", [
      { id: "1", text: "Pacific", isCorrect: true },
      { id: "2", text: "Atlantic", isCorrect: true },
      { id: "3", text: "Mt Fuji", isCorrect: false },
      { id: "4", text: "Amazon River", isCorrect: false },
    ]);
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("Pacific"));
    await user.click(screen.getByText("Atlantic"));
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then
    expect(onAnswer).toHaveBeenCalledTimes(1);
    const [ids, correct] = onAnswer.mock.calls[0];
    expect([...ids].sort()).toEqual(["1", "2"]);
    expect(correct).toBe(true);
  });

  it("should call onAnswer with the partial selection and false when only one of multiple correct choices is selected", async () => {
    // given
    const content = makeContent("Pick the oceans", [
      { id: "1", text: "Pacific", isCorrect: true },
      { id: "2", text: "Atlantic", isCorrect: true },
      { id: "3", text: "Mt Fuji", isCorrect: false },
      { id: "4", text: "Amazon River", isCorrect: false },
    ]);
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("Pacific"));
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then
    expect(onAnswer).toHaveBeenCalledTimes(1);
    const [ids, correct] = onAnswer.mock.calls[0];
    expect([...ids].sort()).toEqual(["1"]);
    expect(correct).toBe(false);
  });

  it("should call onAnswer with the superset and false when an incorrect choice is added to all correct ones", async () => {
    // given
    const content = makeContent("Pick the oceans", [
      { id: "1", text: "Pacific", isCorrect: true },
      { id: "2", text: "Atlantic", isCorrect: true },
      { id: "3", text: "Mt Fuji", isCorrect: false },
    ]);
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("Pacific"));
    await user.click(screen.getByText("Atlantic"));
    await user.click(screen.getByText("Mt Fuji"));
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then
    expect(onAnswer).toHaveBeenCalledTimes(1);
    const [ids, correct] = onAnswer.mock.calls[0];
    expect([...ids].sort()).toEqual(["1", "2", "3"]);
    expect(correct).toBe(false);
  });

  it("should toggle off a previously selected choice when clicked again before Check", async () => {
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
    await user.click(screen.getByText("A"));

    // then
    expect(screen.getByRole("button", { name: "Check" })).toBeDisabled();
  });

  it("should disable choices after Check is clicked", async () => {
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
    await user.click(screen.getByRole("button", { name: "Check" }));

    // then
    expect(screen.getByText("A").closest("button")).toBeDisabled();
    expect(screen.getByText("B").closest("button")).toBeDisabled();
  });

  it("should keep Check disabled when nothing is selected", () => {
    // given
    const content = makeContent("Q?", [
      { id: "1", text: "A", isCorrect: true },
      { id: "2", text: "B", isCorrect: false },
    ]);
    const onAnswer = vi.fn();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);

    // then
    expect(screen.getByRole("button", { name: "Check" })).toBeDisabled();
  });

  it("should render fallback when content is invalid", () => {
    // given
    const onAnswer = vi.fn();

    // when
    render(<MultipleChoiceCard content="bad json" onAnswer={onAnswer} />);

    // then
    expect(screen.getByText("bad json")).toBeInTheDocument();
  });

  it("should show explanation after Check", async () => {
    // given
    const content = makeContent("Q?", [{ id: "1", text: "A", isCorrect: true }], {
      explanation: "Because reasons",
    });
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("A"));
    await user.click(screen.getByRole("button", { name: "Check" }));

    // then
    expect(screen.getByText("Because reasons")).toBeInTheDocument();
  });

  it("should show 'Select N' hint when showCorrectCount is true", () => {
    // given
    const content = makeContent(
      "Pick the oceans",
      [
        { id: "1", text: "Pacific", isCorrect: true },
        { id: "2", text: "Atlantic", isCorrect: true },
        { id: "3", text: "Mt Fuji", isCorrect: false },
      ],
      { showCorrectCount: true },
    );
    const onAnswer = vi.fn();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);

    // then
    expect(screen.getByText("Select 2")).toBeInTheDocument();
  });

  it("should not show the 'Select N' hint when showCorrectCount is false", () => {
    // given
    const content = makeContent(
      "Pick the oceans",
      [
        { id: "1", text: "Pacific", isCorrect: true },
        { id: "2", text: "Atlantic", isCorrect: true },
      ],
      { showCorrectCount: false },
    );
    const onAnswer = vi.fn();

    // when
    render(<MultipleChoiceCard content={content} onAnswer={onAnswer} />);

    // then
    expect(screen.queryByText("Select 2")).not.toBeInTheDocument();
  });
});
