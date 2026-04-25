import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { parseWordFillContent } from "~/components/workbook/schemas";

type WordFillCardProps = {
  content: string;
  onAnswer: (correct: boolean) => void;
};

function extractBlanks(text: string): { segments: string[]; answers: string[] } {
  const segments = text.split(/\{\{([^}]+)\}\}/g);
  const answers: string[] = [];
  for (let i = 1; i < segments.length; i += 2) {
    answers.push(segments[i]);
  }
  return { segments, answers };
}

export function WordFillCard({ content, onAnswer }: WordFillCardProps) {
  const { t } = useTranslation();
  const parsed = parseWordFillContent(content);
  const [inputs, setInputs] = useState<string[]>([]);
  const [checked, setChecked] = useState(false);

  if (!parsed?.target?.text) {
    return <p className="text-sm text-muted-foreground">{content}</p>;
  }

  const { segments, answers } = extractBlanks(parsed.target.text);

  if (inputs.length === 0 && answers.length > 0) {
    setInputs(new Array(answers.length).fill(""));
  }

  const results = answers.map(
    (answer, i) => (inputs[i] ?? "").trim().toLowerCase() === answer.trim().toLowerCase(),
  );
  const allCorrect = results.every(Boolean);

  function handleCheck() {
    setChecked(true);
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
          return (
            <span key={`blank-${inputIndex}`} className="inline-flex flex-col items-center">
              <Input
                aria-label={t("workbooks.study.blankInput", { number: inputIndex + 1 })}
                className={`mx-1 inline-block w-32 text-center ${
                  checked
                    ? results[inputIndex]
                      ? "border-green-500 bg-green-50 dark:bg-green-950/30"
                      : "border-red-500 bg-red-50 dark:bg-red-950/30"
                    : ""
                }`}
                value={inputs[inputIndex] ?? ""}
                onChange={(e) => {
                  const next = [...inputs];
                  next[inputIndex] = e.target.value;
                  setInputs(next);
                }}
                disabled={checked}
              />
              {checked && !results[inputIndex] && (
                <span className="text-xs text-green-600">{answers[inputIndex]}</span>
              )}
            </span>
          );
        })}
      </div>

      {parsed.explanation && checked && (
        <p className="text-sm text-muted-foreground">{parsed.explanation}</p>
      )}

      {!checked ? (
        <div className="flex justify-end">
          <Button onClick={handleCheck} disabled={inputs.some((v) => !v.trim())}>
            {t("workbooks.study.check")}
          </Button>
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
