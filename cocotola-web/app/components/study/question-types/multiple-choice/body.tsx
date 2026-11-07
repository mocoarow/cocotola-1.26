import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import type { BodyProps } from "~/components/study/question-shell/types";
import { shuffle } from "~/components/study/question-types/shuffle";
import { Button } from "~/components/ui/button";
import type { Choice } from "~/components/workbook/schemas";
import { type MultipleChoiceContent, toMultipleChoiceDraft } from "./answer";

type ChoiceVariant = "outline" | "default" | "destructive";

const CORRECT_CHOICE_CLASS =
  "border-green-500 bg-green-100 text-green-900 hover:bg-green-100 dark:bg-green-950/30 dark:text-green-300";

function choiceStyle(
  choice: Choice,
  isSelected: boolean,
  revealed: boolean,
): { variant: ChoiceVariant; className: string } {
  if (!revealed) return { variant: isSelected ? "default" : "outline", className: "" };
  if (choice.isCorrect) {
    return { variant: isSelected ? "default" : "outline", className: CORRECT_CHOICE_CLASS };
  }
  return { variant: isSelected ? "destructive" : "outline", className: "" };
}

export function MultipleChoiceBody({
  content,
  phase,
  onDraftChange,
}: BodyProps<"multiple_choice", MultipleChoiceContent>) {
  const { t } = useTranslation();
  const [choices] = useState(() => {
    const original = content.choices ?? [];
    return content.shuffleChoices ? shuffle(original) : original;
  });
  const [selectedIds, setSelectedIds] = useState<ReadonlySet<string>>(() => new Set());
  const correctIds = useMemo(
    () => new Set(choices.filter((c) => c.isCorrect).map((c) => c.id)),
    [choices],
  );
  const revealed = phase === "revealed";

  function toggle(choice: Choice) {
    if (revealed) return;
    const next = selectedIds.has(choice.id)
      ? new Set([...selectedIds].filter((id) => id !== choice.id))
      : new Set([...selectedIds, choice.id]);
    setSelectedIds(next);
    onDraftChange(toMultipleChoiceDraft(next, correctIds));
  }

  return (
    <>
      <div className="rounded-lg bg-muted/50 p-4">
        <p className="text-lg font-medium">{content.questionText}</p>
        {content.showCorrectCount === true && (
          <p className="mt-2 text-sm text-muted-foreground">
            {t("workbooks.study.selectCount", { count: correctIds.size })}
          </p>
        )}
      </div>

      <div className="space-y-2">
        {choices.map((choice) => {
          const isSelected = selectedIds.has(choice.id);
          const { variant, className } = choiceStyle(choice, isSelected, revealed);
          return (
            <Button
              key={choice.id}
              variant={variant}
              className={`w-full justify-start text-left ${className}`}
              onClick={() => toggle(choice)}
              aria-pressed={isSelected}
              disabled={revealed}
            >
              {choice.text}
            </Button>
          );
        })}
      </div>
    </>
  );
}
