import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { parseWordFillContent } from "~/components/workbook/schemas";

type WordFillCardProps = {
  content: string;
  onAnswer: (correct: boolean) => void;
};

type Phase = "input" | "result";

function extractBlanks(text: string): { segments: string[]; answers: string[] } {
  const segments = text.split(/\{\{([^}]+)\}\}/g);
  const answers: string[] = [];
  for (let i = 1; i < segments.length; i += 2) {
    answers.push(segments[i]);
  }
  return { segments, answers };
}

function isCorrectAnswer(value: string, answer: string): boolean {
  return value.trim().toLowerCase() === answer.trim().toLowerCase();
}

// findNextUnlocked walks forward from `from` (exclusive) and returns the first
// blank whose `correct` flag is false, wrapping around. Returns -1 when every
// blank is already correct.
function findNextUnlocked(from: number, correct: boolean[]): number {
  for (let step = 1; step <= correct.length; step++) {
    const candidate = (from + step) % correct.length;
    if (!correct[candidate]) return candidate;
  }
  return -1;
}

export function WordFillCard({ content, onAnswer }: WordFillCardProps) {
  const { t } = useTranslation();
  const parsed = parseWordFillContent(content);
  const [inputs, setInputs] = useState<string[]>([]);
  const [phase, setPhase] = useState<Phase>("input");
  const inputRefs = useRef<HTMLInputElement[]>([]);

  // Focus the first blank when this question card mounts.
  useEffect(() => {
    inputRefs.current[0]?.focus();
  }, []);

  if (!parsed?.target?.text) {
    return <p className="text-sm text-muted-foreground">{content}</p>;
  }

  const { segments, answers } = extractBlanks(parsed.target.text);

  if (inputs.length === 0 && answers.length > 0) {
    setInputs(new Array(answers.length).fill(""));
  }

  // A blank is "locked" once it holds the correct answer. Locked blanks become
  // read-only and are skipped by focus traversal — the user cannot accidentally
  // overwrite a value they have already gotten right.
  const correct = answers.map((answer, i) => isCorrectAnswer(inputs[i] ?? "", answer));
  const allCorrect = correct.length > 0 && correct.every(Boolean);
  const isResult = phase === "result";

  function handleInputChange(index: number, value: string) {
    if (correct[index]) return;

    const next = [...inputs];
    next[index] = value;
    setInputs(next);

    if (isResult) return;
    if (!isCorrectAnswer(value, answers[index])) return;

    const nextCorrect = answers.map((answer, i) => isCorrectAnswer(next[i] ?? "", answer));
    if (nextCorrect.every(Boolean)) {
      // Surface the result screen instead of advancing — the user reviews
      // their answer and explicitly continues via the Next button.
      setPhase("result");
      return;
    }

    // Move to the next blank that is still empty/wrong, wrapping around.
    const focusIndex = findNextUnlocked(index, nextCorrect);
    if (focusIndex >= 0) {
      inputRefs.current[focusIndex]?.focus();
      inputRefs.current[focusIndex]?.select();
    }
  }

  function handleReveal() {
    setPhase("result");
  }

  function handleNext() {
    onAnswer(allCorrect);
  }

  return (
    <div className="space-y-6">
      {parsed.source?.text && (
        <div className="rounded-lg bg-muted/50 p-4">
          <p className="text-lg font-medium">{parsed.source.text}</p>
        </div>
      )}

      <div className="flex flex-wrap items-center gap-1 text-lg">
        {segments.map((segment, i) => {
          if (i % 2 === 0) {
            return <span key={`text-${segment}`}>{segment}</span>;
          }
          const inputIndex = Math.floor(i / 2);
          const isLocked = correct[inputIndex];
          return (
            <span key={`blank-${inputIndex}`} className="inline-flex flex-col items-center">
              <Input
                ref={(el: HTMLInputElement | null) => {
                  if (el) inputRefs.current[inputIndex] = el;
                }}
                aria-label={t("workbooks.study.blankInput", { number: inputIndex + 1 })}
                className={`mx-1 inline-block w-32 text-center ${
                  isLocked
                    ? "border-green-500 bg-green-50 dark:bg-green-950/30"
                    : isResult
                      ? "border-red-500 bg-red-50 dark:bg-red-950/30"
                      : ""
                }`}
                value={inputs[inputIndex] ?? ""}
                onChange={(e) => handleInputChange(inputIndex, e.target.value)}
                disabled={isResult || isLocked}
                readOnly={isLocked}
              />
              {isResult && !isLocked && (
                <span className="text-xs text-green-600">{answers[inputIndex]}</span>
              )}
            </span>
          );
        })}
      </div>

      {parsed.explanation && isResult && (
        <p className="text-sm text-muted-foreground">{parsed.explanation}</p>
      )}

      {!isResult ? (
        <div className="flex justify-end">
          <Button onClick={handleReveal}>{t("workbooks.study.showAnswer")}</Button>
        </div>
      ) : (
        <div className="flex items-center justify-end gap-3">
          <span className={`text-sm font-medium ${allCorrect ? "text-green-600" : "text-red-600"}`}>
            {allCorrect ? t("workbooks.study.correct") : t("workbooks.study.incorrect")}
          </span>
          <Button onClick={handleNext}>{t("workbooks.study.next")}</Button>
        </div>
      )}
    </div>
  );
}
