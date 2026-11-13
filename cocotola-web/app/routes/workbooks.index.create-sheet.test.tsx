import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRoutesStub } from "react-router";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDialogProvider } from "~/components/confirm-dialog-provider";
import i18n from "~/i18n/config";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/space.server", () => ({
  listSpaces: vi.fn(),
  findPrivateSpace: vi.fn(),
}));

vi.mock("~/lib/api/workbook.server", async (importOriginal) => {
  const actual = await importOriginal<typeof import("~/lib/api/workbook.server")>();
  return {
    listWorkbooks: vi.fn(),
    createWorkbook: vi.fn(),
    getWorkbook: vi.fn(),
    updateWorkbook: vi.fn(),
    deleteWorkbook: vi.fn(),
    WorkbookValidationError: actual.WorkbookValidationError,
  };
});

vi.mock("~/lib/api/sharing.server", () => ({
  listPublicWorkbooks: vi.fn(),
  listSharedWorkbooks: vi.fn(),
  shareWorkbook: vi.fn(),
  unshareWorkbook: vi.fn(),
}));

import WorkbooksIndex from "./workbooks.index";

function renderIndex(createAction: () => unknown = () => ({ ok: true })) {
  const Stub = createRoutesStub([
    {
      path: "/workbooks",
      Component: WorkbooksIndex,
      HydrateFallback: () => null,
      loader: () => ({ workbooks: [], addedWorkbooks: [], spaceId: "sp-1" }),
      action: createAction,
    },
  ]);
  render(
    <ConfirmDialogProvider>
      <Stub initialEntries={["/workbooks"]} />
    </ConfirmDialogProvider>,
  );
  return userEvent.setup();
}

async function submitCreate(user: ReturnType<typeof userEvent.setup>) {
  await user.click(
    await screen.findByRole("button", { name: i18n.t("workbooks.index.newWorkbook") }),
  );
  await user.type(
    screen.getByLabelText(new RegExp(i18n.t("workbooks.index.titleLabel"))),
    "French verbs",
  );
  await user.click(screen.getByRole("button", { name: i18n.t("common.create") }));
}

async function createOnce(user: ReturnType<typeof userEvent.setup>) {
  await submitCreate(user);
  await vi.waitFor(() =>
    expect(screen.queryByText(i18n.t("workbooks.index.createTitle"))).not.toBeInTheDocument(),
  );
}

describe("WorkbooksIndex create sheet", () => {
  it("should close when a workbook was created", async () => {
    // given
    const user = renderIndex();

    // when
    await submitCreate(user);

    // then
    await vi.waitFor(() =>
      expect(screen.queryByText(i18n.t("workbooks.index.createTitle"))).not.toBeInTheDocument(),
    );
  });

  it("should stay open when the creation is rejected", async () => {
    // given
    const user = renderIndex(() => ({ ok: false, errorKey: "workbooks.index.errors.invalid" }));

    // when
    await submitCreate(user);

    // then
    expect(await screen.findByText(i18n.t("workbooks.index.errors.invalid"))).toBeInTheDocument();
  });

  it("should stay open when reopened after a workbook was created", async () => {
    // given
    const user = renderIndex();
    await createOnce(user);

    // when
    await user.click(screen.getByRole("button", { name: i18n.t("workbooks.index.newWorkbook") }));

    // then
    expect(await screen.findByText(i18n.t("workbooks.index.createTitle"))).toBeInTheDocument();
  });

  it("should close again when a second workbook was created", async () => {
    // given
    const user = renderIndex();
    await createOnce(user);

    // when
    await submitCreate(user);

    // then
    await vi.waitFor(() =>
      expect(screen.queryByText(i18n.t("workbooks.index.createTitle"))).not.toBeInTheDocument(),
    );
  });
});
