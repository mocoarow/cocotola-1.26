import { useTranslation } from "react-i18next";
import type { FetcherWithComponents } from "react-router";

export const QUESTION_INVALID_ERROR_KEY = "workbooks.addQuestion.errors.invalid";

export type QuestionActionData = { ok: boolean; errorKey?: string };

export type AddQuestionActionData = QuestionActionData & { added?: boolean };

export type QuestionEditFetcher = FetcherWithComponents<QuestionActionData>;

export function QuestionFormError({ errorKey }: { errorKey?: string }) {
  const { t } = useTranslation();
  return (
    <p aria-live="polite" className="text-sm text-destructive">
      {errorKey && t(errorKey)}
    </p>
  );
}
