import { fetchWithIdToken, getQuestionUrl } from "./fetch.server";

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

export async function listWorkbooks(accessToken: string, spaceId: string): Promise<Workbook[]> {
  console.info(`[workbook] listWorkbooks called: spaceId=${spaceId}`);

  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook?spaceId=${encodeURIComponent(spaceId)}`;
  console.info(`[workbook] fetching workbooks: url=${url}`);

  const response = await fetchWithIdToken(baseUrl, url, {
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

export async function createWorkbook(
  accessToken: string,
  data: { spaceId: string; title: string; description: string; visibility: "private" | "public" },
): Promise<Workbook> {
  console.info(`[workbook] createWorkbook called: spaceId=${data.spaceId}, title=${data.title}`);

  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook`;
  console.info(`[workbook] creating workbook: url=${url}`);

  const response = await fetchWithIdToken(baseUrl, url, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    console.error(`[workbook] createWorkbook failed: status=${response.status}, url=${url}`);
    throw new Response("Failed to create workbook", { status: response.status });
  }

  const workbook = (await response.json()) as Workbook;
  console.info(`[workbook] createWorkbook succeeded: workbookId=${workbook.workbookId}`);
  return workbook;
}

export async function getWorkbook(accessToken: string, workbookId: string): Promise<Workbook> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}`;

  const response = await fetchWithIdToken(baseUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to fetch workbook", { status: response.status });
  }

  return (await response.json()) as Workbook;
}

export async function updateWorkbook(
  accessToken: string,
  workbookId: string,
  data: { title: string; description: string; visibility: "private" | "public" },
): Promise<Workbook> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}`;

  const response = await fetchWithIdToken(baseUrl, url, {
    method: "PUT",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    throw new Response("Failed to update workbook", { status: response.status });
  }

  return (await response.json()) as Workbook;
}

export async function deleteWorkbook(accessToken: string, workbookId: string): Promise<void> {
  console.info(`[workbook] deleteWorkbook called: workbookId=${workbookId}`);

  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}`;
  console.info(`[workbook] deleting workbook: url=${url}`);

  const response = await fetchWithIdToken(baseUrl, url, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[workbook] deleteWorkbook failed: status=${response.status}, url=${url}`);
    throw new Response("Failed to delete workbook", { status: response.status });
  }

  console.info(`[workbook] deleteWorkbook succeeded: workbookId=${workbookId}`);
}
