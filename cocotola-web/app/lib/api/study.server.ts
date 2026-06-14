import { fetchWithIdToken, getQuestionUrl } from "./fetch.server";

export type StudyQuestionAudioRef = {
  url: string;
  durationSec: number;
};

export type StudyQuestionAudio = {
  source?: StudyQuestionAudioRef;
  target?: StudyQuestionAudioRef;
};

export type StudyQuestion = {
  questionId: string;
  questionType: string;
  content: string;
  tags?: string[];
  orderIndex: number;
  /**
   * Audio assets for this question. The whole object — and individual
   * `source`/`target` keys — may be omitted when the audio batch has not yet
   * generated the files. Empty `url` strings should be treated the same as
   * "missing": the audio batch sometimes returns the object with empty subfields
   * during transitional states.
   */
  audio?: StudyQuestionAudio;
};

export type GetStudyQuestionsResponse = {
  questions: StudyQuestion[];
  totalDue: number;
  newCount: number;
  reviewCount: number;
};

export type StudySummary = {
  newCount: number;
  reviewCount: number;
  totalDue: number;
  reviewRatioNumerator: number;
  reviewRatioDenominator: number;
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
  excludeIds: readonly string[] = [],
): Promise<GetStudyQuestionsResponse> {
  const clampedLimit = Math.max(1, Math.min(100, Math.floor(limit)));
  const baseUrl = getQuestionUrl();
  const params = new URLSearchParams({ limit: String(clampedLimit) });
  if (practice) params.set("practice", "true");
  for (const id of excludeIds) {
    if (id !== "") params.append("excludeIds", id);
  }
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/study?${params.toString()}`;

  const response = await fetchWithIdToken(baseUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to fetch study questions", { status: response.status });
  }

  return (await response.json()) as GetStudyQuestionsResponse;
}

export async function getStudySummary(
  accessToken: string,
  workbookId: string,
  practice = false,
): Promise<StudySummary> {
  const baseUrl = getQuestionUrl();
  const params = new URLSearchParams();
  if (practice) params.set("practice", "true");
  const query = params.toString();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/study/summary${
    query ? `?${query}` : ""
  }`;

  const response = await fetchWithIdToken(baseUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to fetch study summary", { status: response.status });
  }

  return (await response.json()) as StudySummary;
}

/**
 * Locale headers sent with the answer POST so the backend can bucket the
 * answer into the dashboard contribution graph using the user's local
 * "today". Both fields are optional at the wire level — the backend will
 * still record the SRS answer if absent, but the daily-stats counter will
 * not advance.
 */
export type AnswerLocaleHeaders = {
  localDateKey: string;
  timezone: string;
};

async function postRecordAnswer(
  accessToken: string,
  workbookId: string,
  questionId: string,
  body: Record<string, unknown>,
  locale?: AnswerLocaleHeaders,
): Promise<RecordAnswerResponse> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/study/${encodeURIComponent(questionId)}/answer`;

  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
    "Content-Type": "application/json",
  };
  if (locale && locale.localDateKey !== "" && locale.timezone !== "") {
    headers["X-Local-Date"] = locale.localDateKey;
    headers["X-Local-Timezone"] = locale.timezone;
  }

  const response = await fetchWithIdToken(baseUrl, url, {
    method: "POST",
    headers,
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
  locale?: AnswerLocaleHeaders,
): Promise<RecordAnswerResponse> {
  return postRecordAnswer(accessToken, workbookId, questionId, { correct }, locale);
}

export function recordAnswerForMultipleChoice(
  accessToken: string,
  workbookId: string,
  questionId: string,
  selectedChoiceIds: string[],
  locale?: AnswerLocaleHeaders,
): Promise<RecordAnswerResponse> {
  return postRecordAnswer(accessToken, workbookId, questionId, { selectedChoiceIds }, locale);
}

export async function deleteStudyHistory(accessToken: string, workbookId: string): Promise<void> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/study`;

  const response = await fetchWithIdToken(baseUrl, url, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to delete study history", { status: response.status });
  }
}

export type StudyRecord = {
  workbookId: string;
  questionId: string;
  consecutiveCorrect: number;
  lastAnsweredAt: string;
  nextDueAt: string;
  totalCorrect: number;
  totalIncorrect: number;
};

export type ListStudyRecordsResponse = {
  records: StudyRecord[];
};

export async function listStudyRecords(
  accessToken: string,
  workbookId: string,
): Promise<ListStudyRecordsResponse> {
  const baseUrl = getQuestionUrl();
  const url = `${baseUrl}/api/v1/workbook/${encodeURIComponent(workbookId)}/study/records`;

  const response = await fetchWithIdToken(baseUrl, url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!response.ok) {
    throw new Response("Failed to list study records", { status: response.status });
  }

  return (await response.json()) as ListStudyRecordsResponse;
}
