import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ConfirmDialogProvider } from "~/components/confirm-dialog-provider";
import type { Question } from "~/lib/api/question.server";

import "~/i18n/config";

const submitMock = vi.fn();

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useFetcher: vi.fn(() => ({ state: "idle", Form: "form", submit: submitMock })),
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

function renderCard(question: Question = buildQuestion(), onEdit = vi.fn()) {
  render(
    <ConfirmDialogProvider>
      <QuestionCard question={question} onEdit={onEdit} />
    </ConfirmDialogProvider>,
  );
}

const WORD_FILL_CONTENT = JSON.stringify({
  source: { text: "ゴミを捨てる", lang: "ja" },
  target: { text: "{{throw}} it {{away}}", lang: "en" },
});

const MULTIPLE_CHOICE_CONTENT = JSON.stringify({
  questionText: "Which is a fruit?",
  choices: [
    { id: "1", text: "Apple", isCorrect: true },
    { id: "2", text: "Desk", isCorrect: false },
  ],
});

const HEADER_CASES = [
  { questionType: "word_fill", content: WORD_FILL_CONTENT, badge: "Word Fill" },
  { questionType: "multiple_choice", content: MULTIPLE_CHOICE_CONTENT, badge: "Multiple Choice" },
  { questionType: "ordering", content: ORDERING_CONTENT, badge: "Ordering" },
];

describe.each(HEADER_CASES)("QuestionCard header for $questionType", (tt) => {
  const question = () =>
    buildQuestion({
      questionType: tt.questionType,
      content: tt.content,
      tags: ["level:beginner"],
    });

  it("should show the type badge", () => {
    // given / when
    renderCard(question());

    // then
    expect(screen.getByText(tt.badge)).toBeInTheDocument();
  });

  it("should show the tag", () => {
    // given / when
    renderCard(question());

    // then
    expect(screen.getByText("level:beginner")).toBeInTheDocument();
  });

  it("should show the edit button", () => {
    // given / when
    renderCard(question());

    // then
    expect(screen.getByRole("button", { name: "Edit question" })).toBeInTheDocument();
  });

  it("should show the delete button", () => {
    // given / when
    renderCard(question());

    // then
    expect(screen.getByRole("button", { name: "Delete question" })).toBeInTheDocument();
  });
});

describe("QuestionCard actions", () => {
  afterEach(() => {
    submitMock.mockReset();
  });

  it("should submit deleteQuestion when the deletion is confirmed", async () => {
    // given
    const user = userEvent.setup();
    renderCard(buildQuestion({ questionId: "q-42" }));

    // when
    await user.click(screen.getByRole("button", { name: "Delete question" }));
    await user.click(screen.getByRole("button", { name: "Delete" }));

    // then
    expect(submitMock).toHaveBeenCalledWith(
      { intent: "deleteQuestion", questionId: "q-42" },
      { method: "post" },
    );
  });

  it("should not submit when the deletion is canceled", async () => {
    // given
    const user = userEvent.setup();
    renderCard();

    // when
    await user.click(screen.getByRole("button", { name: "Delete question" }));
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    // then
    expect(submitMock).not.toHaveBeenCalled();
  });

  it("should pass the question to onEdit when the edit button is clicked", async () => {
    // given
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const question = buildQuestion({ questionId: "q-42" });
    renderCard(question, onEdit);

    // when
    await user.click(screen.getByRole("button", { name: "Edit question" }));

    // then
    expect(onEdit).toHaveBeenCalledWith(question);
  });
});

describe.each(HEADER_CASES)("QuestionCard body for unparsable $questionType", (tt) => {
  it("should show the raw content", () => {
    // given / when
    renderCard(buildQuestion({ questionType: tt.questionType, content: "not json" }));

    // then
    expect(screen.getByText("not json")).toBeInTheDocument();
  });
});

describe("QuestionCard for an unknown type", () => {
  const unknownQuestion = () =>
    buildQuestion({ questionType: "unknown_type", content: "raw unknown content" });

  it("should show the type name as the badge", () => {
    // given / when
    renderCard(unknownQuestion());

    // then
    expect(screen.getByText("unknown_type")).toBeInTheDocument();
  });

  it("should show the raw content", () => {
    // given / when
    renderCard(unknownQuestion());

    // then
    expect(screen.getByText("raw unknown content")).toBeInTheDocument();
  });

  it("should not show the edit button", () => {
    // given / when
    renderCard(unknownQuestion());

    // then
    expect(screen.queryByRole("button", { name: "Edit question" })).not.toBeInTheDocument();
  });
});

describe("QuestionCard for ordering", () => {
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
});
