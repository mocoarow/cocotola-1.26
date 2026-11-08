import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { OrderingContent } from "~/components/study/question-types/ordering/answer";
import i18n from "~/i18n/config";
import type { Question } from "~/lib/api/question.server";
import { SAME_LANGUAGE_ERROR_KEY } from "../lang-pair";
import { QUESTION_INVALID_ERROR_KEY, type QuestionEditFetcher } from "../question-form-error";
import { OrderingEditForm } from "./ordering-edit-form";

const fetcher = { state: "idle", Form: "form" } as unknown as QuestionEditFetcher;

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

function renderForm(parsed: OrderingContent | null = PARSED, formFetcher = fetcher) {
  return render(
    <OrderingEditForm
      question={QUESTION}
      parsed={parsed}
      fetcher={formFetcher}
      onCancel={vi.fn()}
    />,
  );
}

afterEach(() => {
  vi.restoreAllMocks();
});

function saveButton() {
  return screen.getByRole("button", { name: i18n.t("common.save") });
}

describe("OrderingEditForm", () => {
  it("should disable the save button when the target language equals the source language", async () => {
    // given
    const user = userEvent.setup();
    renderForm();

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "ko");

    // then
    expect(saveButton()).toBeDisabled();
  });

  it("should show the same language error when the source language equals the target language", async () => {
    // given
    const user = userEvent.setup();
    renderForm();

    // when
    await user.selectOptions(screen.getByLabelText("Source Language"), "fr");

    // then
    expect(screen.getByText(i18n.t(SAME_LANGUAGE_ERROR_KEY))).toBeInTheDocument();
  });

  it("should mark the target language invalid when it equals the source language", async () => {
    // given
    const user = userEvent.setup();
    renderForm();

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "ko");

    // then
    expect(screen.getByLabelText("Target Language")).toHaveAttribute("aria-invalid", "true");
  });

  it("should describe the target language with the same language error when it equals the source language", async () => {
    // given
    const user = userEvent.setup();
    renderForm();

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "ko");

    // then
    expect(screen.getByLabelText("Target Language")).toHaveAccessibleDescription(
      i18n.t(SAME_LANGUAGE_ERROR_KEY),
    );
  });

  it("should enable the save button when the target language differs from the source language again", async () => {
    // given
    const user = userEvent.setup();
    renderForm();
    await user.selectOptions(screen.getByLabelText("Target Language"), "ko");

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "en");

    // then
    expect(saveButton()).toBeEnabled();
  });

  it("should disable the save button on open when the saved languages are the same", () => {
    // given
    const parsed: OrderingContent = { ...PARSED, target: { ...PARSED.target, lang: "ko" } };

    // when
    renderForm(parsed);

    // then
    expect(saveButton()).toBeDisabled();
  });

  it("should capitalize the preview in the selected target language", async () => {
    // given
    const user = userEvent.setup();
    renderForm();
    const toLocaleUpperCase = vi.spyOn(String.prototype, "toLocaleUpperCase");

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "de");

    // then
    expect(toLocaleUpperCase).toHaveBeenCalledWith("de");
  });

  it("should show the error key returned by the action", () => {
    // given
    const rejected = {
      ...fetcher,
      data: { ok: false, errorKey: QUESTION_INVALID_ERROR_KEY },
    } as unknown as QuestionEditFetcher;

    // when
    renderForm(PARSED, rejected);

    // then
    expect(screen.getByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).toBeInTheDocument();
  });

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
