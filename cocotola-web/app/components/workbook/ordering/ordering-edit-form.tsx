import { useTranslation } from "react-i18next";
import type { OrderingContent } from "~/components/study/question-types/ordering/answer";
import { Button } from "~/components/ui/button";
import { SheetFooter } from "~/components/ui/sheet";
import type { Question } from "~/lib/api/question.server";
import { DEFAULT_SOURCE_LANG, DEFAULT_TARGET_LANG } from "../lang-select";
import { type QuestionEditFetcher, QuestionFormError } from "../question-form-error";
import { formatOrderingNotation } from "./notation";
import {
  OrderingFields,
  type OrderingFieldsDefaults,
  useOrderingNotation,
} from "./ordering-fields";

function editDefaults(question: Question, parsed: OrderingContent | null): OrderingFieldsDefaults {
  return {
    sourceText: parsed?.source.text ?? "",
    sourceLang: parsed?.source.lang ?? DEFAULT_SOURCE_LANG,
    targetLang: parsed?.target.lang ?? DEFAULT_TARGET_LANG,
    explanation: parsed?.explanation ?? "",
    tags: question.tags?.join(", ") ?? "",
  };
}

export function OrderingEditForm({
  question,
  parsed,
  fetcher,
  onCancel,
}: {
  question: Question;
  parsed: OrderingContent | null;
  fetcher: QuestionEditFetcher;
  onCancel: () => void;
}) {
  const { t } = useTranslation();
  const notation = useOrderingNotation(
    parsed ? formatOrderingNotation(parsed.target.segments) : "",
  );
  const isSubmitting = fetcher.state !== "idle";

  return (
    <fetcher.Form method="post" className="flex flex-1 flex-col gap-4 overflow-y-auto px-4">
      <input type="hidden" name="intent" value="updateQuestion" />
      <input type="hidden" name="questionId" value={question.questionId} />
      <input type="hidden" name="questionType" value="ordering" />
      <input type="hidden" name="orderIndex" value={question.orderIndex} />
      <OrderingFields
        idPrefix="edit-ordering-"
        defaults={editDefaults(question, parsed)}
        state={notation}
      />
      <QuestionFormError errorKey={fetcher.data?.errorKey} />
      <SheetFooter>
        <Button type="submit" disabled={isSubmitting || !notation.result.ok}>
          {isSubmitting ? t("common.saving") : t("common.save")}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel}>
          {t("common.cancel")}
        </Button>
      </SheetFooter>
    </fetcher.Form>
  );
}
