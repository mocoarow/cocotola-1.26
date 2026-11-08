import { render, screen } from "@testing-library/react";
import type { useFetcher } from "react-router";
import { describe, expect, it, vi } from "vitest";
import i18n from "~/i18n/config";
import type { Question } from "~/lib/api/question.server";
import { WordFillEditForm } from "./word-fill-edit-form";

const fetcher = { state: "idle", Form: "form" } as unknown as ReturnType<typeof useFetcher>;

const QUESTION: Question = {
  questionId: "q-1",
  questionType: "word_fill",
  content: "",
  tags: ["level:beginner"],
  orderIndex: 1,
  createdAt: "2026-11-01T00:00:00Z",
  updatedAt: "2026-11-01T00:00:00Z",
};

const PARSED = {
  source: { text: "ゴミを捨てる", lang: "ja" },
  target: { text: "{{throw}} it {{away}}", lang: "en" },
};

function renderForm() {
  return render(
    <WordFillEditForm question={QUESTION} parsed={PARSED} fetcher={fetcher} onCancel={vi.fn()} />,
  );
}

describe("WordFillEditForm", () => {
  it("should describe the target text with the hint", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Target Text")).toHaveAccessibleDescription(
      i18n.t("workbooks.addQuestion.targetHint"),
    );
  });

  it("should describe the tags with the hint", () => {
    // given / when
    renderForm();

    // then
    expect(screen.getByLabelText("Tags (optional)")).toHaveAccessibleDescription(
      i18n.t("workbooks.addQuestion.tagsHint"),
    );
  });
});
