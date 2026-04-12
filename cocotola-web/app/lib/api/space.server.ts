import { fetchWithIdToken } from "./fetch.server";

export type Space = {
  spaceId: string;
  organizationId: string;
  ownerId: string;
  keyName: string;
  name: string;
  spaceType: "public" | "private";
  deleted: boolean;
};

type ListSpacesResponse = {
  spaces: Space[];
};

export async function listSpaces(accessToken: string): Promise<Space[]> {
  console.info("[space] listSpaces called");

  const authUrl = process.env.AUTH_BASE_URL;
  if (!authUrl) {
    throw new Error("AUTH_BASE_URL environment variable is required");
  }

  const url = `${authUrl}/api/v1/auth/space`;
  console.info(`[space] fetching spaces: url=${url}`);

  const response = await fetchWithIdToken(authUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[space] listSpaces failed: status=${response.status}, url=${url}`);
    throw new Response("Failed to fetch spaces", { status: response.status });
  }

  const data = (await response.json()) as ListSpacesResponse;
  const spaces = data.spaces ?? [];
  console.info(`[space] listSpaces succeeded: count=${spaces.length}`);
  return spaces;
}

export async function findPrivateSpace(accessToken: string): Promise<Space | undefined> {
  console.info("[space] findPrivateSpace called");
  const spaces = await listSpaces(accessToken);
  const privateSpace = spaces.find((s) => s.spaceType === "private" && !s.deleted);
  console.info(`[space] findPrivateSpace result: found=${!!privateSpace}, spaceId=${privateSpace?.spaceId ?? "none"}`);
  return privateSpace;
}
