import { useTranslation } from "react-i18next";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { SheetFooter } from "~/components/ui/sheet";
import type { Question } from "~/lib/api/question.server";
import { SourceLangField, TargetLangField, useLangPair } from "./lang-pair";
import { type QuestionEditFetcher, QuestionFormError } from "./question-form-error";

export function WordFillEditForm({
  question,
  parsed,
  fetcher,
  onCancel,
}: {
  question: Question;
  parsed: {
    source?: { text: string; lang: string };
    target?: { text: string; lang: string };
    explanation?: string;
    explanation1?: string;
    explanation2?: string;
  };
  fetcher: QuestionEditFetcher;
  onCancel: () => void;
}) {
  const { t } = useTranslation();
  const langs = useLangPair(parsed.source?.lang, parsed.target?.lang);
  const isSubmitting = fetcher.state !== "idle";

  return (
    <fetcher.Form method="post" className="flex flex-1 flex-col gap-4 overflow-y-auto px-4">
      <input type="hidden" name="intent" value="updateQuestion" />
      <input type="hidden" name="questionId" value={question.questionId} />
      <input type="hidden" name="questionType" value="word_fill" />
      <input type="hidden" name="orderIndex" value={question.orderIndex} />
      {/* Preserve generated source-attribution explanations across edits; they
          are not user-editable here, so round-trip them as hidden fields. */}
      {parsed.explanation1 && (
        <input type="hidden" name="explanation1" value={parsed.explanation1} />
      )}
      {parsed.explanation2 && (
        <input type="hidden" name="explanation2" value={parsed.explanation2} />
      )}

      <SourceLangField id="edit-sourceLang" langs={langs} />
      <div className="space-y-1.5">
        <label htmlFor="edit-sourceText" className="text-sm font-medium">
          {t("workbooks.addQuestion.sourceText")}
        </label>
        <Input
          id="edit-sourceText"
          name="sourceText"
          defaultValue={parsed.source?.text ?? ""}
          required
        />
      </div>

      <TargetLangField id="edit-targetLang" langs={langs} />
      <div className="space-y-1.5">
        <label htmlFor="edit-targetText" className="text-sm font-medium">
          {t("workbooks.addQuestion.targetText")}
        </label>
        <Input
          id="edit-targetText"
          name="targetText"
          defaultValue={parsed.target?.text ?? ""}
          required
          aria-describedby="edit-targetText-hint"
        />
        <p id="edit-targetText-hint" className="text-xs text-muted-foreground">
          {t("workbooks.addQuestion.targetHint")}
        </p>
      </div>

      <div className="space-y-1.5">
        <label htmlFor="edit-explanation" className="text-sm font-medium">
          {t("workbooks.addQuestion.explanation")}
        </label>
        <Input id="edit-explanation" name="explanation" defaultValue={parsed.explanation ?? ""} />
      </div>
      <div className="space-y-1.5">
        <label htmlFor="edit-tags" className="text-sm font-medium">
          {t("workbooks.addQuestion.tags")}
        </label>
        <Input
          id="edit-tags"
          name="tags"
          defaultValue={question.tags?.join(", ") ?? ""}
          aria-describedby="edit-tags-hint"
        />
        <p id="edit-tags-hint" className="text-xs text-muted-foreground">
          {t("workbooks.addQuestion.tagsHint")}
        </p>
      </div>

      <QuestionFormError errorKey={fetcher.data?.errorKey} />
      <SheetFooter>
        <Button type="submit" disabled={isSubmitting || langs.sameLanguage}>
          {isSubmitting ? t("common.saving") : t("common.save")}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel}>
          {t("common.cancel")}
        </Button>
      </SheetFooter>
    </fetcher.Form>
  );
}
