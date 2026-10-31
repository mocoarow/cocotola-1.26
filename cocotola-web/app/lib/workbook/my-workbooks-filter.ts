import type { SharedWorkbook } from "~/lib/api/sharing.server";
import type { Workbook } from "~/lib/api/workbook.server";

export const MY_WORKBOOKS_TABS = ["all", "created", "added"] as const;

export type MyWorkbooksTab = (typeof MY_WORKBOOKS_TABS)[number];

export type MyWorkbookEntry =
  | { kind: "created"; workbook: Workbook }
  | { kind: "added"; workbook: SharedWorkbook };

export function parseMyWorkbooksTab(value: string | null): MyWorkbooksTab {
  return MY_WORKBOOKS_TABS.find((tab) => tab === value) ?? "all";
}

function sortByAddedAtDesc(added: readonly SharedWorkbook[]): SharedWorkbook[] {
  return [...added].sort((a, b) => Date.parse(b.addedAt) - Date.parse(a.addedAt));
}

export function selectMyWorkbooks(
  tab: MyWorkbooksTab,
  created: readonly Workbook[],
  added: readonly SharedWorkbook[],
): MyWorkbookEntry[] {
  const createdIds = new Set(created.map((workbook) => workbook.workbookId));
  const createdEntries: MyWorkbookEntry[] =
    tab === "added" ? [] : created.map((workbook) => ({ kind: "created", workbook }));
  const addedEntries: MyWorkbookEntry[] =
    tab === "created"
      ? []
      : sortByAddedAtDesc(added.filter((workbook) => !createdIds.has(workbook.workbookId))).map(
          (workbook) => ({ kind: "added", workbook }),
        );
  return [...createdEntries, ...addedEntries];
}
