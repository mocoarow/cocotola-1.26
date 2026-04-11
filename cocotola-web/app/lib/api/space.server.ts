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

  const response = await fetch(`${authUrl}/api/v1/auth/space`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to fetch spaces", { status: response.status });
  }

  const data = (await response.json()) as ListSpacesResponse;
  return data.spaces ?? [];
}

export async function findPrivateSpace(accessToken: string): Promise<Space | undefined> {
  const spaces = await listSpaces(accessToken);
  return spaces.find((s) => s.spaceType === "private" && !s.deleted);
}
