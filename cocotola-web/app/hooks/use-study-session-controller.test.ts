import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { AnswerDraft } from "~/components/study/question-shell/types";
import type { StudyQuestion } from "~/lib/api/study.server";
import {
  type StudySessionScope,
  saveStudySession,
  studySessionStorageKey,
} from "~/lib/study-session";
import { useStudySessionController } from "./use-study-session-controller";

const scope: StudySessionScope = { userId: "user-1", workbookId: "wb-1", practice: false };

const correctDraft: AnswerDraft = {
  correct: true,
  payload: { type: "multiple_choice", selectedChoiceIds: ["c-1"] },
};

const wrongDraft: AnswerDraft = {
  correct: false,
  payload: { type: "multiple_choice", selectedChoiceIds: ["c-2"] },
};

function questionFixture(questionId: string): StudyQuestion {
  return { questionId, questionType: "multiple_choice", content: "{}", orderIndex: 0 };
}

function renderController(
  questionIds: string[],
  options: { submitAnswer?: (id: string, draft: AnswerDraft) => void } = {},
) {
  return renderHook(() =>
    useStudySessionController({
      userId: scope.userId,
      workbookId: scope.workbookId,
      practice: scope.practice,
      questions: questionIds.map(questionFixture),
      submitAnswer: options.submitAnswer ?? vi.fn(),
    }),
  );
}

function storedAnsweredIds(): unknown {
  return JSON.parse(window.localStorage.getItem(studySessionStorageKey(scope)) ?? "{}").answeredIds;
}

describe("useStudySessionController", () => {
  beforeEach(() => {
    // given (shared): a live session saved by useStudyResume on mount
    window.localStorage.clear();
    saveStudySession({ ...scope, startedAtMs: Date.now(), answeredIds: [] });
  });

  it("should present the first question when mounted", () => {
    // given / when
    const { result } = renderController(["q-1", "q-2"]);

    // then
    expect(result.current.question?.questionId).toBe("q-1");
  });

  it("should present the next question when the answer is correct", () => {
    // given
    const { result } = renderController(["q-1", "q-2"]);

    // when
    act(() => result.current.answer(correctDraft));

    // then
    expect(result.current.question?.questionId).toBe("q-2");
  });

  it("should count a correct answer when the answer is correct", () => {
    // given
    const { result } = renderController(["q-1", "q-2"]);

    // when
    act(() => result.current.answer(correctDraft));

    // then
    expect(result.current.correctCount).toBe(1);
  });

  it("should record the question as answered when the answer is correct", () => {
    // given
    const { result } = renderController(["q-1", "q-2"]);

    // when
    act(() => result.current.answer(correctDraft));

    // then
    expect(storedAnsweredIds()).toEqual(["q-1"]);
  });

  it("should keep the saved session when a question other than the last is answered", () => {
    // given
    const { result } = renderController(["q-1", "q-2"]);

    // when
    act(() => result.current.answer(correctDraft));

    // then
    expect(window.localStorage.getItem(studySessionStorageKey(scope))).not.toBeNull();
  });

  it("should submit the answer for the presented question when answered", () => {
    // given
    const submitAnswer = vi.fn();
    const { result } = renderController(["q-1", "q-2"], { submitAnswer });

    // when
    act(() => result.current.answer(wrongDraft));

    // then
    expect(submitAnswer).toHaveBeenCalledWith("q-1", wrongDraft);
  });

  it("should move the question to the back of the queue when the answer is wrong", () => {
    // given
    const { result } = renderController(["q-1", "q-2", "q-3"]);

    // when
    act(() => result.current.answer(wrongDraft));

    // then: q-2 and q-3 come before q-1 again
    act(() => result.current.skip());
    act(() => result.current.skip());
    expect(result.current.question?.questionId).toBe("q-1");
  });

  it("should count an incorrect answer when the answer is wrong", () => {
    // given
    const { result } = renderController(["q-1", "q-2"]);

    // when
    act(() => result.current.answer(wrongDraft));

    // then
    expect(result.current.incorrectCount).toBe(1);
  });

  it("should not record the question as answered when the answer is wrong", () => {
    // given
    const { result } = renderController(["q-1", "q-2"]);

    // when
    act(() => result.current.answer(wrongDraft));

    // then
    expect(storedAnsweredIds()).toEqual([]);
  });

  it("should change the card key when the same question is presented again", () => {
    // given
    const { result } = renderController(["q-1"]);
    const firstCardKey = result.current.cardKey;

    // when
    act(() => result.current.answer(wrongDraft));

    // then
    expect(result.current.cardKey).not.toBe(firstCardKey);
  });

  it("should finish the session when the last question is answered correctly", () => {
    // given
    const { result } = renderController(["q-1"]);

    // when
    act(() => result.current.answer(correctDraft));

    // then
    expect(result.current.question).toBeUndefined();
  });

  it("should clear the saved session when the last question is answered correctly", () => {
    // given
    const { result } = renderController(["q-1"]);

    // when
    act(() => result.current.answer(correctDraft));

    // then
    expect(window.localStorage.getItem(studySessionStorageKey(scope))).toBeNull();
  });

  it("should present the next question when the question is skipped", () => {
    // given
    const { result } = renderController(["q-1", "q-2"]);

    // when
    act(() => result.current.skip());

    // then
    expect(result.current.question?.questionId).toBe("q-2");
  });

  it("should not count a skipped question as correct", () => {
    // given
    const { result } = renderController(["q-1", "q-2"]);

    // when
    act(() => result.current.skip());

    // then
    expect(result.current.correctCount).toBe(0);
  });

  it("should not count a skipped question as incorrect", () => {
    // given
    const { result } = renderController(["q-1", "q-2"]);

    // when
    act(() => result.current.skip());

    // then
    expect(result.current.incorrectCount).toBe(0);
  });

  it("should not record a skipped question as answered", () => {
    // given
    const { result } = renderController(["q-1", "q-2"]);

    // when
    act(() => result.current.skip());

    // then
    expect(storedAnsweredIds()).toEqual([]);
  });

  it("should not submit an answer when the question is skipped", () => {
    // given
    const submitAnswer = vi.fn();
    const { result } = renderController(["q-1", "q-2"], { submitAnswer });

    // when
    act(() => result.current.skip());

    // then
    expect(submitAnswer).not.toHaveBeenCalled();
  });

  it("should clear the saved session when the last question is skipped", () => {
    // given
    const { result } = renderController(["q-1"]);

    // when
    act(() => result.current.skip());

    // then
    expect(window.localStorage.getItem(studySessionStorageKey(scope))).toBeNull();
  });
});
