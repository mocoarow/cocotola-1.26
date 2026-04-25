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

  it("should disable check button when inputs are empty", () => {
    // given
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);

    // then
    expect(screen.getByText("Check")).toBeDisabled();
  });

  it("should enable check button when all inputs are filled", async () => {
    // given
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByRole("textbox"), "hello");

    // then
    expect(screen.getByText("Check")).toBeEnabled();
  });

  it("should show correct feedback after check", async () => {
    // given
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByRole("textbox"), "hello");
    await user.click(screen.getByText("Check"));

    // then
    expect(screen.getByText("Correct!")).toBeInTheDocument();
    expect(screen.getByText("Next")).toBeInTheDocument();
  });

  it("should show incorrect feedback and correct answer after check", async () => {
    // given
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByRole("textbox"), "wrong");
    await user.click(screen.getByText("Check"));

    // then
    expect(screen.getByText("Incorrect")).toBeInTheDocument();
    expect(screen.getByText("hello")).toBeInTheDocument();
  });

  it("should call onAnswer with true when correct and next is clicked", async () => {
    // given
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByRole("textbox"), "hello");
    await user.click(screen.getByText("Check"));
    await user.click(screen.getByText("Next"));

    // then
    expect(onAnswer).toHaveBeenCalledWith(true);
  });

  it("should call onAnswer with false when incorrect and next is clicked", async () => {
    // given
    const content = makeContent("{{hello}}");
    const onAnswer = vi.fn();
    const user = userEvent.setup();

    // when
    render(<WordFillCard content={content} onAnswer={onAnswer} />);
    await user.type(screen.getByRole("textbox"), "wrong");
    await user.click(screen.getByText("Check"));
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
