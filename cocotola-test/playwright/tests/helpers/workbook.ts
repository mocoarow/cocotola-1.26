import { expect, type APIRequestContext } from "@playwright/test";
import { bearer } from "./auth";

export type SpaceType = "private" | "public";

type Space = {
  spaceId: string;
  spaceType: SpaceType;
};

type ListSpacesResponse = {
  spaces: Space[];
};

export type WorkbookVisibility = "private" | "public";

export type Workbook = {
  workbookId: string;
  title: string;
  visibility: WorkbookVisibility;
};

type ListWorkbooksResponse = {
  workbooks: Workbook[];
};

export type PublicWorkbook = {
  workbookId: string;
  title: string;
  description: string;
  language: string;
};

type ListPublicResponse = {
  workbooks: PublicWorkbook[];
};

type CreateWorkbookInput = {
  userToken: string;
  spaceId: string;
  title: string;
  description?: string;
  visibility: WorkbookVisibility;
  language: string;
};

type AddQuestionInput = {
  userToken: string;
  workbookId: string;
  questionType: string;
  content: string;
  orderIndex: number;
  tags?: string[];
};

export type StudyQuestion = {
  questionId: string;
  questionType: string;
  content: string;
  orderIndex: number;
  tags?: string[];
};

export type GetStudyQuestionsResponse = {
  questions: StudyQuestion[];
  totalDue: number;
  newCount: number;
  reviewCount: number;
};

export type RecordAnswerResponse = {
  nextDueAt: string;
  consecutiveCorrect: number;
  totalCorrect: number;
  totalIncorrect: number;
};

const SPACE_WAIT_TIMEOUT_MS = 2_500;
const SPACE_WAIT_INTERVAL_MS = 500;

export async function waitForSpaceByType(
  request: APIRequestContext,
  userToken: string,
  spaceType: SpaceType,
): Promise<string> {
  let spaceId: string | undefined;
  await expect
    .poll(
      async () => {
        const response = await request.get("/api/v1/auth/space", {
          headers: bearer(userToken),
        });
        if (response.status() !== 200) return undefined;
        const body = (await response.json()) as ListSpacesResponse;
        spaceId = body.spaces.find((s) => s.spaceType === spaceType)?.spaceId;
        return spaceId;
      },
      {
        timeout: SPACE_WAIT_TIMEOUT_MS,
        intervals: [SPACE_WAIT_INTERVAL_MS],
        message: `${spaceType} space not provisioned within wait window`,
      },
    )
    .toBeDefined();
  if (!spaceId) {
    throw new Error(`${spaceType} space not provisioned within wait window`);
  }
  return spaceId;
}

export async function createWorkbook(
  request: APIRequestContext,
  input: CreateWorkbookInput,
): Promise<Workbook> {
  const response = await request.post("/api/v1/workbook", {
    headers: bearer(input.userToken),
    data: {
      spaceId: input.spaceId,
      title: input.title,
      description: input.description ?? "",
      visibility: input.visibility,
      language: input.language,
    },
  });
  expect(response.status()).toBe(201);
  const body = (await response.json()) as Workbook;
  expect(body.workbookId).not.toBe("");
  expect(body.visibility).toBe(input.visibility);
  expect(body.title).toBe(input.title);
  return body;
}

export async function addQuestion(
  request: APIRequestContext,
  input: AddQuestionInput,
): Promise<string> {
  const response = await request.post(
    `/api/v1/workbook/${encodeURIComponent(input.workbookId)}/question`,
    {
      headers: bearer(input.userToken),
      data: {
        questionType: input.questionType,
        content: input.content,
        orderIndex: input.orderIndex,
        tags: input.tags ?? [],
      },
    },
  );
  expect(response.status()).toBe(201);
  const body = (await response.json()) as { questionId: string };
  expect(body.questionId).not.toBe("");
  return body.questionId;
}

export async function listWorkbooks(
  request: APIRequestContext,
  userToken: string,
  spaceId: string,
): Promise<Workbook[]> {
  const response = await request.get(
    `/api/v1/workbook?spaceId=${encodeURIComponent(spaceId)}`,
    { headers: bearer(userToken) },
  );
  expect(response.status()).toBe(200);
  const body = (await response.json()) as ListWorkbooksResponse;
  return body.workbooks;
}

export async function listPublicWorkbooks(
  request: APIRequestContext,
  userToken: string,
): Promise<PublicWorkbook[]> {
  const response = await request.get("/api/v1/workbook/public", {
    headers: bearer(userToken),
  });
  expect(response.status()).toBe(200);
  const body = (await response.json()) as ListPublicResponse;
  return body.workbooks;
}

export async function getStudyQuestions(
  request: APIRequestContext,
  userToken: string,
  workbookId: string,
  limit: number,
): Promise<GetStudyQuestionsResponse> {
  const response = await request.get(
    `/api/v1/workbook/${encodeURIComponent(workbookId)}/study?limit=${limit}`,
    { headers: bearer(userToken) },
  );
  expect(response.status()).toBe(200);
  return (await response.json()) as GetStudyQuestionsResponse;
}

export async function recordAnswer(
  request: APIRequestContext,
  userToken: string,
  workbookId: string,
  questionId: string,
  correct: boolean,
): Promise<RecordAnswerResponse> {
  const response = await request.post(
    `/api/v1/workbook/${encodeURIComponent(workbookId)}/study/${encodeURIComponent(questionId)}/answer`,
    {
      headers: bearer(userToken),
      data: { correct },
    },
  );
  expect(response.status()).toBe(200);
  return (await response.json()) as RecordAnswerResponse;
}
