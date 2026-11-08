import { render, screen } from "@testing-library/react";
import type { useFetcher } from "react-router";
import { describe, expect, it, vi } from "vitest";
import type { OrderingContent } from "~/components/study/question-types/ordering/answer";
import i18n from "~/i18n/config";
import type { Question } from "~/lib/api/question.server";
import { OrderingEditForm } from "./ordering-edit-form";

const fetcher = { state: "idle", Form: "form" } as unknown as ReturnType<typeof useFetcher>;

const QUESTION: Question = {
  questionId: "q-1",
  questionType: "ordering",
  content: "",
  tags: ["level:beginner", "topic:be"],
  orderIndex: 3,
  createdAt: "2026-11-01T00:00:00Z",
  updatedAt: "2026-11-01T00:00:00Z",
};

const PARSED: OrderingContent = {
  source: { text: "あなたは先生ですか？", lang: "ko" },
  target: {
    lang: "fr",
    segments: [
      { type: "chip", text: "are" },
      { type: "chip", text: "you" },
      { type: "chip", text: "a teacher" },
      { type: "fixed", text: "?" },
    ],
  },
  explanation: "疑問文",
};

function renderForm(parsed: OrderingContent | null = PARSED) {
  return render(
    <OrderingEditForm question={QUESTION} parsed={parsed} fetcher={fetcher} onCancel={vi.fn()} />,
  );
}

describe("OrderingEditForm", () => {
  it("should fill the notation formatted from the segments", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Correct sentence")).toHaveValue("are / you / a teacher [[?]]");
  });

  it("should fill the source text", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Source Text")).toHaveValue("あなたは先生ですか？");
  });

  it("should select the source language", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Source Language")).toHaveValue("ko");
  });

  it("should select the target language", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Target Language")).toHaveValue("fr");
  });

  it("should fill the explanation", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Explanation (optional)")).toHaveValue("疑問文");
  });

  it("should fill the tags", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Tags (optional)")).toHaveValue("level:beginner, topic:be");
  });

  it("should describe the notation field with the hint and the note", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Correct sentence")).toHaveAccessibleDescription(
      `${i18n.t("workbooks.addQuestion.orderingTargetHint")} ${i18n.t("workbooks.addQuestion.orderingCapitalizationNote")}`,
    );
  });

  it("should describe the tags field with the hint", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Tags (optional)")).toHaveAccessibleDescription(
      i18n.t("workbooks.addQuestion.tagsHint"),
    );
  });

  it("should submit with the ordering question type", () => {
    // given / when
    const { container } = renderForm();

    // then
    expect(container.querySelector('input[name="questionType"]')).toHaveValue("ordering");
  });

  it("should leave the notation empty when the content cannot be parsed", () => {
    // given / when
    renderForm(null);

    // then
    expect(screen.getByLabelText("Correct sentence")).toHaveValue("");
  });
});
