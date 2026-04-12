import { fetchWithIdToken } from "./fetch.server";

export type Workbook = {
  workbookId: string;
  spaceId: string;
  ownerId: string;
  organizationId: string;
  title: string;
  description: string;
  visibility: "private" | "public";
  createdAt: string;
  updatedAt: string;
};

type ListWorkbooksResponse = {
  workbooks: Workbook[];
};

function getQuestionUrl(): string {
  const url = process.env.QUESTION_BASE_URL;
  if (!url) {
    throw new Error("QUESTION_BASE_URL environment variable is required");
  }
  return url;
}

export async function listWorkbooks(accessToken: string, spaceId: string): Promise<Workbook[]> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook?spaceId=${encodeURIComponent(spaceId)}`;
  const response = await fetchWithIdToken(baseUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[api] GET ${url} -> ${response.status}`);
    throw new Response("Failed to fetch workbooks", { status: response.status });
  }

  const data = (await response.json()) as ListWorkbooksResponse;
  return data.workbooks ?? [];
}

export async function deleteWorkbook(accessToken: string, workbookId: string): Promise<void> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}`;
  const response = await fetchWithIdToken(baseUrl, url, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[api] DELETE ${url} -> ${response.status}`);
    throw new Response("Failed to delete workbook", { status: response.status });
  }
}
