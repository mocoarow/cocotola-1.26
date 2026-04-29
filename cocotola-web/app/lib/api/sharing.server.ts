import { fetchWithIdToken, getQuestionUrl } from "./fetch.server";

export type PublicWorkbook = {
  workbookId: string;
  ownerId: string;
  title: string;
  description: string;
  language: string;
  createdAt: string;
};

export type SharedReference = {
  referenceId: string;
  workbookId: string;
  addedAt: string;
};

type ListPublicResponse = {
  workbooks: PublicWorkbook[];
};

type ListSharedResponse = {
  references: SharedReference[];
};

export async function listPublicWorkbooks(accessToken: string): Promise<PublicWorkbook[]> {
  console.info("[sharing] listPublicWorkbooks called");

  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/public`;
  console.info(`[sharing] fetching public workbooks: url=${url}`);

  const response = await fetchWithIdToken(baseUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[sharing] listPublicWorkbooks failed: status=${response.status}, url=${url}`);
    throw new Response("Failed to fetch public workbooks", { status: response.status });
  }

  const data = (await response.json()) as ListPublicResponse;
  const workbooks = data.workbooks ?? [];
  console.info(`[sharing] listPublicWorkbooks succeeded: count=${workbooks.length}`);
  return workbooks;
}

export async function listSharedWorkbooks(accessToken: string): Promise<SharedReference[]> {
  console.info("[sharing] listSharedWorkbooks called");

  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/shared`;
  console.info(`[sharing] fetching shared workbooks: url=${url}`);

  const response = await fetchWithIdToken(baseUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[sharing] listSharedWorkbooks failed: status=${response.status}, url=${url}`);
    throw new Response("Failed to fetch shared workbooks", { status: response.status });
  }

  const data = (await response.json()) as ListSharedResponse;
  const references = data.references ?? [];
  console.info(`[sharing] listSharedWorkbooks succeeded: count=${references.length}`);
  return references;
}

export async function shareWorkbook(
  accessToken: string,
  workbookId: string,
): Promise<SharedReference> {
  console.info(`[sharing] shareWorkbook called: workbookId=${workbookId}`);

  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/share`;
  console.info(`[sharing] sharing workbook: url=${url}`);

  const response = await fetchWithIdToken(baseUrl, url, {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[sharing] shareWorkbook failed: status=${response.status}, url=${url}`);
    throw new Response("Failed to share workbook", { status: response.status });
  }

  const reference = (await response.json()) as SharedReference;
  console.info(`[sharing] shareWorkbook succeeded: referenceId=${reference.referenceId}`);
  return reference;
}

export async function unshareWorkbook(accessToken: string, referenceId: string): Promise<void> {
  console.info(`[sharing] unshareWorkbook called: referenceId=${referenceId}`);

  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/shared/${encodeURIComponent(referenceId)}`;
  console.info(`[sharing] unsharing workbook: url=${url}`);

  const response = await fetchWithIdToken(baseUrl, url, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    console.error(`[sharing] unshareWorkbook failed: status=${response.status}, url=${url}`);
    throw new Response("Failed to unshare workbook", { status: response.status });
  }

  console.info(`[sharing] unshareWorkbook succeeded: referenceId=${referenceId}`);
}
