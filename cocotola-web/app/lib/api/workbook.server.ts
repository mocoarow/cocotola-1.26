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
  console.info(`[workbook] listWorkbooks called: spaceId=${spaceId}`);

  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook?spaceId=${encodeURIComponent(spaceId)}`;
  console.info(`[workbook] fetching workbooks: url=${url}`);

  const response = await fetchWithIdToken("cocotola-question", url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[workbook] listWorkbooks failed: status=${response.status}, url=${url}`);
    throw new Response("Failed to fetch workbooks", { status: response.status });
  }

  const data = (await response.json()) as ListWorkbooksResponse;
  const workbooks = data.workbooks ?? [];
  console.info(`[workbook] listWorkbooks succeeded: count=${workbooks.length}`);
  return workbooks;
}

export async function deleteWorkbook(accessToken: string, workbookId: string): Promise<void> {
  console.info(`[workbook] deleteWorkbook called: workbookId=${workbookId}`);

  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}`;
  console.info(`[workbook] deleting workbook: url=${url}`);

  const response = await fetchWithIdToken("cocotola-question", url, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[workbook] deleteWorkbook failed: status=${response.status}, url=${url}`);
    throw new Response("Failed to delete workbook", { status: response.status });
  }

  console.info(`[workbook] deleteWorkbook succeeded: workbookId=${workbookId}`);
}
