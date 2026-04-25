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
): Promise<GetStudyQuestionsResponse> {
  const clampedLimit = Math.max(1, Math.min(100, Math.floor(limit)));
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/study?limit=${clampedLimit}`;

  const response = await fetchWithIdToken(baseUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to fetch study questions", { status: response.status });
  }

  return (await response.json()) as GetStudyQuestionsResponse;
}

export async function recordAnswer(
  accessToken: string,
  workbookId: string,
  questionId: string,
  correct: boolean,
): Promise<RecordAnswerResponse> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/study/${encodeURIComponent(questionId)}/answer`;

  const response = await fetchWithIdToken(baseUrl, url, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ correct }),
  });

  if (!response.ok) {
    throw new Response("Failed to record answer", { status: response.status });
  }

  return (await response.json()) as RecordAnswerResponse;
}
