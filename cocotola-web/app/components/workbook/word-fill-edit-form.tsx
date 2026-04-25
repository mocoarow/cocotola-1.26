import { CheckIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { useFetcher } from "react-router";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import type { Question } from "~/lib/api/question.server";
import { LangSelect } from "./lang-select";

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
  };
  fetcher: ReturnType<typeof useFetcher>;
  onCancel: () => void;
}) {
  const { t } = useTranslation();
  const isSubmitting = fetcher.state !== "idle";

  return (
    <fetcher.Form method="post" className="space-y-3">
      <input type="hidden" name="intent" value="updateQuestion" />
      <input type="hidden" name="questionId" value={question.questionId} />
      <input type="hidden" name="questionType" value="word_fill" />
      <input type="hidden" name="orderIndex" value={question.orderIndex} />
      <div className="grid gap-3 sm:grid-cols-2">
        <div className="space-y-1">
          <label htmlFor="edit-sourceLang" className="text-xs font-medium">
            {t("workbooks.addQuestion.sourceLang")}
          </label>
          <LangSelect
            id="edit-sourceLang"
            name="sourceLang"
            defaultValue={parsed.source?.lang ?? "ja"}
          />
        </div>
        <div className="space-y-1">
          <label htmlFor="edit-sourceText" className="text-xs font-medium">
            {t("workbooks.addQuestion.sourceText")}
          </label>
          <Input
            id="edit-sourceText"
            name="sourceText"
            defaultValue={parsed.source?.text ?? ""}
            required
          />
        </div>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <div className="space-y-1">
          <label htmlFor="edit-targetLang" className="text-xs font-medium">
            {t("workbooks.addQuestion.targetLang")}
          </label>
          <LangSelect
            id="edit-targetLang"
            name="targetLang"
            defaultValue={parsed.target?.lang ?? "en"}
          />
        </div>
        <div className="space-y-1">
          <label htmlFor="edit-targetText" className="text-xs font-medium">
            {t("workbooks.addQuestion.targetText")}
          </label>
          <Input
            id="edit-targetText"
            name="targetText"
            defaultValue={parsed.target?.text ?? ""}
            required
          />
        </div>
      </div>
      <div className="space-y-1">
        <label htmlFor="edit-explanation" className="text-xs font-medium">
          {t("workbooks.addQuestion.explanation")}
        </label>
        <Input id="edit-explanation" name="explanation" defaultValue={parsed.explanation ?? ""} />
      </div>
      <div className="space-y-1">
        <label htmlFor="edit-tags" className="text-xs font-medium">
          {t("workbooks.addQuestion.tags")}
        </label>
        <Input id="edit-tags" name="tags" defaultValue={question.tags?.join(", ") ?? ""} />
      </div>
      <div className="flex gap-2">
        <Button type="submit" size="sm" disabled={isSubmitting}>
          <CheckIcon data-icon="inline-start" className="size-3.5" />
          <span>{isSubmitting ? t("common.saving") : t("common.save")}</span>
        </Button>
        <Button type="button" size="sm" variant="outline" onClick={onCancel}>
          {t("common.cancel")}
        </Button>
      </div>
    </fetcher.Form>
  );
}
