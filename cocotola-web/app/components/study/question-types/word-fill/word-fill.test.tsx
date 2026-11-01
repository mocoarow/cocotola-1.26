import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

import { StudyQuestionShell } from "~/components/study/question-shell/study-question-shell";
import type { StudyQuestion, StudyQuestionAudio } from "~/lib/api/study.server";

function makeContent(
  target: string,
  source?: string,
  options: { explanation?: string; explanation1?: string; explanation2?: string } = {},
) {
  return JSON.stringify({
    source: source ? { text: source, lang: "ja" } : undefined,
    target: { text: target, lang: "en" },
    ...options,
  });
}

function wfQuestion(content: string, audio?: StudyQuestionAudio): StudyQuestion {
  return { questionId: "q-1", questionType: "word_fill", content, orderIndex: 0, audio };
}

const correctDraft = { correct: true, payload: { type: "word_fill", correct: true } };
const incorrectDraft = { correct: false, payload: { type: "word_fill", correct: false } };

describe("word_fill question", () => {
  it("should render source text", () => {
    // given
    const content = makeContent("{{hello}}", "こんにちは");
    const onAnswer = vi.fn();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );

    // then
    expect(screen.getByText("こんにちは")).toBeInTheDocument();
  });

  it("should not render source audio button when audio prop is missing", () => {
    // given
    const content = makeContent("{{hello}}", "こんにちは");
    const onAnswer = vi.fn();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );

    // then
    expect(screen.queryByRole("button", { name: /source/i })).not.toBeInTheDocument();
  });

  it("should render source audio button when audio.source.url is provided", () => {
    // given
    const content = makeContent("{{hello}}", "こんにちは");
    const onAnswer = vi.fn();
    const audio = {
      source: { url: "https://example.test/audio/source.opus", durationSec: 1.5 },
    };

    // when
    render(
      <StudyQuestionShell
        question={wfQuestion(content, audio)}
        onAnswer={onAnswer}
        onSkip={vi.fn()}
      />,
    );

    // then
    expect(screen.getByRole("button", { name: /source/i })).toBeInTheDocument();
  });

  it("should not leak target audio button before result phase", () => {
    // given
    const content = makeContent("{{hello}}", "こんにちは");
    const onAnswer = vi.fn();
    const audio = {
      target: { url: "https://example.test/audio/target.opus", durationSec: 1.2 },
    };

    // when
    render(
      <StudyQuestionShell
        question={wfQuestion(content, audio)}
        onAnswer={onAnswer}
        onSkip={vi.fn()}
      />,
    );

    // then
    expect(screen.queryByRole("button", { name: /target/i })).not.toBeInTheDocument();
  });

  it("should render input for each blank", () => {
    // given
    const content = makeContent("{{throw}} it {{away}}");
    const onAnswer = vi.fn();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );

    // then
    const inputs = screen.getAllByRole("textbox");
    expect(inputs).toHaveLength(2);
  });

  it("should focus the first blank on mount", () => {
    // given
    const content = makeContent("{{hello}} {{world}}");
    const onAnswer = vi.fn();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );

    // then
    expect(screen.getByLabelText("Blank 1")).toHaveFocus();
  });

  it("should move focus to the next blank when the current answer is correct", async () => {
    // given
    const content = makeContent("{{hello}} {{world}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
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
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
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
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
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
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
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
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
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
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
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
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
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
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
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
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
    await user.type(screen.getByLabelText("Blank 1"), "hello");
    await user.keyboard("{Enter}");

    // then
    expect(onAnswer).toHaveBeenCalledWith(correctDraft);
  });

  it("should call onAnswer with true only after the user clicks Next on the correct result screen", async () => {
    // given
    const content = makeContent("{{hello}} {{world}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
    await user.type(screen.getByLabelText("Blank 1"), "hello");
    await user.type(screen.getByLabelText("Blank 2"), "world");
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then
    expect(onAnswer).toHaveBeenCalledTimes(1);
    expect(onAnswer).toHaveBeenCalledWith(correctDraft);
  });

  it("should reveal the answer and show next button when show-answer is clicked", async () => {
    // given
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
    await user.click(screen.getByText("Show answer"));

    // then
    expect(screen.getByText("Incorrect")).toBeInTheDocument();
    expect(screen.getByText("Next")).toBeInTheDocument();
    expect(screen.getByText("hello")).toBeInTheDocument();
  });

  it("should show Correct! when revealing a card whose only blank is whitespace", async () => {
    // given
    const content = makeContent("say {{ }} here");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
    await user.click(screen.getByText("Show answer"));

    // then
    expect(screen.getByText("Correct!")).toBeInTheDocument();
  });

  it("should show Incorrect when revealing a card that has no blanks", async () => {
    // given
    const content = makeContent("no blanks here");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
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
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
    await user.click(screen.getByText("Show answer"));
    await user.click(screen.getByText("Next"));

    // then
    expect(onAnswer).toHaveBeenCalledWith(incorrectDraft);
  });

  it.each([
    ["not JSON", "not json"],
    ["missing target text", JSON.stringify({ source: { text: "hi", lang: "en" } })],
  ])("should offer Skip when content is %s", (_label, content) => {
    // given
    const onAnswer = vi.fn();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );

    // then
    expect(screen.getByRole("button", { name: "Skip" })).toBeInTheDocument();
  });

  it("should have aria-label on blank inputs", () => {
    // given
    const content = makeContent("{{hello}} and {{world}}");
    const onAnswer = vi.fn();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );

    // then
    expect(screen.getByLabelText("Blank 1")).toBeInTheDocument();
    expect(screen.getByLabelText("Blank 2")).toBeInTheDocument();
  });

  it("should render explanation1 link before the answer is revealed", () => {
    // given
    const content = makeContent("{{hello}}", undefined, {
      explanation1: "Source: [#1](https://tatoeba.org/en/sentences/show/1)",
    });
    const onAnswer = vi.fn();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );

    // then
    expect(screen.getByRole("link", { name: "#1" })).toBeInTheDocument();
  });

  it("should render explanation2 link after the answer is revealed", async () => {
    // given
    const content = makeContent("{{hello}}", undefined, {
      explanation2: "Source: [#2](https://tatoeba.org/en/sentences/show/2)",
    });
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
    await user.click(screen.getByText("Show answer"));

    // then
    expect(screen.getByRole("link", { name: "#2" })).toBeInTheDocument();
  });

  it("should render the legacy explanation link after reveal when explanation2 is missing", async () => {
    // given
    const content = makeContent("{{hello}}", undefined, {
      explanation: "Source: [#3](https://tatoeba.org/en/sentences/show/3)",
    });
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(
      <StudyQuestionShell question={wfQuestion(content)} onAnswer={onAnswer} onSkip={vi.fn()} />,
    );
    await user.click(screen.getByText("Show answer"));

    // then
    expect(screen.getByRole("link", { name: "#3" })).toBeInTheDocument();
  });
});
