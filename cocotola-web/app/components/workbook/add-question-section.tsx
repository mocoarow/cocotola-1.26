import { CircleCheckIcon, CircleIcon, PlusIcon, XIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFetcher } from "react-router";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { LangSelect } from "./lang-select";
import type { Choice, QuestionType } from "./schemas";

function useFormReset(
  fetcher: ReturnType<typeof useFetcher<{ ok: boolean; added?: boolean }>>,
  onReset?: () => void,
) {
  const [formKey, setFormKey] = useState(0);
  const prevState = useRef(fetcher.state);

  useEffect(() => {
    if (prevState.current === "loading" && fetcher.state === "idle" && fetcher.data?.added) {
      setFormKey((k) => k + 1);
      onReset?.();
    }
    prevState.current = fetcher.state;
  }, [fetcher.state, fetcher.data, onReset]);

  return formKey;
}

function AddWordFillForm({ onChangeType }: { onChangeType: () => void }) {
  const fetcher = useFetcher();
  const { t } = useTranslation();
  const isSubmitting = fetcher.state !== "idle";
  const formKey = useFormReset(fetcher);

  return (
    <div className="rounded-lg border bg-card p-5 shadow-sm">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-base font-semibold">{t("workbooks.addQuestion.titleWordFill")}</h2>
        <Button size="sm" variant="ghost" onClick={onChangeType}>
          {t("workbooks.addQuestion.changeType")}
        </Button>
      </div>
      <fetcher.Form method="post" className="space-y-4" key={formKey}>
        <input type="hidden" name="intent" value="addWordFill" />

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <label htmlFor="sourceLang" className="text-sm font-medium">
              {t("workbooks.addQuestion.sourceLang")}
            </label>
            <LangSelect id="sourceLang" name="sourceLang" defaultValue="ja" />
          </div>
          <div className="space-y-2">
            <label htmlFor="sourceText" className="text-sm font-medium">
              {t("workbooks.addQuestion.sourceText")}
            </label>
            <Input id="sourceText" name="sourceText" placeholder="ゴミを捨てる" required />
          </div>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <label htmlFor="targetLang" className="text-sm font-medium">
              {t("workbooks.addQuestion.targetLang")}
            </label>
            <LangSelect id="targetLang" name="targetLang" defaultValue="en" />
          </div>
          <div className="space-y-2">
            <label htmlFor="targetText" className="text-sm font-medium">
              {t("workbooks.addQuestion.targetText")}
            </label>
            <Input id="targetText" name="targetText" placeholder="{{throw}} it {{away}}" required />
            <p className="text-xs text-muted-foreground">{t("workbooks.addQuestion.targetHint")}</p>
          </div>
        </div>

        <div className="space-y-2">
          <label htmlFor="explanation" className="text-sm font-medium">
            {t("workbooks.addQuestion.explanation")}
          </label>
          <Input id="explanation" name="explanation" placeholder="throw away は句動詞です。" />
        </div>

        <div className="space-y-2">
          <label htmlFor="tags" className="text-sm font-medium">
            {t("workbooks.addQuestion.tags")}
          </label>
          <Input id="tags" name="tags" placeholder="level:beginner,topic:phrasal-verbs" />
          <p className="text-xs text-muted-foreground">{t("workbooks.addQuestion.tagsHint")}</p>
        </div>

        <div className="flex gap-2">
          <Button type="submit" disabled={isSubmitting}>
            <PlusIcon data-icon="inline-start" className="size-3.5" />
            <span>
              {isSubmitting ? t("workbooks.addQuestion.adding") : t("workbooks.addQuestion.submit")}
            </span>
          </Button>
        </div>
      </fetcher.Form>
    </div>
  );
}

const INITIAL_CHOICES: Choice[] = [
  { id: "1", text: "", isCorrect: true },
  { id: "2", text: "", isCorrect: false },
];

