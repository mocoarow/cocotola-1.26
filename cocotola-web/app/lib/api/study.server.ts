import { z } from "zod";

import type { AnswerPayload } from "~/lib/study/answer-payload";
import type { BackendClient } from "./client.server";

const studyQuestionAudioRefSchema = z.object({
  url: z.string(),
  durationSec: z.number(),
});

export type StudyQuestionAudioRef = z.infer<typeof studyQuestionAudioRefSchema>;

const studyQuestionAudioSchema = z.object({
  source: studyQuestionAudioRefSchema.optional(),
  target: studyQuestionAudioRefSchema.optional(),
});

export type StudyQuestionAudio = z.infer<typeof studyQuestionAudioSchema>;

const studyQuestionSchema = z.object({
  questionId: z.string(),
  questionType: z.string(),
  content: z.string(),
  tags: z.array(z.string()).optional(),
  orderIndex: z.number().int(),
  /** An empty `url` means the audio has not been generated yet; treat it as missing. */
  audio: studyQuestionAudioSchema.optional(),
});

export type StudyQuestion = z.infer<typeof studyQuestionSchema>;

const getStudyQuestionsResponseSchema = z.object({
  questions: z.array(studyQuestionSchema),
  totalDue: z.number().int(),
  newCount: z.number().int(),
  reviewCount: z.number().int(),
});

export type GetStudyQuestionsResponse = z.infer<typeof getStudyQuestionsResponseSchema>;

const studySummarySchema = z.object({
  newCount: z.number().int(),
  reviewCount: z.number().int(),
  totalDue: z.number().int(),
  reviewRatioNumerator: z.number().int(),
  reviewRatioDenominator: z.number().int(),
});

export type StudySummary = z.infer<typeof studySummarySchema>;

const recordAnswerResponseSchema = z.object({
  nextDueAt: z.string(),
  consecutiveCorrect: z.number().int(),
  totalCorrect: z.number().int(),
  totalIncorrect: z.number().int(),
});

export type RecordAnswerResponse = z.infer<typeof recordAnswerResponseSchema>;

const studyRecordSchema = z.object({
  workbookId: z.string(),
  questionId: z.string(),
  consecutiveCorrect: z.number().int(),
  lastAnsweredAt: z.string(),
  nextDueAt: z.string(),
  totalCorrect: z.number().int(),
  totalIncorrect: z.number().int(),
});

export type StudyRecord = z.infer<typeof studyRecordSchema>;

const listStudyRecordsResponseSchema = z.object({
  records: z.array(studyRecordSchema),
});

export type ListStudyRecordsResponse = z.infer<typeof listStudyRecordsResponseSchema>;

const parseStudyQuestions = (raw: unknown): GetStudyQuestionsResponse =>
  getStudyQuestionsResponseSchema.parse(raw);

const parseStudySummary = (raw: unknown): StudySummary => studySummarySchema.parse(raw);

const parseRecordAnswerResponse = (raw: unknown): RecordAnswerResponse =>
  recordAnswerResponseSchema.parse(raw);

const parseStudyRecords = (raw: unknown): ListStudyRecordsResponse =>
  listStudyRecordsResponseSchema.parse(raw);

function studyPath(workbookId: string): string {
  return `/api/v1/workbook/${encodeURIComponent(workbookId)}/study`;
}

export async function getStudyQuestions(
  client: BackendClient,
  workbookId: string,
  limit: number,
  practice = false,
  excludeIds: readonly string[] = [],
): Promise<GetStudyQuestionsResponse> {
  const clampedLimit = Math.max(1, Math.min(100, Math.floor(limit)));
  const params = new URLSearchParams({ limit: String(clampedLimit) });
  if (practice) params.set("practice", "true");
  for (const id of excludeIds) {
    if (id !== "") params.append("excludeIds", id);
  }

  return client.get("question", `${studyPath(workbookId)}?${params.toString()}`, {
    source: "study:getStudyQuestions",
    errorMessage: "Failed to fetch study questions",
    parse: parseStudyQuestions,
  });
}

export async function getStudySummary(
  client: BackendClient,
  workbookId: string,
  practice = false,
): Promise<StudySummary> {
  const params = new URLSearchParams();
  if (practice) params.set("practice", "true");
  const query = params.toString();

  return client.get("question", `${studyPath(workbookId)}/summary${query ? `?${query}` : ""}`, {
    source: "study:getStudySummary",
    errorMessage: "Failed to fetch study summary",
    parse: parseStudySummary,
  });
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

function toLocaleHeaders(locale?: AnswerLocaleHeaders): Record<string, string> {
  if (!locale || locale.localDateKey === "" || locale.timezone === "") return {};
  return { "X-Local-Date": locale.localDateKey, "X-Local-Timezone": locale.timezone };
}

function toRecordAnswerBody(payload: AnswerPayload): Record<string, unknown> {
  switch (payload.type) {
    case "multiple_choice":
      return { selectedChoiceIds: payload.selectedChoiceIds };
    case "word_fill":
      return { correct: payload.correct };
    case "ordering":
      return { regions: payload.regions };
    default: {
      const _exhaustive: never = payload;
      throw new Error("unsupported answer payload type");
    }
  }
}

export async function recordAnswer(
  client: BackendClient,
  workbookId: string,
  questionId: string,
  payload: AnswerPayload,
  locale?: AnswerLocaleHeaders,
): Promise<RecordAnswerResponse> {
  return client.post(
    "question",
    `${studyPath(workbookId)}/${encodeURIComponent(questionId)}/answer`,
    {
      source: "study:recordAnswer",
      errorMessage: "Failed to record answer",
      headers: toLocaleHeaders(locale),
      body: toRecordAnswerBody(payload),
      parse: parseRecordAnswerResponse,
    },
  );
}

export async function deleteStudyHistory(client: BackendClient, workbookId: string): Promise<void> {
  await client.delete("question", studyPath(workbookId), {
    source: "study:deleteStudyHistory",
    errorMessage: "Failed to delete study history",
  });
}

export async function listStudyRecords(
  client: BackendClient,
  workbookId: string,
): Promise<ListStudyRecordsResponse> {
  return client.get("question", `${studyPath(workbookId)}/records`, {
    source: "study:listStudyRecords",
    errorMessage: "Failed to list study records",
    parse: parseStudyRecords,
  });
}
