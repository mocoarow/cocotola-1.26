import { CircleCheckIcon, CircleIcon, PlusIcon, XIcon } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import type { useFetcher } from "react-router";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { SheetFooter } from "~/components/ui/sheet";
import type { Question } from "~/lib/api/question.server";
import type { Choice } from "./schemas";

export function MultipleChoiceEditForm({
  question,
  parsed,
  fetcher,
  onCancel,
}: {
  question: Question;
  parsed: {
    questionText?: string;
    explanation?: string;
    choices?: Choice[];
    shuffleChoices?: boolean;
    showCorrectCount?: boolean;
  };
  fetcher: ReturnType<typeof useFetcher>;
  onCancel: () => void;
}) {
  const { t } = useTranslation();
  const isSubmitting = fetcher.state !== "idle";
  const [choices, setChoices] = useState<Choice[]>(
    parsed.choices ?? [
      { id: "1", text: "", isCorrect: true },
      { id: "2", text: "", isCorrect: false },
    ],
  );
  const [shuffleChoices, setShuffleChoices] = useState(parsed.shuffleChoices ?? true);
  const [showCorrectCount, setShowCorrectCount] = useState(parsed.showCorrectCount ?? false);

  return (
    <fetcher.Form method="post" className="flex flex-1 flex-col gap-4 overflow-y-auto px-4">
      <input type="hidden" name="intent" value="updateQuestion" />
      <input type="hidden" name="questionId" value={question.questionId} />
      <input type="hidden" name="questionType" value="multiple_choice" />
      <input type="hidden" name="orderIndex" value={question.orderIndex} />
      <input type="hidden" name="choices" value={JSON.stringify(choices)} />
      <input type="hidden" name="shuffleChoices" value={String(shuffleChoices)} />
      <input type="hidden" name="showCorrectCount" value={String(showCorrectCount)} />

      <div className="space-y-1.5">
        <label htmlFor="edit-mc-questionText" className="text-sm font-medium">
          {t("workbooks.addQuestion.questionText")}
        </label>
        <Input
          id="edit-mc-questionText"
          name="questionText"
          defaultValue={parsed.questionText ?? ""}
          required
        />
      </div>

      <div className="space-y-2">
        <span className="text-sm font-medium">{t("workbooks.addQuestion.choices")}</span>
        {choices.map((choice, index) => (
          <div key={choice.id} className="flex items-center gap-2">
            <button
              type="button"
              onClick={() =>
                setChoices(
                  choices.map((c) => (c.id === choice.id ? { ...c, isCorrect: !c.isCorrect } : c)),
                )
              }
              className="shrink-0"
              title={t("workbooks.addQuestion.correct")}
            >
              {choice.isCorrect ? (
                <CircleCheckIcon className="size-5 text-green-600" />
              ) : (
                <CircleIcon className="size-5 text-muted-foreground" />
              )}
            </button>
            <Input
              value={choice.text}
              onChange={(e) =>
                setChoices(
                  choices.map((c) => (c.id === choice.id ? { ...c, text: e.target.value } : c)),
                )
              }
              placeholder={`${t("workbooks.addQuestion.choiceText")} ${index + 1}`}
              className="flex-1"
              required
            />
            {choices.length > 1 && (
              <Button
                type="button"
                size="icon-sm"
                variant="ghost"
                onClick={() => setChoices(choices.filter((c) => c.id !== choice.id))}
              >
                <XIcon className="size-3.5 text-destructive" />
              </Button>
            )}
          </div>
        ))}
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() =>
            setChoices([...choices, { id: crypto.randomUUID(), text: "", isCorrect: false }])
          }
        >
          <PlusIcon className="size-3.5" />
          <span>{t("workbooks.addQuestion.addChoice")}</span>
        </Button>
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="edit-mc-shuffle"
          checked={shuffleChoices}
          onChange={(e) => setShuffleChoices(e.target.checked)}
          className="size-4 rounded border-input"
        />
        <label htmlFor="edit-mc-shuffle" className="text-sm font-medium">
          {t("workbooks.addQuestion.shuffleChoices")}
        </label>
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="edit-mc-show-correct-count"
          checked={showCorrectCount}
          onChange={(e) => setShowCorrectCount(e.target.checked)}
          className="size-4 rounded border-input"
        />
        <label htmlFor="edit-mc-show-correct-count" className="text-sm font-medium">
          {t("workbooks.addQuestion.showCorrectCount")}
        </label>
      </div>

      <div className="space-y-1.5">
        <label htmlFor="edit-mc-explanation" className="text-sm font-medium">
          {t("workbooks.addQuestion.explanation")}
        </label>
        <Input
          id="edit-mc-explanation"
          name="explanation"
          defaultValue={parsed.explanation ?? ""}
        />
      </div>

      <div className="space-y-1.5">
        <label htmlFor="edit-mc-tags" className="text-sm font-medium">
          {t("workbooks.addQuestion.tags")}
        </label>
        <Input id="edit-mc-tags" name="tags" defaultValue={question.tags?.join(", ") ?? ""} />
      </div>

      <SheetFooter>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? t("common.saving") : t("common.save")}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel}>
          {t("common.cancel")}
        </Button>
      </SheetFooter>
    </fetcher.Form>
  );
}
