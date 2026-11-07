import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import "~/i18n/config";

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useFetcher: vi.fn(() => ({ state: "idle", Form: "form", submit: vi.fn() })),
  };
});

import { useFetcher } from "react-router";
import { OrderingAddForm } from "./ordering-add-form";

async function renderWithNotation(notation: string) {
  const user = userEvent.setup();
  const view = render(<OrderingAddForm onAdded={vi.fn()} />);
  await user.click(screen.getByLabelText("Correct sentence"));
  await user.paste(notation);
  return { ...view, user };
}

function idleFetcher() {
  return { state: "idle", Form: "form", submit: vi.fn() } as unknown as ReturnType<
    typeof useFetcher
  >;
}

beforeEach(() => {
  vi.mocked(useFetcher).mockImplementation(idleFetcher);
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("OrderingAddForm", () => {
  it("should submit with the addOrdering intent", () => {
    // given / when
    const { container } = render(<OrderingAddForm onAdded={vi.fn()} />);

    // then
    expect(container.querySelector('input[name="intent"]')).toHaveValue("addOrdering");
  });

  it("should describe the notation field with the error when the notation is invalid", async () => {
    // given / when
    await renderWithNotation("I / am .]]");

    // then
    expect(screen.getByLabelText("Correct sentence")).toHaveAccessibleDescription(
      "]] has no matching [[",
    );
  });

  it("should mark the notation field invalid when the notation is invalid", async () => {
    // given / when
    await renderWithNotation("I / am .]]");

    // then
    expect(screen.getByLabelText("Correct sentence")).toHaveAttribute("aria-invalid", "true");
  });

  it("should disable the submit button when the notation is invalid", async () => {
    // given / when
    await renderWithNotation("I / am .]]");

    // then
    expect(screen.getByRole("button", { name: "Add Question" })).toBeDisabled();
  });

  it("should show the preview with capitalized chips when the notation is valid", async () => {
    // given / when
    await renderWithNotation("i / am [[.]]");

    // then
    expect(screen.getByRole("list", { name: "Preview" })).toHaveTextContent("Iam.");
  });

  it("should capitalize the preview in the selected target language", async () => {
    // given
    const { user } = await renderWithNotation("i / am [[.]]");
    const toLocaleUpperCase = vi.spyOn(String.prototype, "toLocaleUpperCase");

    // when
    await user.selectOptions(screen.getByLabelText("Target Language"), "fr");

    // then
    expect(toLocaleUpperCase).toHaveBeenCalledWith("fr");
  });

  it("should show the capitalization note", () => {
    // given / when
    render(<OrderingAddForm onAdded={vi.fn()} />);

    // then
    expect(screen.getByText(/shown capitalized on the study screen/)).toBeInTheDocument();
  });

  it("should call onAdded when the question has been added", () => {
    // given
    vi.mocked(useFetcher).mockImplementation(
      () =>
        ({
          state: "idle",
          Form: "form",
          submit: vi.fn(),
          data: { ok: true, added: true },
        }) as unknown as ReturnType<typeof useFetcher>,
    );
    const onAdded = vi.fn();

    // when
    render(<OrderingAddForm onAdded={onAdded} />);

    // then
    expect(onAdded).toHaveBeenCalledTimes(1);
  });
});
