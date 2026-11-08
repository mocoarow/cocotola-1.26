import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import i18n from "~/i18n/config";

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useFetcher: vi.fn(() => ({ state: "idle", Form: "form", submit: vi.fn() })),
  };
});

import { useFetcher } from "react-router";
import { AddQuestionSection } from "./add-question-section";
import { SAME_LANGUAGE_ERROR_KEY } from "./lang-pair";
import { QUESTION_INVALID_ERROR_KEY } from "./question-form-error";

function idleFetcher(data?: unknown) {
  return { state: "idle", Form: "form", submit: vi.fn(), data } as unknown as ReturnType<
    typeof useFetcher
  >;
}

async function openWordFillForm() {
  const user = userEvent.setup();
  render(<AddQuestionSection />);
  await user.click(screen.getByRole("button", { name: "Add Question" }));
  await user.click(screen.getByRole("button", { name: /^Word Fill/ }));
  return user;
}

function submitButton() {
  return screen.getByRole("button", { name: i18n.t("workbooks.addQuestion.submit") });
}

beforeEach(() => {
  vi.mocked(useFetcher).mockImplementation(() => idleFetcher());
});

afterEach(() => {
  vi.resetAllMocks();
});

describe("AddQuestionSection", () => {
  it.each([
    { type: "word fill", button: /^Word Fill/ },
    { type: "multiple choice", button: /^Multiple Choice/ },
  ])("should show the error key returned for the $type form", async ({ button }) => {
    // given
    vi.mocked(useFetcher).mockImplementation(() =>
      idleFetcher({ ok: false, errorKey: QUESTION_INVALID_ERROR_KEY }),
    );
    const user = userEvent.setup();
    render(<AddQuestionSection />);
    await user.click(screen.getByRole("button", { name: "Add Question" }));

    // when
    await user.click(screen.getByRole("button", { name: button }));

    // then
    expect(screen.getByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).toBeInTheDocument();
  });

  it("should show the ordering form when the ordering type is selected", async () => {
    // given
    const user = userEvent.setup();
    render(<AddQuestionSection />);
    await user.click(screen.getByRole("button", { name: "Add Question" }));

    // when
    await user.click(screen.getByRole("button", { name: /^Ordering/ }));

    // then
    expect(screen.getByLabelText("Correct sentence")).toBeInTheDocument();
  });

  it("should default the word fill source language to ja", async () => {
    // given
    const user = userEvent.setup();
    render(<AddQuestionSection />);
    await user.click(screen.getByRole("button", { name: "Add Question" }));

    // when
    await user.click(screen.getByRole("button", { name: /^Word Fill/ }));

    // then
    expect(screen.getByLabelText("Source Language")).toHaveValue("ja");
  });

  it("should default the word fill target language to en", async () => {
    // given
    const user = userEvent.setup();
    render(<AddQuestionSection />);
    await user.click(screen.getByRole("button", { name: "Add Question" }));

    // when
    await user.click(screen.getByRole("button", { name: /^Word Fill/ }));

    // then
    expect(screen.getByLabelText("Target Language")).toHaveValue("en");
  });

  it("should describe the word fill target text with the hint", async () => {
    // given
    const user = userEvent.setup();
    render(<AddQuestionSection />);
    await user.click(screen.getByRole("button", { name: "Add Question" }));

    // when
    await user.click(screen.getByRole("button", { name: /^Word Fill/ }));

    // then
    expect(screen.getByLabelText("Target Text")).toHaveAccessibleDescription(
      i18n.t("workbooks.addQuestion.targetHint"),
    );
  });

  it("should describe the word fill tags with the hint", async () => {
    // given
    const user = userEvent.setup();
    render(<AddQuestionSection />);
    await user.click(screen.getByRole("button", { name: "Add Question" }));

    // when
    await user.click(screen.getByRole("button", { name: /^Word Fill/ }));

    // then
    expect(screen.getByLabelText("Tags (optional)")).toHaveAccessibleDescription(
      i18n.t("workbooks.addQuestion.tagsHint"),
    );
  });

  it("should disable the word fill submit button when the target language equals the source language", async () => {
    // given
    const user = await openWordFillForm();

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "ja");

    // then
    expect(submitButton()).toBeDisabled();
  });

  it("should show the same language error when the word fill target language equals the source language", async () => {
    // given
    const user = await openWordFillForm();

    // when
    await user.selectOptions(screen.getByLabelText("Source Language"), "en");

    // then
    expect(screen.getByText(i18n.t(SAME_LANGUAGE_ERROR_KEY))).toBeInTheDocument();
  });

  it("should mark the word fill target language invalid when it equals the source language", async () => {
    // given
    const user = await openWordFillForm();

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "ja");

    // then
    expect(screen.getByLabelText("Target Language")).toHaveAttribute("aria-invalid", "true");
  });

  it("should describe the word fill target language with the same language error when it equals the source language", async () => {
    // given
    const user = await openWordFillForm();

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "ja");

    // then
    expect(screen.getByLabelText("Target Language")).toHaveAccessibleDescription(
      i18n.t(SAME_LANGUAGE_ERROR_KEY),
    );
  });

  it("should enable the word fill submit button when the target language differs from the source language again", async () => {
    // given
    const user = await openWordFillForm();
    await user.selectOptions(screen.getByLabelText("Target Language"), "ja");

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "fr");

    // then
    expect(submitButton()).toBeEnabled();
  });

  it("should hide the same language error when the word fill target language differs from the source language again", async () => {
    // given
    const user = await openWordFillForm();
    await user.selectOptions(screen.getByLabelText("Target Language"), "ja");

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "fr");

    // then
    expect(screen.queryByText(i18n.t(SAME_LANGUAGE_ERROR_KEY))).not.toBeInTheDocument();
  });

  it("should describe the multiple choice tags with the hint", async () => {
    // given
    const user = userEvent.setup();
    render(<AddQuestionSection />);
    await user.click(screen.getByRole("button", { name: "Add Question" }));

    // when
    await user.click(screen.getByRole("button", { name: /^Multiple Choice/ }));

    // then
    expect(screen.getByLabelText("Tags (optional)")).toHaveAccessibleDescription(
      i18n.t("workbooks.addQuestion.tagsHint"),
    );
  });
});
