import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

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
});
