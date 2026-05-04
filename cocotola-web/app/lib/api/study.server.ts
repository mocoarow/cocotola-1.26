import { fetchWithIdToken, getQuestionUrl } from "./fetch.server";

export type StudyQuestion = {
  questionId: string;
  questionType: string;
  content: string;
  tags?: string[];
  orderIndex: number;
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

export async function getStudyQuestions(
  accessToken: string,
  workbookId: string,
  limit: number,
  practice = false,
): Promise<GetStudyQuestionsResponse> {
  const clampedLimit = Math.max(1, Math.min(100, Math.floor(limit)));
  const baseUrl = getQuestionUrl();
  const params = new URLSearchParams({ limit: String(clampedLimit) });
  if (practice) params.set("practice", "true");
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/study?${params.toString()}`;

  const response = await fetchWithIdToken(baseUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to fetch study questions", { status: response.status });
  }

  return (await response.json()) as GetStudyQuestionsResponse;
}

async function postRecordAnswer(
  accessToken: string,
  workbookId: string,
  questionId: string,
  body: Record<string, unknown>,
): Promise<RecordAnswerResponse> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/study/${encodeURIComponent(questionId)}/answer`;

  const response = await fetchWithIdToken(baseUrl, url, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });

  if (!response.ok) {
    throw new Response("Failed to record answer", { status: response.status });
  }

  return (await response.json()) as RecordAnswerResponse;
}

export function recordAnswerForWordFill(
  accessToken: string,
  workbookId: string,
  questionId: string,
  correct: boolean,
): Promise<RecordAnswerResponse> {
  return postRecordAnswer(accessToken, workbookId, questionId, { correct });
}

export function recordAnswerForMultipleChoice(
  accessToken: string,
  workbookId: string,
  questionId: string,
  selectedChoiceIds: string[],
): Promise<RecordAnswerResponse> {
  return postRecordAnswer(accessToken, workbookId, questionId, { selectedChoiceIds });
}
