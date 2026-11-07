import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDialogProvider } from "~/components/confirm-dialog-provider";
import type { Question } from "~/lib/api/question.server";

import "~/i18n/config";

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useFetcher: vi.fn(() => ({ state: "idle", Form: "form", submit: vi.fn() })),
  };
});

import { QuestionCard } from "./question-card";

const ORDERING_CONTENT = JSON.stringify({
  source: { text: "私は学生です。", lang: "ja" },
  target: {
    lang: "en",
    segments: [
      { type: "chip", text: "i" },
      { type: "chip", text: "am" },
      { type: "chip", text: "a student" },
      { type: "fixed", text: "." },
    ],
  },
  explanation: "be 動詞",
});

function buildQuestion(overrides: Partial<Question> = {}): Question {
  return {
    questionId: "q-1",
    questionType: "ordering",
    content: ORDERING_CONTENT,
    orderIndex: 0,
    createdAt: "2026-11-01T00:00:00Z",
    updatedAt: "2026-11-01T00:00:00Z",
    ...overrides,
  };
}

function renderCard(question: Question = buildQuestion()) {
  render(
    <ConfirmDialogProvider>
      <QuestionCard question={question} onEdit={vi.fn()} />
    </ConfirmDialogProvider>,
  );
}

describe("QuestionCard for ordering", () => {
  it("should show the ordering badge", () => {
    // given / when
    renderCard();

    // then
    expect(screen.getByText("Ordering")).toBeInTheDocument();
  });

  it("should show the source text", () => {
    // given / when
    renderCard();

    // then
    expect(screen.getByText("私は学生です。")).toBeInTheDocument();
  });

  it("should show the correct order with capitalized chips", () => {
    // given / when
    renderCard();

    // then
    expect(screen.getByRole("list", { name: "Correct order" })).toHaveTextContent("Iama student.");
  });

  it("should show the explanation", () => {
    // given / when
    renderCard();

    // then
    expect(screen.getByText("be 動詞")).toBeInTheDocument();
  });

  it("should show the raw content when the content cannot be parsed", () => {
    // given / when
    renderCard(buildQuestion({ content: "not json" }));

    // then
    expect(screen.getByText("not json")).toBeInTheDocument();
  });
});
