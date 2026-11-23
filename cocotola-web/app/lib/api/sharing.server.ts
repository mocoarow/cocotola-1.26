import { z } from "zod";

import type { BackendClient } from "./client.server";

const publicWorkbookSchema = z.object({
  workbookId: z.string(),
  ownerId: z.string(),
  title: z.string(),
  description: z.string(),
  language: z.string(),
  createdAt: z.string(),
});

export type PublicWorkbook = z.infer<typeof publicWorkbookSchema>;

const sharedReferenceSchema = z.object({
  referenceId: z.string(),
  workbookId: z.string(),
  addedAt: z.string(),
});

export type SharedReference = z.infer<typeof sharedReferenceSchema>;

const sharedWorkbookSchema = sharedReferenceSchema.extend({
  title: z.string(),
  description: z.string(),
  language: z.string(),
});

export type SharedWorkbook = z.infer<typeof sharedWorkbookSchema>;

const listPublicResponseSchema = z.object({
  workbooks: z
    .array(publicWorkbookSchema)
    .nullish()
    .transform((workbooks) => workbooks ?? []),
});

const listSharedResponseSchema = z.object({
  references: z
    .array(sharedWorkbookSchema)
    .nullish()
    .transform((references) => references ?? []),
});

const parsePublicWorkbookList = (raw: unknown): PublicWorkbook[] =>
  listPublicResponseSchema.parse(raw).workbooks;

const parseSharedWorkbookList = (raw: unknown): SharedWorkbook[] =>
  listSharedResponseSchema.parse(raw).references;

const parseSharedReference = (raw: unknown): SharedReference => sharedReferenceSchema.parse(raw);

export async function listPublicWorkbooks(client: BackendClient): Promise<PublicWorkbook[]> {
  console.info("[sharing] listPublicWorkbooks called");

  const workbooks = await client.get("question", "/api/v1/workbook/public", {
    source: "sharing:listPublicWorkbooks",
    errorMessage: "Failed to fetch public workbooks",
    parse: parsePublicWorkbookList,
  });
  console.info(`[sharing] listPublicWorkbooks succeeded: count=${workbooks.length}`);
  return workbooks;
}

export async function listSharedWorkbooks(client: BackendClient): Promise<SharedWorkbook[]> {
  console.info("[sharing] listSharedWorkbooks called");

  const references = await client.get("question", "/api/v1/workbook/shared", {
    source: "sharing:listSharedWorkbooks",
    errorMessage: "Failed to fetch shared workbooks",
    parse: parseSharedWorkbookList,
  });
  console.info(`[sharing] listSharedWorkbooks succeeded: count=${references.length}`);
  return references;
}

export async function shareWorkbook(
  client: BackendClient,
  workbookId: string,
): Promise<SharedReference> {
  console.info(`[sharing] shareWorkbook called: workbookId=${workbookId}`);

  const path = `/api/v1/workbook/${encodeURIComponent(workbookId)}/share`;
  const reference = await client.post("question", path, {
    source: "sharing:shareWorkbook",
    errorMessage: "Failed to share workbook",
    parse: parseSharedReference,
  });
  console.info(`[sharing] shareWorkbook succeeded: referenceId=${reference.referenceId}`);
  return reference;
}

export async function unshareWorkbook(client: BackendClient, referenceId: string): Promise<void> {
  console.info(`[sharing] unshareWorkbook called: referenceId=${referenceId}`);

  const path = `/api/v1/workbook/shared/${encodeURIComponent(referenceId)}`;
  await client.delete("question", path, {
    source: "sharing:unshareWorkbook",
    errorMessage: "Failed to unshare workbook",
  });
  console.info(`[sharing] unshareWorkbook succeeded: referenceId=${referenceId}`);
}
