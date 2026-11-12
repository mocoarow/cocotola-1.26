import { act, render, screen, within } from "@testing-library/react";
import type userEvent from "@testing-library/user-event";
import { vi } from "vitest";

import "~/i18n/config";

import { StudyQuestionShell } from "~/components/study/question-shell/study-question-shell";
import type { StudyQuestion } from "~/lib/api/study.server";

type Segment = { type: "chip" | "fixed"; text: string };

export const chip = (text: string): Segment => ({ type: "chip", text });
export const fixed = (text: string): Segment => ({ type: "fixed", text });

// With Math.random() fixed to 0, three chips [a, b, c] are offered as [b, c, a].
export const THREE_CHIPS = [chip("you"), chip("are"), chip("kind")];

export function makeContent(segments: Segment[], explanation = "") {
  return JSON.stringify({
    source: { text: "あなたは親切です", lang: "ja" },
    target: { lang: "en", segments },
    explanation,
  });
}

function orderingQuestion(content: string): StudyQuestion {
  return { questionId: "q-1", questionType: "ordering", content, orderIndex: 0 };
}

export function renderQuestion(content: string, onAnswer = vi.fn()) {
  render(
    <StudyQuestionShell
      question={orderingQuestion(content)}
      onAnswer={onAnswer}
      onSkip={vi.fn()}
    />,
  );
  return onAnswer;
}

export function pool() {
  return screen.getByRole("toolbar", { name: "Word chips" });
}

export function answerArea() {
  return screen.getByRole("toolbar", { name: "Your answer" });
}

export function revealedAnswer() {
  return screen.getByRole("group", { name: "Your answer" });
}

export function buttonTexts(container: HTMLElement): string[] {
  return within(container)
    .queryAllByRole("button")
    .map((b) => b.textContent ?? "");
}

export function tabStops(container: HTMLElement): string[] {
  return within(container)
    .queryAllByRole("button")
    .filter((b) => b.tabIndex === 0)
    .map((b) => b.textContent ?? "");
}

export function focusChip(container: HTMLElement, name: string) {
  act(() => within(container).getByRole("button", { name }).focus());
}

export async function tapChips(user: ReturnType<typeof userEvent.setup>, texts: string[]) {
  for (const text of texts) {
    await user.click(within(pool()).getByRole("button", { name: text }));
  }
}
