import { z } from "zod";

import type { BackendClient } from "./client.server";

const workbookSchema = z.object({
  workbookId: z.string(),
  spaceId: z.string(),
  ownerId: z.string(),
  organizationId: z.string(),
  title: z.string(),
  description: z.string(),
  visibility: z.enum(["private", "public"]),
  language: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type Workbook = z.infer<typeof workbookSchema>;

const listWorkbooksResponseSchema = z.object({
  workbooks: z
    .array(workbookSchema)
    .nullish()
    .transform((workbooks) => workbooks ?? []),
});

const parseWorkbook = (raw: unknown): Workbook => workbookSchema.parse(raw);

const parseWorkbookList = (raw: unknown): Workbook[] =>
  listWorkbooksResponseSchema.parse(raw).workbooks;

export class WorkbookValidationError extends Error {
  constructor() {
    super("Workbook rejected as invalid");
    this.name = "WorkbookValidationError";
  }
}

function workbookSaveError(status: number): Error | undefined {
  return status === 400 ? new WorkbookValidationError() : undefined;
}

function workbookPath(workbookId: string): string {
  return `/api/v1/workbook/${encodeURIComponent(workbookId)}`;
}

export async function listWorkbooks(client: BackendClient, spaceId: string): Promise<Workbook[]> {
  console.info(`[workbook] listWorkbooks called: spaceId=${spaceId}`);

  const path = `/api/v1/workbook?spaceId=${encodeURIComponent(spaceId)}`;
  const workbooks = await client.get("question", path, {
    source: "workbook:listWorkbooks",
    errorMessage: "Failed to fetch workbooks",
    parse: parseWorkbookList,
  });
  console.info(`[workbook] listWorkbooks succeeded: count=${workbooks.length}`);
  return workbooks;
}

export async function createWorkbook(
  client: BackendClient,
  data: {
    spaceId: string;
    title: string;
    description: string;
    visibility: "private" | "public";
    language: string;
  },
): Promise<Workbook> {
  console.info(
    `[workbook] createWorkbook called: spaceId=${data.spaceId}, titleLength=${data.title.length}`,
  );

  const workbook = await client.post("question", "/api/v1/workbook", {
    source: "workbook:createWorkbook",
    errorMessage: "Failed to create workbook",
    body: data,
    mapError: workbookSaveError,
    parse: parseWorkbook,
  });
  console.info(`[workbook] createWorkbook succeeded: workbookId=${workbook.workbookId}`);
  return workbook;
}

export async function getWorkbook(client: BackendClient, workbookId: string): Promise<Workbook> {
  return client.get("question", workbookPath(workbookId), {
    source: "workbook:getWorkbook",
    errorMessage: "Failed to fetch workbook",
    parse: parseWorkbook,
  });
}

export async function updateWorkbook(
  client: BackendClient,
  workbookId: string,
  data: { title: string; description: string; visibility: "private" | "public"; language: string },
): Promise<Workbook> {
  return client.put("question", workbookPath(workbookId), {
    source: "workbook:updateWorkbook",
    errorMessage: "Failed to update workbook",
    body: data,
    mapError: workbookSaveError,
    parse: parseWorkbook,
  });
}

export async function deleteWorkbook(client: BackendClient, workbookId: string): Promise<void> {
  console.info(`[workbook] deleteWorkbook called: workbookId=${workbookId}`);

  await client.delete("question", workbookPath(workbookId), {
    source: "workbook:deleteWorkbook",
    errorMessage: "Failed to delete workbook",
  });
  console.info(`[workbook] deleteWorkbook succeeded: workbookId=${workbookId}`);
}
