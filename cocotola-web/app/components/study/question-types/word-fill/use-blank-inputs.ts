import { useEffect, useRef, useState } from "react";
import type { AnswerDraft } from "~/components/study/question-shell/types";
import { findNextUnlocked, isCorrectAnswer, toWordFillDraft } from "./answer";

type UseBlankInputsArgs = {
  answers: readonly string[];
  revealed: boolean;
  onCommit: (draft: AnswerDraft<"word_fill">) => void;
};

export function useBlankInputs({ answers, revealed, onCommit }: UseBlankInputsArgs) {
  const [inputs, setInputs] = useState<string[]>(() => answers.map(() => ""));
  const inputRefs = useRef<HTMLInputElement[]>([]);

  useEffect(() => {
    inputRefs.current[0]?.focus();
  }, []);

  // A blank is locked once it holds the correct answer: read-only and skipped by focus traversal.
  const correct = answers.map((answer, i) => isCorrectAnswer(inputs[i] ?? "", answer));

  function handleInputChange(index: number, value: string) {
    if (correct[index] || revealed) return;

    const next = inputs.map((v, i) => (i === index ? value : v));
    setInputs(next);

    const nextCorrect = answers.map((answer, i) => isCorrectAnswer(next[i] ?? "", answer));
    if (nextCorrect.every(Boolean)) {
      onCommit(toWordFillDraft(true));
      return;
    }
    if (!nextCorrect[index]) return;

    const focusIndex = findNextUnlocked(index, nextCorrect);
    if (focusIndex >= 0) {
      inputRefs.current[focusIndex]?.focus();
      inputRefs.current[focusIndex]?.select();
    }
  }

  function registerInput(index: number) {
    return (el: HTMLInputElement | null) => {
      if (el) inputRefs.current[index] = el;
    };
  }

  return { inputs, correct, handleInputChange, registerInput };
}
