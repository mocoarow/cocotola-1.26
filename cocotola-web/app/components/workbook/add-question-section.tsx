import { CircleCheckIcon, CircleIcon, PlusIcon, XIcon } from "lucide-react";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFetcher } from "react-router";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "~/components/ui/sheet";
import { LangSelect } from "./lang-select";
import type { Choice, QuestionType } from "./schemas";

function WordFillForm({ onAdded }: { onAdded: () => void }) {
  const fetcher = useFetcher<{ ok: boolean; added?: boolean }>();
  const { t } = useTranslation();
  const isSubmitting = fetcher.state !== "idle";

  const added = useRef(false);
  if (fetcher.state === "idle" && fetcher.data?.added && !added.current) {
    added.current = true;
    onAdded();
  }
  if (fetcher.state === "submitting") {
    added.current = false;
  }

  return (
    <fetcher.Form method="post" className="flex flex-1 flex-col gap-4 overflow-y-auto px-4">
      <input type="hidden" name="intent" value="addWordFill" />

      <div className="space-y-1.5">
        <label htmlFor="sourceLang" className="text-sm font-medium">
          {t("workbooks.addQuestion.sourceLang")}
        </label>
        <LangSelect id="sourceLang" name="sourceLang" defaultValue="ja" />
      </div>
      <div className="space-y-1.5">
        <label htmlFor="sourceText" className="text-sm font-medium">
          {t("workbooks.addQuestion.sourceText")}
        </label>
        <Input id="sourceText" name="sourceText" placeholder="ゴミを捨てる" required />
      </div>

      <div className="space-y-1.5">
        <label htmlFor="targetLang" className="text-sm font-medium">
          {t("workbooks.addQuestion.targetLang")}
        </label>
        <LangSelect id="targetLang" name="targetLang" defaultValue="en" />
      </div>
      <div className="space-y-1.5">
        <label htmlFor="targetText" className="text-sm font-medium">
          {t("workbooks.addQuestion.targetText")}
        </label>
        <Input id="targetText" name="targetText" placeholder="{{throw}} it {{away}}" required />
        <p className="text-xs text-muted-foreground">{t("workbooks.addQuestion.targetHint")}</p>
      </div>

      <div className="space-y-1.5">
        <label htmlFor="explanation" className="text-sm font-medium">
          {t("workbooks.addQuestion.explanation")}
        </label>
        <Input id="explanation" name="explanation" placeholder="throw away は句動詞です。" />
      </div>

      <div className="space-y-1.5">
        <label htmlFor="tags" className="text-sm font-medium">
          {t("workbooks.addQuestion.tags")}
        </label>
        <Input id="tags" name="tags" placeholder="level:beginner,topic:phrasal-verbs" />
        <p className="text-xs text-muted-foreground">{t("workbooks.addQuestion.tagsHint")}</p>
      </div>

      <SheetFooter>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? t("workbooks.addQuestion.adding") : t("workbooks.addQuestion.submit")}
        </Button>
      </SheetFooter>
    </fetcher.Form>
  );
}

const INITIAL_CHOICES: Choice[] = [
  { id: "1", text: "", isCorrect: true },
  { id: "2", text: "", isCorrect: false },
];

function MultipleChoiceForm({ onAdded }: { onAdded: () => void }) {
  const fetcher = useFetcher<{ ok: boolean; added?: boolean }>();
  const { t } = useTranslation();
  const isSubmitting = fetcher.state !== "idle";
  const [choices, setChoices] = useState<Choice[]>(INITIAL_CHOICES);
  const [shuffleChoices, setShuffleChoices] = useState(true);
  const [showCorrectCount, setShowCorrectCount] = useState(false);

  const added = useRef(false);
  if (fetcher.state === "idle" && fetcher.data?.added && !added.current) {
    added.current = true;
    onAdded();
  }
  if (fetcher.state === "submitting") {
    added.current = false;
  }

  return (
    <fetcher.Form method="post" className="flex flex-1 flex-col gap-4 overflow-y-auto px-4">
      <input type="hidden" name="intent" value="addMultipleChoice" />
      <input type="hidden" name="choices" value={JSON.stringify(choices)} />
      <input type="hidden" name="shuffleChoices" value={String(shuffleChoices)} />
      <input type="hidden" name="showCorrectCount" value={String(showCorrectCount)} />

      <div className="space-y-1.5">
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

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="mc-show-correct-count"
          checked={showCorrectCount}
          onChange={(e) => setShowCorrectCount(e.target.checked)}
          className="size-4 rounded border-input"
        />
        <label htmlFor="mc-show-correct-count" className="text-sm font-medium">
          {t("workbooks.addQuestion.showCorrectCount")}
        </label>
      </div>

      <div className="space-y-1.5">
        <label htmlFor="mc-explanation" className="text-sm font-medium">
          {t("workbooks.addQuestion.explanation")}
        </label>
        <Input id="mc-explanation" name="explanation" />
      </div>

      <div className="space-y-1.5">
        <label htmlFor="mc-tags" className="text-sm font-medium">
          {t("workbooks.addQuestion.tags")}
        </label>
        <Input id="mc-tags" name="tags" placeholder="level:beginner,topic:grammar" />
        <p className="text-xs text-muted-foreground">{t("workbooks.addQuestion.tagsHint")}</p>
      </div>

      <SheetFooter>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? t("workbooks.addQuestion.adding") : t("workbooks.addQuestion.submit")}
        </Button>
      </SheetFooter>
    </fetcher.Form>
  );
}

export function AddQuestionSection() {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [selectedType, setSelectedType] = useState<QuestionType | null>(null);
  const [formKey, setFormKey] = useState(0);

  const handleClose = () => setOpen(false);

  const sheetTitle =
    selectedType === "word_fill"
      ? t("workbooks.addQuestion.titleWordFill")
      : selectedType === "multiple_choice"
        ? t("workbooks.addQuestion.titleMultipleChoice")
        : t("workbooks.addQuestion.selectType");

  return (
    <Sheet
      open={open}
      onOpenChange={(nextOpen) => {
        setOpen(nextOpen);
        if (!nextOpen) {
          setSelectedType(null);
          setFormKey((k) => k + 1);
        }
      }}
    >
      <SheetTrigger
        render={
          <Button size="sm">
            <PlusIcon data-icon="inline-start" className="size-4" />
            <span>{t("workbooks.addQuestion.submit")}</span>
          </Button>
        }
      />
      <SheetContent side="right">
        <SheetHeader>
          <SheetTitle>{sheetTitle}</SheetTitle>
          {selectedType && (
            <SheetDescription>
              <button
                type="button"
                onClick={() => setSelectedType(null)}
                className="text-sm text-primary hover:underline"
              >
                {t("workbooks.addQuestion.changeType")}
              </button>
            </SheetDescription>
          )}
        </SheetHeader>

        {selectedType === null && (
          <div className="flex flex-col gap-3 px-4">
            <button
              type="button"
              onClick={() => setSelectedType("word_fill")}
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
              onClick={() => setSelectedType("multiple_choice")}
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
        )}

        {selectedType === "word_fill" && <WordFillForm key={formKey} onAdded={handleClose} />}
        {selectedType === "multiple_choice" && (
          <MultipleChoiceForm key={formKey} onAdded={handleClose} />
        )}
      </SheetContent>
    </Sheet>
  );
}
