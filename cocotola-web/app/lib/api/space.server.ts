import { z } from "zod";

import type { BackendClient } from "./client.server";

const spaceSchema = z.object({
  spaceId: z.string(),
  organizationId: z.string(),
  ownerId: z.string(),
  keyName: z.string(),
  name: z.string(),
  spaceType: z.enum(["public", "private"]),
  deleted: z.boolean(),
});

export type Space = z.infer<typeof spaceSchema>;

const listSpacesResponseSchema = z.object({
  spaces: z
    .array(spaceSchema)
    .nullish()
    .transform((spaces) => spaces ?? []),
});

const parseSpaceList = (raw: unknown): Space[] => listSpacesResponseSchema.parse(raw).spaces;

export async function listSpaces(client: BackendClient): Promise<Space[]> {
  console.info("[space] listSpaces called");

  const spaces = await client.get("auth", "/api/v1/auth/space", {
    source: "space:listSpaces",
    errorMessage: "Failed to fetch spaces",
    parse: parseSpaceList,
  });
  console.info(`[space] listSpaces succeeded: count=${spaces.length}`);
  return spaces;
}

export async function findPrivateSpace(client: BackendClient): Promise<Space | undefined> {
  console.info("[space] findPrivateSpace called");
  const spaces = await listSpaces(client);
  const privateSpace = spaces.find((s) => s.spaceType === "private" && !s.deleted);
  console.info(
    `[space] findPrivateSpace result: found=${!!privateSpace}, spaceId=${privateSpace?.spaceId ?? "none"}`,
  );
  return privateSpace;
}
