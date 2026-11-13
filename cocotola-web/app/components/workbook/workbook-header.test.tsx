import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRoutesStub } from "react-router";
import { describe, expect, it, vi } from "vitest";
import i18n from "~/i18n/config";
import { WorkbookHeader } from "./workbook-header";

function renderHeader(saveAction: () => unknown = () => ({ ok: true })) {
  const Stub = createRoutesStub([
    {
      path: "/workbooks/:workbookId",
      Component: () => (
        <WorkbookHeader title="Phrasal verbs" description="" visibility="private" language="fr" />
      ),
      HydrateFallback: () => null,
      action: saveAction,
    },
  ]);
  render(<Stub initialEntries={["/workbooks/wb-1"]} />);
  return userEvent.setup();
}

async function submitSave(user: ReturnType<typeof userEvent.setup>) {
  await user.click(
    await screen.findByRole("button", { name: i18n.t("workbooks.detail.editWorkbook") }),
  );
  await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));
}

async function saveOnce(user: ReturnType<typeof userEvent.setup>) {
  await submitSave(user);
  await vi.waitFor(() =>
    expect(screen.queryByRole("button", { name: i18n.t("common.save") })).not.toBeInTheDocument(),
  );
}

describe("WorkbookHeader", () => {
  it("should leave edit mode when the save succeeds", async () => {
    // given
    const user = renderHeader();

    // when
    await submitSave(user);

    // then
    await vi.waitFor(() =>
      expect(screen.queryByRole("button", { name: i18n.t("common.save") })).not.toBeInTheDocument(),
    );
  });

  it("should stay in edit mode when the save is rejected", async () => {
    // given
    const user = renderHeader(() => ({ ok: false, errorKey: "workbooks.detail.errors.invalid" }));

    // when
    await submitSave(user);

    // then
    expect(await screen.findByText(i18n.t("workbooks.detail.errors.invalid"))).toBeInTheDocument();
  });

  it("should stay in edit mode when editing again after a save", async () => {
    // given
    const user = renderHeader();
    await saveOnce(user);

    // when
    await user.click(screen.getByRole("button", { name: i18n.t("workbooks.detail.editWorkbook") }));

    // then
    expect(screen.getByRole("button", { name: i18n.t("common.save") })).toBeInTheDocument();
  });
});
