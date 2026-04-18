import { fetchWithIdToken, getQuestionUrl } from "./fetch.server";

export type Question = {
  questionId: string;
  questionType: string;
  content: string;
  tags?: string[];
  orderIndex: number;
  createdAt: string;
  updatedAt: string;
};

type ListQuestionsResponse = {
  questions: Question[];
};

type AddQuestionBody = {
  questionType: string;
  content: string;
  tags?: string[];
  orderIndex: number;
};

export async function listQuestions(accessToken: string, workbookId: string): Promise<Question[]> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/question`;

  const response = await fetchWithIdToken(baseUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to fetch questions", { status: response.status });
  }

  const data = (await response.json()) as ListQuestionsResponse;
  return data.questions ?? [];
}

export async function addQuestion(
  accessToken: string,
  workbookId: string,
  body: AddQuestionBody,
): Promise<Question> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/question`;

  const response = await fetchWithIdToken(baseUrl, url, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Response(text || "Failed to add question", { status: response.status });
  }

  return (await response.json()) as Question;
}
