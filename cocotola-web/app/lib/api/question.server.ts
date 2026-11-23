import { z } from "zod";

import type { BackendClient } from "./client.server";

const questionSchema = z.object({
  questionId: z.string(),
  questionType: z.string(),
  content: z.string(),
  tags: z.array(z.string()).optional(),
  orderIndex: z.number().int(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type Question = z.infer<typeof questionSchema>;

const listQuestionsResponseSchema = z.object({
  questions: z
    .array(questionSchema)
    .nullish()
    .transform((questions) => questions ?? []),
});

const parseQuestion = (raw: unknown): Question => questionSchema.parse(raw);

const parseQuestionList = (raw: unknown): Question[] =>
  listQuestionsResponseSchema.parse(raw).questions;

export class QuestionValidationError extends Error {
  constructor() {
    super("Question rejected as invalid");
    this.name = "QuestionValidationError";
  }
}

function questionSaveError(status: number): Error | undefined {
  return status === 400 ? new QuestionValidationError() : undefined;
}

function questionsPath(workbookId: string): string {
  return `/api/v1/workbook/${encodeURIComponent(workbookId)}/question`;
}

function questionPath(workbookId: string, questionId: string): string {
  return `${questionsPath(workbookId)}/${encodeURIComponent(questionId)}`;
}

type AddQuestionBody = {
  questionType: string;
  content: string;
  tags?: string[];
  orderIndex: number;
};

type UpdateQuestionBody = {
  content: string;
  tags?: string[];
  orderIndex: number;
};

export async function listQuestions(
  client: BackendClient,
  workbookId: string,
): Promise<Question[]> {
  return client.get("question", questionsPath(workbookId), {
    source: "question:listQuestions",
    errorMessage: "Failed to fetch questions",
    parse: parseQuestionList,
  });
}

export async function updateQuestion(
  client: BackendClient,
  workbookId: string,
  questionId: string,
  body: UpdateQuestionBody,
): Promise<Question> {
  return client.put("question", questionPath(workbookId, questionId), {
    source: "question:updateQuestion",
    errorMessage: "Failed to update question",
    body,
    mapError: questionSaveError,
    parse: parseQuestion,
  });
}

export async function deleteQuestion(
  client: BackendClient,
  workbookId: string,
  questionId: string,
): Promise<void> {
  await client.delete("question", questionPath(workbookId, questionId), {
    source: "question:deleteQuestion",
    errorMessage: "Failed to delete question",
  });
}

export async function addQuestion(
  client: BackendClient,
  workbookId: string,
  body: AddQuestionBody,
): Promise<Question> {
  return client.post("question", questionsPath(workbookId), {
    source: "question:addQuestion",
    errorMessage: "Failed to add question",
    body,
    mapError: questionSaveError,
    parse: parseQuestion,
  });
}
