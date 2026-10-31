import { describe, expect, it } from "vitest";
import type { SharedWorkbook } from "~/lib/api/sharing.server";
import type { Workbook } from "~/lib/api/workbook.server";
import { parseMyWorkbooksTab, selectMyWorkbooks } from "./my-workbooks-filter";

function buildCreated(workbookId: string): Workbook {
  return {
    workbookId,
    spaceId: "sp-1",
    ownerId: "user-1",
    organizationId: "org-1",
    title: `Created ${workbookId}`,
    description: "",
    visibility: "private",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  };
}

function buildAdded(workbookId: string, addedAt: string): SharedWorkbook {
  return {
    referenceId: `ref-${workbookId}`,
    workbookId,
    title: `Added ${workbookId}`,
    description: "",
    language: "en",
    addedAt,
  };
}

const created = [buildCreated("c-1")];
const added = [buildAdded("a-1", "2026-03-01T00:00:00Z")];

function ids(tab: Parameters<typeof selectMyWorkbooks>[0]): string[] {
  return selectMyWorkbooks(tab, created, added).map((entry) => entry.workbook.workbookId);
}

describe("parseMyWorkbooksTab", () => {
  it.each([
    ["all", "all"],
    ["created", "created"],
    ["added", "added"],
  ] as const)("should return %s when the value is %s", (value, expected) => {
    // given / when
    const tab = parseMyWorkbooksTab(value);

    // then
    expect(tab).toBe(expected);
  });

  it.each([null, "", "unknown"])("should fall back to all when the value is %j", (value) => {
    // given / when
    const tab = parseMyWorkbooksTab(value);

    // then
    expect(tab).toBe("all");
  });
});

describe("selectMyWorkbooks", () => {
  it("should return only created workbooks when tab is created", () => {
    // given / when
    const result = ids("created");

    // then
    expect(result).toEqual(["c-1"]);
  });

  it("should return only added workbooks when tab is added", () => {
    // given / when
    const result = ids("added");

    // then
    expect(result).toEqual(["a-1"]);
  });

  it("should place created workbooks before added workbooks when tab is all", () => {
    // given / when
    const result = ids("all");

    // then
    expect(result).toEqual(["c-1", "a-1"]);
  });

  it("should exclude added workbooks that the user also created", () => {
    // given
    const ownAdded = buildAdded("c-1", "2026-03-01T00:00:00Z");

    // when
    const result = selectMyWorkbooks("all", created, [ownAdded]).map((e) => e.kind);

    // then
    expect(result).toEqual(["created"]);
  });

  it("should order added workbooks by addedAt descending", () => {
    // given
    const unordered = [
      buildAdded("old", "2026-01-01T00:00:00Z"),
      buildAdded("new", "2026-03-01T00:00:00Z"),
      buildAdded("mid", "2026-02-01T00:00:00Z"),
    ];

    // when
    const result = selectMyWorkbooks("added", [], unordered).map((e) => e.workbook.workbookId);

    // then
    expect(result).toEqual(["new", "mid", "old"]);
  });
});
