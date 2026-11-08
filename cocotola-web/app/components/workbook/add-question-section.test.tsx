import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import i18n from "~/i18n/config";

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useFetcher: vi.fn(() => ({ state: "idle", Form: "form", submit: vi.fn() })),
  };
});

import { AddQuestionSection } from "./add-question-section";

describe("AddQuestionSection", () => {
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
