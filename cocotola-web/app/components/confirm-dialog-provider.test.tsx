import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

import { ConfirmDialogProvider, useConfirm } from "./confirm-dialog-provider";

function ConfirmTrigger({
  onResult,
  options,
}: {
  onResult: (value: boolean) => void;
  options?: {
    title?: string;
    description?: string;
    confirmLabel?: string;
    cancelLabel?: string;
  };
}) {
  const confirm = useConfirm();
  return (
    <button
      type="button"
      onClick={async () => {
        const result = await confirm({
          title: options?.title ?? "Test Title",
          description: options?.description ?? "Test Description",
          confirmLabel: options?.confirmLabel,
          cancelLabel: options?.cancelLabel,
        });
        onResult(result);
      }}
    >
      open
    </button>
  );
}

describe("ConfirmDialogProvider", () => {
  it("should resolve true when confirm button is clicked", async () => {
    // given
    const onResult = vi.fn();
    const user = userEvent.setup();
    render(
      <ConfirmDialogProvider>
        <ConfirmTrigger onResult={onResult} />
      </ConfirmDialogProvider>,
    );

    // when
    await user.click(screen.getByText("open"));
    await user.click(screen.getByText("Confirm"));

    // then
    expect(onResult).toHaveBeenCalledWith(true);
  });

  it("should resolve false when cancel button is clicked", async () => {
    // given
    const onResult = vi.fn();
    const user = userEvent.setup();
    render(
      <ConfirmDialogProvider>
        <ConfirmTrigger onResult={onResult} />
      </ConfirmDialogProvider>,
    );

    // when
    await user.click(screen.getByText("open"));
    await user.click(screen.getByText("Cancel"));

    // then
    expect(onResult).toHaveBeenCalledWith(false);
  });

  it("should display custom title and description", async () => {
    // given
    const user = userEvent.setup();
    render(
      <ConfirmDialogProvider>
        <ConfirmTrigger
          onResult={vi.fn()}
          options={{ title: "Delete item", description: "Are you sure?" }}
        />
      </ConfirmDialogProvider>,
    );

    // when
    await user.click(screen.getByText("open"));

    // then
    expect(screen.getByText("Delete item")).toBeInTheDocument();
    expect(screen.getByText("Are you sure?")).toBeInTheDocument();
  });

  it("should use custom confirm and cancel labels when provided", async () => {
    // given
    const user = userEvent.setup();
    render(
      <ConfirmDialogProvider>
        <ConfirmTrigger
          onResult={vi.fn()}
          options={{ confirmLabel: "Yes, do it", cancelLabel: "Nope" }}
        />
      </ConfirmDialogProvider>,
    );

    // when
    await user.click(screen.getByText("open"));

    // then
    expect(screen.getByText("Yes, do it")).toBeInTheDocument();
    expect(screen.getByText("Nope")).toBeInTheDocument();
  });

  it("should warn and resolve false when called while another dialog is open", async () => {
    // given
    const onResult = vi.fn();
    const user = userEvent.setup();
    const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

    function DoubleTrigger() {
      const confirm = useConfirm();
      return (
        <button
          type="button"
          onClick={() => {
            void confirm({ title: "First", description: "" });
            void confirm({ title: "Second", description: "" }).then(onResult);
          }}
        >
          double
        </button>
      );
    }

    render(
      <ConfirmDialogProvider>
        <DoubleTrigger />
      </ConfirmDialogProvider>,
    );

    // when
    await user.click(screen.getByText("double"));

    // then
    expect(onResult).toHaveBeenCalledWith(false);
    expect(warnSpy).toHaveBeenCalled();
    warnSpy.mockRestore();
  });

  it("should throw when useConfirm is called outside provider", () => {
    // given
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

    function Lonely() {
      useConfirm();
      return null;
    }

    // when / then
    expect(() => render(<Lonely />)).toThrow(
      "useConfirm must be used within ConfirmDialogProvider",
    );
    errorSpy.mockRestore();
  });

  it("should allow opening another dialog after the previous one is resolved", async () => {
    // given
    const onResult = vi.fn();
    const user = userEvent.setup();
    render(
      <ConfirmDialogProvider>
        <ConfirmTrigger onResult={onResult} />
      </ConfirmDialogProvider>,
    );

    // when
    await user.click(screen.getByText("open"));
    await user.click(screen.getByText("Cancel"));
    await act(async () => {
      // allow close animation/state to settle
    });
    await user.click(screen.getByText("open"));
    await user.click(screen.getByText("Confirm"));

    // then
    expect(onResult).toHaveBeenNthCalledWith(1, false);
    expect(onResult).toHaveBeenNthCalledWith(2, true);
  });
});