function AddMultipleChoiceForm({ onChangeType }: { onChangeType: () => void }) {
  const fetcher = useFetcher();
  const { t } = useTranslation();
  const isSubmitting = fetcher.state !== "idle";
  const [choices, setChoices] = useState<Choice[]>(INITIAL_CHOICES);
  const [shuffleChoices, setShuffleChoices] = useState(true);

  const resetForm = useRef(() => {
    setChoices(INITIAL_CHOICES);
    setShuffleChoices(true);
  });
  const formKey = useFormReset(fetcher, resetForm.current);

  return (
    <div className="rounded-lg border bg-card p-5 shadow-sm">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-base font-semibold">
          {t("workbooks.addQuestion.titleMultipleChoice")}
        </h2>
        <Button size="sm" variant="ghost" onClick={onChangeType}>
          {t("workbooks.addQuestion.changeType")}
        </Button>
      </div>
      <fetcher.Form method="post" className="space-y-4" key={formKey}>
        <input type="hidden" name="intent" value="addMultipleChoice" />
        <input type="hidden" name="choices" value={JSON.stringify(choices)} />
        <input type="hidden" name="shuffleChoices" value={String(shuffleChoices)} />

        <div className="space-y-2">
          <label htmlFor="mc-questionText" className="text-sm font-medium">
            {t("workbooks.addQuestion.questionText")}
          </label>
          <Input
            id="mc-questionText"
            name="questionText"
            placeholder={t("workbooks.addQuestion.questionTextPlaceholder")}
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
                    choices.map((c) =>
                      c.id === choice.id ? { ...c, isCorrect: !c.isCorrect } : c,
                    ),
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
                  title={t("workbooks.addQuestion.removeChoice")}
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
            id="mc-shuffle"
            checked={shuffleChoices}
            onChange={(e) => setShuffleChoices(e.target.checked)}
            className="size-4 rounded border-input"
          />
          <label htmlFor="mc-shuffle" className="text-sm font-medium">
            {t("workbooks.addQuestion.shuffleChoices")}
          </label>
        </div>

        <div className="space-y-2">
          <label htmlFor="mc-explanation" className="text-sm font-medium">
            {t("workbooks.addQuestion.explanation")}
          </label>
          <Input id="mc-explanation" name="explanation" />
        </div>

        <div className="space-y-2">
          <label htmlFor="mc-tags" className="text-sm font-medium">
            {t("workbooks.addQuestion.tags")}
          </label>
          <Input id="mc-tags" name="tags" placeholder="level:beginner,topic:grammar" />
          <p className="text-xs text-muted-foreground">{t("workbooks.addQuestion.tagsHint")}</p>
        </div>

        <div className="flex gap-2">
          <Button type="submit" disabled={isSubmitting}>
            <PlusIcon data-icon="inline-start" className="size-3.5" />
            <span>
              {isSubmitting ? t("workbooks.addQuestion.adding") : t("workbooks.addQuestion.submit")}
            </span>
          </Button>
        </div>
      </fetcher.Form>
    </div>
  );
}

function QuestionTypeSelector({ onSelect }: { onSelect: (type: QuestionType) => void }) {
  const { t } = useTranslation();

  return (
    <div className="rounded-lg border bg-card p-5 shadow-sm">
      <h2 className="mb-4 text-base font-semibold">{t("workbooks.addQuestion.selectType")}</h2>
      <div className="grid gap-3 sm:grid-cols-2">
        <button
          type="button"
          onClick={() => onSelect("word_fill")}
          className="rounded-lg border-2 border-transparent bg-blue-50 p-4 text-left transition-colors hover:border-blue-300 dark:bg-blue-900/20 dark:hover:border-blue-600"
        >
          <p className="font-medium text-blue-700 dark:text-blue-400">
            {t("workbooks.addQuestion.wordFill")}
          </p>
          <p className="mt-1 text-sm text-muted-foreground">
            {t("workbooks.addQuestion.wordFillDescription")}
          </p>
        </button>
        <button
          type="button"
          onClick={() => onSelect("multiple_choice")}
          className="rounded-lg border-2 border-transparent bg-purple-50 p-4 text-left transition-colors hover:border-purple-300 dark:bg-purple-900/20 dark:hover:border-purple-600"
        >
          <p className="font-medium text-purple-700 dark:text-purple-400">
            {t("workbooks.addQuestion.multipleChoice")}
          </p>
          <p className="mt-1 text-sm text-muted-foreground">
            {t("workbooks.addQuestion.multipleChoiceDescription")}
          </p>
        </button>
      </div>
    </div>
  );
}

export function AddQuestionSection() {
  const [selectedType, setSelectedType] = useState<QuestionType | null>(null);

  if (selectedType === null) {
    return <QuestionTypeSelector onSelect={setSelectedType} />;
  }

  if (selectedType === "word_fill") {
    return <AddWordFillForm onChangeType={() => setSelectedType(null)} />;
  }

  return <AddMultipleChoiceForm onChangeType={() => setSelectedType(null)} />;
}
