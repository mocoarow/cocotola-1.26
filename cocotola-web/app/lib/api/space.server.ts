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
  const authUrl = process.env.COCOTOLA_AUTH_URL;
  if (!authUrl) {
    throw new Error("COCOTOLA_AUTH_URL environment variable is required");
  }

  const url = `${authUrl}/api/v1/auth/space`;
  const response = await fetchWithIdToken(authUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[api] GET ${url} -> ${response.status}`);
    throw new Response("Failed to fetch spaces", { status: response.status });
  }

  const data = (await response.json()) as ListSpacesResponse;
  return data.spaces ?? [];
}

export async function findPrivateSpace(accessToken: string): Promise<Space | undefined> {
  const spaces = await listSpaces(accessToken);
  return spaces.find((s) => s.spaceType === "private" && !s.deleted);
}
