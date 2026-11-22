import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRoutesStub, data } from "react-router";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDialogProvider } from "~/components/confirm-dialog-provider";
import { QUESTION_INVALID_ERROR_KEY } from "~/components/workbook/question-form-error";
import i18n from "~/i18n/config";
import type { Question } from "~/lib/api/question.server";
import type { Workbook } from "~/lib/api/workbook.server";
import WorkbookDetail from "./workbooks.$workbookId";

const BOUNDARY_TEXT = "route error boundary";

const WORKBOOK: Workbook = {
  workbookId: "wb-1",
  spaceId: "space-1",
  ownerId: "user-1",
  organizationId: "org-1",
  title: "Phrasal verbs",
  description: "",
  visibility: "private",
  language: "fr",
  createdAt: "2026-11-01T00:00:00Z",
  updatedAt: "2026-11-01T00:00:00Z",
};

function wordFillQuestion(questionId: string, sourceText: string): Question {
  return {
    questionId,
    questionType: "word_fill",
    content: JSON.stringify({
      source: { text: sourceText, lang: "ja" },
      target: { text: "{{throw}} it {{away}}", lang: "en" },
    }),
    orderIndex: 0,
    createdAt: "2026-11-01T00:00:00Z",
    updatedAt: "2026-11-01T00:00:00Z",
  };
}

const QUESTIONS = [wordFillQuestion("q-1", "ゴミを捨てる"), wordFillQuestion("q-2", "電気を消す")];

function rejectAsInvalid() {
  return data({ ok: false, errorKey: QUESTION_INVALID_ERROR_KEY }, { status: 400 });
}

function revalidatedTitle(loadCount: number) {
  return `${WORKBOOK.title} ${loadCount}`;
}

function renderDetail(editAction: (args: { request: Request }) => unknown) {
  let loadCount = 0;
  const Stub = createRoutesStub([
    {
      path: "/workbooks/:workbookId",
      Component: WorkbookDetail,
      HydrateFallback: () => null,
      ErrorBoundary: () => <p>{BOUNDARY_TEXT}</p>,
      loader: () => {
        loadCount += 1;
        return {
          workbook: { ...WORKBOOK, title: revalidatedTitle(loadCount) },
          questions: QUESTIONS,
        };
      },
      action: editAction,
      shouldRevalidate: () => true,
    },
  ]);
  render(
    <ConfirmDialogProvider>
      <Stub initialEntries={["/workbooks/wb-1"]} />
    </ConfirmDialogProvider>,
  );
  return userEvent.setup();
}

async function openEditor(user: ReturnType<typeof userEvent.setup>, index: number) {
  const editButtons = await screen.findAllByRole("button", {
    name: i18n.t("workbooks.detail.editQuestion"),
  });
  await user.click(editButtons[index]);
}

async function failOnce(user: ReturnType<typeof userEvent.setup>, index: number) {
  await openEditor(user, index);
  await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));
  await screen.findByText(i18n.t(QUESTION_INVALID_ERROR_KEY));
  await user.click(screen.getByRole("button", { name: i18n.t("common.cancel") }));
}

function deferredAction() {
  let settle: (outcome: { value?: unknown; error?: unknown }) => void = () => {};
  const editAction = vi.fn(
    () =>
      new Promise((resolve, reject) => {
        settle = ({ value, error }) => (error === undefined ? resolve(value) : reject(error));
      }),
  );
  return { editAction, settle: (outcome: { value?: unknown; error?: unknown }) => settle(outcome) };
}

async function closeWhileSaving(
  user: ReturnType<typeof userEvent.setup>,
  editAction: ReturnType<typeof deferredAction>["editAction"],
) {
  await openEditor(user, 0);
  await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));
  await vi.waitFor(() => expect(editAction).toHaveBeenCalled());
  await user.click(screen.getByRole("button", { name: i18n.t("common.cancel") }));
}

async function saveSuccessfully(user: ReturnType<typeof userEvent.setup>) {
  await openEditor(user, 0);
  await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));
  await screen.findByText(revalidatedTitle(2));
}

describe("WorkbookDetail edit sheet", () => {
  it("should show the error key in the edit form when the input is rejected", async () => {
    // given
    const user = renderDetail(rejectAsInvalid);
    await openEditor(user, 0);

    // when
    await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));

    // then
    expect(await screen.findByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).toBeInTheDocument();
  });

  it("should not show the previous error when another question is opened", async () => {
    // given
    const user = renderDetail(rejectAsInvalid);
    await failOnce(user, 0);

    // when
    await openEditor(user, 1);

    // then
    expect(screen.queryByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).not.toBeInTheDocument();
  });

  it("should not show the previous error when the same question is reopened", async () => {
    // given
    const user = renderDetail(rejectAsInvalid);
    await failOnce(user, 0);

    // when
    await openEditor(user, 0);

    // then
    expect(screen.queryByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).not.toBeInTheDocument();
  });

  it("should not show the error that arrives after the sheet was closed while submitting", async () => {
    // given
    const { editAction, settle } = deferredAction();
    const user = renderDetail(editAction);
    await closeWhileSaving(user, editAction);
    await openEditor(user, 0);

    // when
    act(() => settle({ value: rejectAsInvalid() }));
    await screen.findByText(revalidatedTitle(2));

    // then
    expect(screen.queryByText(i18n.t(QUESTION_INVALID_ERROR_KEY))).not.toBeInTheDocument();
  });

  it("should show the route error boundary when the save fails after the sheet was closed", async () => {
    // given
    const { editAction, settle } = deferredAction();
    const user = renderDetail(editAction);
    await closeWhileSaving(user, editAction);

    // when
    act(() => settle({ error: new Response("Failed to update question", { status: 500 }) }));

    // then
    expect(await screen.findByText(BOUNDARY_TEXT)).toBeInTheDocument();
  });

  it("should close the edit sheet when the save succeeds", async () => {
    // given
    const user = renderDetail(() => ({ ok: true }));

    // when
    await saveSuccessfully(user);

    // then
    expect(screen.queryByRole("button", { name: i18n.t("common.save") })).not.toBeInTheDocument();
  });

  it("should keep the edit sheet open when the same question is reopened after a save", async () => {
    // given
    const user = renderDetail(() => ({ ok: true }));
    await saveSuccessfully(user);

    // when
    await openEditor(user, 0);

    // then
    expect(screen.getByRole("button", { name: i18n.t("common.save") })).toBeInTheDocument();
  });
});

describe("WorkbookDetail header", () => {
  it("should submit the workbook's language when the title is saved", async () => {
    // given
    let submitted: FormData | undefined;
    const user = renderDetail(async ({ request }) => {
      submitted = await request.formData();
      return { ok: true };
    });
    await user.click(
      await screen.findByRole("button", { name: i18n.t("workbooks.detail.editWorkbook") }),
    );

    // when
    await user.click(screen.getByRole("button", { name: i18n.t("common.save") }));

    // then
    await vi.waitFor(() => expect(submitted?.get("language")).toBe(WORKBOOK.language));
  });
});
