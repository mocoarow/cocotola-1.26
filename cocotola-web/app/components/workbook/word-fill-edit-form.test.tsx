import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import i18n from "~/i18n/config";
import type { Question } from "~/lib/api/question.server";
import { SAME_LANGUAGE_ERROR_KEY } from "./lang-pair";
import { QUESTION_INVALID_ERROR_KEY, type QuestionEditFetcher } from "./question-form-error";
import { WordFillEditForm } from "./word-fill-edit-form";

const fetcher = { state: "idle", Form: "form" } as unknown as QuestionEditFetcher;

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

function renderForm(formFetcher = fetcher, parsed = PARSED) {
  return render(
    <WordFillEditForm
      question={QUESTION}
      parsed={parsed}
      fetcher={formFetcher}
      onCancel={vi.fn()}
    />,
  );
}

function saveButton() {
  return screen.getByRole("button", { name: i18n.t("common.save") });
}

describe("WordFillEditForm", () => {
  it("should disable the save button when the target language equals the source language", async () => {
    // given
    const user = userEvent.setup();
    renderForm();

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "ja");

    // then
    expect(saveButton()).toBeDisabled();
  });

  it("should show the same language error when the source language equals the target language", async () => {
    // given
    const user = userEvent.setup();
    renderForm();

    // when
    await user.selectOptions(screen.getByLabelText("Source Language"), "en");

    // then
    expect(screen.getByText(i18n.t(SAME_LANGUAGE_ERROR_KEY))).toBeInTheDocument();
  });

  it("should mark the target language invalid when it equals the source language", async () => {
    // given
    const user = userEvent.setup();
    renderForm();

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "ja");

    // then
    expect(screen.getByLabelText("Target Language")).toHaveAttribute("aria-invalid", "true");
  });

  it("should describe the target language with the same language error when it equals the source language", async () => {
    // given
    const user = userEvent.setup();
    renderForm();

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "ja");

    // then
    expect(screen.getByLabelText("Target Language")).toHaveAccessibleDescription(
      i18n.t(SAME_LANGUAGE_ERROR_KEY),
    );
  });

  it("should enable the save button when the target language differs from the source language again", async () => {
    // given
    const user = userEvent.setup();
    renderForm();
    await user.selectOptions(screen.getByLabelText("Target Language"), "ja");

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "fr");

    // then
    expect(saveButton()).toBeEnabled();
  });

  it("should disable the save button on open when the saved languages are the same", () => {
    // given
    const parsed = { ...PARSED, target: { ...PARSED.target, lang: "ja" } };

    // when
    renderForm(fetcher, parsed);

    // then
    expect(saveButton()).toBeDisabled();
  });

  it("should show the error key returned by the action", () => {
    // given
    const rejected = {
      ...fetcher,
      data: { ok: false, errorKey: QUESTION_INVALID_ERROR_KEY },
    } as unknown as QuestionEditFetcher;

    // when
    renderForm(rejected);

    // then
    expect(screen.getByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).toBeInTheDocument();
  });

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
