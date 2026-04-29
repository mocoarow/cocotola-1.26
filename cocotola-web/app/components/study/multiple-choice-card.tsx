import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/components/ui/button";
import { type Choice, parseMultipleChoiceContent } from "~/components/workbook/schemas";

type MultipleChoiceCardProps = {
  content: string;
  // The card reports both the user's selection and the locally-computed strict-match
  // result. The selection is what the server scores authoritatively; the boolean is
  // an optimistic UX value (counter on the result page). Once partial-credit math
  // ships server-side, the counter must switch to read the server's response score.
  onAnswer: (selectedChoiceIds: string[], correct: boolean) => void;
};

function shuffle<T>(arr: T[]): T[] {
  const result = [...arr];
  for (let i = result.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [result[i], result[j]] = [result[j], result[i]];
  }
  return result;
}

function setsEqual(a: ReadonlySet<string>, b: ReadonlySet<string>): boolean {
  if (a.size !== b.size) return false;
  for (const v of a) {
    if (!b.has(v)) return false;
  }
  return true;
}

export function MultipleChoiceCard({ content, onAnswer }: MultipleChoiceCardProps) {
  const { t } = useTranslation();
  const parsed = parseMultipleChoiceContent(content);
  const [selectedIds, setSelectedIds] = useState<ReadonlySet<string>>(() => new Set());
  const [checked, setChecked] = useState(false);

  const choices = useMemo(() => {
    if (!parsed?.choices) return [];
    return parsed.shuffleChoices ? shuffle(parsed.choices) : parsed.choices;
  }, [parsed?.choices, parsed?.shuffleChoices]);

  const correctIds = useMemo(
    () => new Set(choices.filter((c) => c.isCorrect).map((c) => c.id)),
    [choices],
  );

  if (!parsed) {
    return <p className="text-sm text-muted-foreground">{content}</p>;
  }

  const allCorrect = setsEqual(selectedIds, correctIds);

  function toggle(choice: Choice) {
    if (checked) return;
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(choice.id)) {
        next.delete(choice.id);
      } else {
        next.add(choice.id);
      }
      return next;
    });
  }

  function handleCheck() {
    setChecked(true);
  }

  function handleNext() {
    onAnswer(Array.from(selectedIds), allCorrect);
  }

  const showCorrectCount = parsed.showCorrectCount === true;
  const correctCount = correctIds.size;

  return (
    <div className="space-y-6">
      <div className="rounded-lg bg-muted/50 p-4">
        <p className="text-lg font-medium">{parsed.questionText}</p>
        {showCorrectCount && (
          <p className="mt-2 text-sm text-muted-foreground">
            {t("workbooks.study.selectCount", { count: correctCount })}
          </p>
        )}
      </div>

      <div className="space-y-2">
        {choices.map((choice) => {
          const isSelected = selectedIds.has(choice.id);
          let variant: "outline" | "default" | "destructive" = "outline";
          let extraClass = "";

          if (checked) {
            if (choice.isCorrect) {
              variant = isSelected ? "default" : "outline";
              extraClass =
                "border-green-500 bg-green-100 text-green-900 hover:bg-green-100 dark:bg-green-950/30 dark:text-green-300";
            } else if (isSelected) {
              variant = "destructive";
            }
          } else if (isSelected) {
            variant = "default";
          }

          return (
            <Button
              key={choice.id}
              variant={variant}
              className={`w-full justify-start text-left ${extraClass}`}
              onClick={() => toggle(choice)}
              aria-pressed={isSelected}
              disabled={checked}
            >
              {choice.text}
            </Button>
          );
        })}
      </div>

      {parsed.explanation && checked && (
        <p className="text-sm text-muted-foreground">{parsed.explanation}</p>
      )}

      {!checked ? (
        <div className="flex justify-end">
          <Button onClick={handleCheck} disabled={selectedIds.size === 0}>
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
