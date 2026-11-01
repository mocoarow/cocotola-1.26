import type { ComponentType } from "react";
import type { QuestionType } from "~/components/workbook/schemas";
import type { StudyQuestionAudio } from "~/lib/api/study.server";
import type { AnswerPayload } from "~/lib/study/answer-payload";

export type ShellPhase = "answering" | "revealed";

export type AnswerDraft<T extends QuestionType = QuestionType> = {
  correct: boolean;
  payload: Extract<AnswerPayload, { type: T }>;
};

export type BodyProps<T extends QuestionType, C> = {
  content: C;
  audio?: StudyQuestionAudio;
  phase: ShellPhase;
  onDraftChange: (draft: AnswerDraft<T> | null) => void;
  onCommit: (draft: AnswerDraft<T>) => void;
};

export type QuestionModule<T extends QuestionType, C> = {
  parse: (content: string) => C | null;
  initialDraft: (content: C) => AnswerDraft<T> | null;
  revealLabelKey: "workbooks.study.check" | "workbooks.study.showAnswer";
  explanation: (content: C, phase: ShellPhase) => string | undefined;
  Body: ComponentType<BodyProps<T, C>>;
};
