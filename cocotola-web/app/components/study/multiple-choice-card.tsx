import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/components/ui/button";
import { type Choice, parseMultipleChoiceContent } from "~/components/workbook/schemas";

type MultipleChoiceCardProps = {
  content: string;
  onAnswer: (correct: boolean) => void;
};

function shuffle<T>(arr: T[]): T[] {
  const result = [...arr];
  for (let i = result.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [result[i], result[j]] = [result[j], result[i]];
  }
  return result;
}

export function MultipleChoiceCard({ content, onAnswer }: MultipleChoiceCardProps) {
  const { t } = useTranslation();
  const parsed = parseMultipleChoiceContent(content);
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const choices = useMemo(() => {
    if (!parsed?.choices) return [];
    return parsed.shuffleChoices ? shuffle(parsed.choices) : parsed.choices;
  }, [parsed?.choices, parsed?.shuffleChoices]);

  if (!parsed) {
    return <p className="text-sm text-muted-foreground">{content}</p>;
  }

  const answered = selectedId !== null;

  function handleSelect(choice: Choice) {
    if (answered) return;
    setSelectedId(choice.id);
  }

  function handleNext() {
    const selected = choices.find((c) => c.id === selectedId);
    onAnswer(selected?.isCorrect ?? false);
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg bg-muted/50 p-4">
        <p className="text-lg font-medium">{parsed.questionText}</p>
      </div>

      <div className="space-y-2">
        {choices.map((choice) => {
          let variant: "outline" | "default" | "destructive" = "outline";
          let extraClass = "";

          if (answered) {
            if (choice.isCorrect) {
              variant = "default";
              extraClass =
                "border-green-500 bg-green-100 text-green-900 hover:bg-green-100 dark:bg-green-950/30 dark:text-green-300";
            } else if (choice.id === selectedId) {
              variant = "destructive";
            }
          }

          return (
            <Button
              key={choice.id}
              variant={variant}
              className={`w-full justify-start text-left ${extraClass}`}
              onClick={() => handleSelect(choice)}
              disabled={answered}
            >
              {choice.text}
            </Button>
          );
        })}
      </div>

      {parsed.explanation && answered && (
        <p className="text-sm text-muted-foreground">{parsed.explanation}</p>
      )}

      {answered &&
        (() => {
          const isCorrect = choices.find((c) => c.id === selectedId)?.isCorrect ?? false;
          return (
            <div className="flex items-center justify-end gap-3">
              <span
                className={`text-sm font-medium ${isCorrect ? "text-green-600" : "text-red-600"}`}
              >
                {isCorrect ? t("workbooks.study.correct") : t("workbooks.study.incorrect")}
              </span>
              <Button onClick={handleNext}>{t("workbooks.study.next")}</Button>
            </div>
          );
        })()}
    </div>
  );
}
