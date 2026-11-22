import { useState } from "react";

import type { AnswerDraft } from "~/components/study/question-shell/types";
import type { StudyQuestion } from "~/lib/api/study.server";
import { appendAnsweredId, clearStudySession } from "~/lib/study-session";

export type UseStudySessionControllerArgs = {
  userId: string | null;
  workbookId: string;
  practice: boolean;
  questions: StudyQuestion[];
  submitAnswer: (questionId: string, draft: AnswerDraft) => void;
};

export type StudySessionController = {
  /** undefined once every question has been answered correctly or skipped. */
  question: StudyQuestion | undefined;
  correctCount: number;
  incorrectCount: number;
  cardKey: string;
  answer: (draft: AnswerDraft) => void;
  skip: () => void;
};

/**
 * Drives one study session: a correct answer removes the head of the queue,
 * a wrong answer sends it to the back, and a skip removes it without counting.
 * The queue is seeded once from `questions`; callers remount to re-seed.
 */
export function useStudySessionController({
  userId,
  workbookId,
  practice,
  questions,
  submitAnswer,
}: UseStudySessionControllerArgs): StudySessionController {
  const [queue, setQueue] = useState<StudyQuestion[]>(() => questions);
  const [correctCount, setCorrectCount] = useState(0);
  const [incorrectCount, setIncorrectCount] = useState(0);
  const [attemptCounts, setAttemptCounts] = useState<Record<string, number>>({});

  const question = queue[0];
  // Skip the persistence side when userId is unknown so we never write under an ambiguous scope.
  const scope = userId === null ? null : { userId, workbookId, practice };

  function dequeueHead() {
    // Read the queue length here, BEFORE setQueue, so the "session done"
    // decision lives outside the setState updater. The updater itself must
    // stay pure — StrictMode runs it twice.
    const finishesSession = queue.length === 1;
    setQueue((q) => q.slice(1));
    if (scope !== null && finishesSession) {
      clearStudySession(scope);
    }
  }

  function answer(draft: AnswerDraft) {
    if (question === undefined) return;
    submitAnswer(question.questionId, draft);
    setAttemptCounts((prev) => ({
      ...prev,
      [question.questionId]: (prev[question.questionId] ?? 0) + 1,
    }));
    if (!draft.correct) {
      setIncorrectCount((c) => c + 1);
      setQueue(rotateHead);
      return;
    }
    // Only correct answers are persisted: a reload resumes past finished
    // questions, while a wrong one that was not retried yet reappears fresh.
    if (scope !== null) {
      appendAnsweredId(scope, question.questionId, Date.now());
    }
    setCorrectCount((c) => c + 1);
    dequeueHead();
  }

  const cardKey = cardKeyOf(question, attemptCounts);
  return { question, correctCount, incorrectCount, cardKey, answer, skip: dequeueHead };
}

function rotateHead(queue: StudyQuestion[]): StudyQuestion[] {
  const [head, ...rest] = queue;
  return [...rest, head];
}

function cardKeyOf(
  question: StudyQuestion | undefined,
  attemptCounts: Record<string, number>,
): string {
  if (question === undefined) return "";
  return `${question.questionId}-${attemptCounts[question.questionId] ?? 0}`;
}
