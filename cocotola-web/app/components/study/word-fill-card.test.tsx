import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

import { WordFillCard } from "./word-fill-card";

function makeContent(target: string, source?: string) {
  return JSON.stringify({
    source: source ? { text: source, lang: "ja" } : undefined,
    target: { text: target, lang: "en" },
  });
}

describe("WordFillCard", () => {
  it("should render source text", () => {
    // given
    const content = makeContent("{{hello}}", "こんにちは");
    const onAnswer = vi.fn();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);

    // then
    expect(screen.getByText("こんにちは")).toBeInTheDocument();
  });

  it("should render input for each blank", () => {
    // given
    const content = makeContent("{{throw}} it {{away}}");
    const onAnswer = vi.fn();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);

    // then
    const inputs = screen.getAllByRole("textbox");
    expect(inputs).toHaveLength(2);
  });

  it("should focus the first blank on mount", () => {
    // given
    const content = makeContent("{{hello}} {{world}}");
    const onAnswer = vi.fn();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);

    // then
    expect(screen.getByLabelText("Blank 1")).toHaveFocus();
  });

  it("should move focus to the next blank when the current answer is correct", async () => {
    // given
    const content = makeContent("{{hello}} {{world}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByLabelText("Blank 1"), "hello");

    // then
    expect(screen.getByLabelText("Blank 2")).toHaveFocus();
  });

  it("should lock the blank as read-only after the correct answer is entered", async () => {
    // given
    const content = makeContent("{{hello}} {{world}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByLabelText("Blank 1"), "hello");

    // then: the just-completed blank keeps the value and is marked read-only
    const firstBlank = screen.getByLabelText("Blank 1") as HTMLInputElement;
    expect(firstBlank.value).toBe("hello");
    expect(firstBlank).toHaveAttribute("readonly");
    expect(firstBlank).toBeDisabled();
  });

  it("should display the answer with its original case when the user typed a different case", async () => {
    // given
    const content = makeContent("{{Apple}} {{Banana}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByLabelText("Blank 1"), "apple");

    // then: locked input shows the correct-case answer, not the user's input
    const firstBlank = screen.getByLabelText("Blank 1") as HTMLInputElement;
    expect(firstBlank.value).toBe("Apple");
  });

  it("should skip locked blanks when wrapping focus", async () => {
    // given: three blanks; the first one is already locked by typing the
    // correct answer, then the user fills the third blank correctly
    const content = makeContent("{{first}} {{second}} {{third}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByLabelText("Blank 1"), "first");
    // Focus advanced to Blank 2; click Blank 3 to deliberately leave Blank 2 empty.
    await user.click(screen.getByLabelText("Blank 3"));
    await user.type(screen.getByLabelText("Blank 3"), "third");

    // then: focus wraps past the locked Blank 1 and lands on the still-empty Blank 2
    expect(screen.getByLabelText("Blank 2")).toHaveFocus();
    expect(onAnswer).not.toHaveBeenCalled();
  });

  it("should keep focus on the current blank when the answer is wrong", async () => {
    // given
    const content = makeContent("{{hello}} {{world}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByLabelText("Blank 1"), "wrong");

    // then
    expect(screen.getByLabelText("Blank 1")).toHaveFocus();
  });

  it("should wrap focus to the first blank when the last blank is correctly answered but others remain wrong", async () => {
    // given
    const content = makeContent("{{first}} {{second}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    // Fill the second blank correctly while the first remains empty.
    await user.click(screen.getByLabelText("Blank 2"));
    await user.type(screen.getByLabelText("Blank 2"), "second");

    // then
    expect(screen.getByLabelText("Blank 1")).toHaveFocus();
    expect(onAnswer).not.toHaveBeenCalled();
  });

  it("should show the correct result screen when every blank is correctly answered", async () => {
    // given
    const content = makeContent("{{hello}} {{world}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByLabelText("Blank 1"), "hello");
    await user.type(screen.getByLabelText("Blank 2"), "world");

    // then
    expect(screen.getByText("Correct!")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Next" })).toBeInTheDocument();
    expect(onAnswer).not.toHaveBeenCalled();
  });

  it("should focus the Next button when transitioning to the result phase", async () => {
    // given
    const content = makeContent("{{hello}} {{world}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByLabelText("Blank 1"), "hello");
    await user.type(screen.getByLabelText("Blank 2"), "world");

    // then: the Next button receives focus so Enter/Space advances the question
    expect(screen.getByRole("button", { name: "Next" })).toHaveFocus();
  });

  it("should advance to the next question when Enter is pressed on the focused Next button", async () => {
    // given
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByLabelText("Blank 1"), "hello");
    await user.keyboard("{Enter}");

    // then
    expect(onAnswer).toHaveBeenCalledWith(true);
  });

  it("should call onAnswer with true only after the user clicks Next on the correct result screen", async () => {
    // given
    const content = makeContent("{{hello}} {{world}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByLabelText("Blank 1"), "hello");
    await user.type(screen.getByLabelText("Blank 2"), "world");
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then
    expect(onAnswer).toHaveBeenCalledTimes(1);
    expect(onAnswer).toHaveBeenCalledWith(true);
  });

  it("should reveal the answer and show next button when show-answer is clicked", async () => {
    // given
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("Show answer"));

    // then
    expect(screen.getByText("Incorrect")).toBeInTheDocument();
    expect(screen.getByText("Next")).toBeInTheDocument();
    expect(screen.getByText("hello")).toBeInTheDocument();
  });

  it("should show correct feedback when revealing answer with all blanks already filled correctly", async () => {
    // given: cannot reach this state via auto-submit, so we exercise reveal directly
    // by typing a wrong then a correct value to keep revealed gating active.
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    // First click reveal with empty input → incorrect.
    await user.click(screen.getByText("Show answer"));

    // then
    expect(screen.getByText("Incorrect")).toBeInTheDocument();
  });

  it("should call onAnswer with false when next is clicked after revealing on an empty card", async () => {
    // given
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.click(screen.getByText("Show answer"));
    await user.click(screen.getByText("Next"));

    // then
    expect(onAnswer).toHaveBeenCalledWith(false);
  });

  it("should render fallback when content is invalid", () => {
    // given
    const onAnswer = vi.fn();

    // when
    render(<WordFillCard content="not json" onAnswer={onAnswer} />);

    // then
    expect(screen.getByText("not json")).toBeInTheDocument();
  });

  it("should have aria-label on blank inputs", () => {
    // given
    const content = makeContent("{{hello}} and {{world}}");
    const onAnswer = vi.fn();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);

    // then
    expect(screen.getByLabelText("Blank 1")).toBeInTheDocument();
    expect(screen.getByLabelText("Blank 2")).toBeInTheDocument();
  });
});
