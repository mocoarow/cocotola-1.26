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
  const url = process.env.COCOTOLA_QUESTION_URL;
  if (!url) {
    throw new Error("COCOTOLA_QUESTION_URL environment variable is required");
  }
  return url;
}

export async function listWorkbooks(accessToken: string, spaceId: string): Promise<Workbook[]> {
  const url = getQuestionUrl();
  const response = await fetch(`${url}/api/v1/workbook?spaceId=${encodeURIComponent(spaceId)}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to fetch workbooks", { status: response.status });
  }

  const data = (await response.json()) as ListWorkbooksResponse;
  return data.workbooks ?? [];
}

export async function deleteWorkbook(accessToken: string, workbookId: string): Promise<void> {
  const url = getQuestionUrl();
  const response = await fetch(`${url}/api/v1/workbook/${encodeURIComponent(workbookId)}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to delete workbook", { status: response.status });
  }
}
