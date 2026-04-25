import { CircleCheckIcon, CircleIcon, PencilIcon, Trash2Icon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useFetcher } from "react-router";
import { Button } from "~/components/ui/button";
import type { Question } from "~/lib/api/question.server";
import { parseMultipleChoiceContent, parseWordFillContent } from "./schemas";

export function QuestionCard({
  question,
  onEdit,
}: { question: Question; onEdit: (question: Question) => void }) {
  const { t } = useTranslation();
  const deleteFetcher = useFetcher();

  const typeBadge =
    question.questionType === "word_fill" ? (
      <span className="rounded-full bg-blue-100 px-2 py-0.5 text-[11px] font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
        {t("workbooks.addQuestion.wordFill")}
      </span>
    ) : (
      <span className="rounded-full bg-purple-100 px-2 py-0.5 text-[11px] font-medium text-purple-700 dark:bg-purple-900/30 dark:text-purple-400">
        {t("workbooks.addQuestion.multipleChoice")}
      </span>
    );

  const actionButtons = (
    <div className="flex items-center gap-1">
      <Button size="icon-sm" variant="ghost" onClick={() => onEdit(question)}>
        <PencilIcon className="size-3.5" />
        <span className="sr-only">{t("workbooks.detail.editQuestion")}</span>
      </Button>
      <deleteFetcher.Form
        method="post"
        onSubmit={(e) => {
          if (!window.confirm(t("workbooks.detail.deleteQuestionConfirm"))) {
            e.preventDefault();
          }
        }}
      >
        <input type="hidden" name="intent" value="deleteQuestion" />
        <input type="hidden" name="questionId" value={question.questionId} />
        <Button type="submit" size="icon-sm" variant="ghost">
          <Trash2Icon className="size-3.5 text-destructive" />
          <span className="sr-only">{t("workbooks.detail.deleteQuestion")}</span>
        </Button>
      </deleteFetcher.Form>
    </div>
  );

  if (question.questionType === "word_fill") {
    const parsed = parseWordFillContent(question.content);

    return (
      <div className="rounded-lg border bg-card p-4 shadow-sm">
        <div className="mb-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            {typeBadge}
            {question.tags?.map((tag) => (
              <span
                key={tag}
                className="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] text-gray-600 dark:bg-gray-800 dark:text-gray-400"
              >
                {tag}
              </span>
            ))}
          </div>
          {actionButtons}
        </div>
        {parsed ? (
          <div className="space-y-1 text-sm">
            <p>
              <span className="font-medium text-muted-foreground">[{parsed.source?.lang}]</span>{" "}
              {parsed.source?.text}
            </p>
            <p>
              <span className="font-medium text-muted-foreground">[{parsed.target?.lang}]</span>{" "}
              {parsed.target?.text}
            </p>
            {parsed.explanation && (
              <p className="text-xs text-muted-foreground">{parsed.explanation}</p>
            )}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">{question.content}</p>
        )}
      </div>
    );
  }

  if (question.questionType === "multiple_choice") {
    const parsed = parseMultipleChoiceContent(question.content);

    return (
      <div className="rounded-lg border bg-card p-4 shadow-sm">
        <div className="mb-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            {typeBadge}
            {question.tags?.map((tag) => (
              <span
                key={tag}
                className="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] text-gray-600 dark:bg-gray-800 dark:text-gray-400"
              >
                {tag}
              </span>
            ))}
          </div>
          {actionButtons}
        </div>
        {parsed ? (
          <div className="space-y-2 text-sm">
            <p className="font-medium">{parsed.questionText}</p>
            <div className="space-y-1 pl-2">
              {parsed.choices?.map((choice) => (
                <div key={choice.id} className="flex items-center gap-2">
                  {choice.isCorrect ? (
                    <CircleCheckIcon className="size-4 text-green-600" />
                  ) : (
                    <CircleIcon className="size-4 text-muted-foreground" />
                  )}
                  <span className={choice.isCorrect ? "font-medium" : ""}>{choice.text}</span>
                </div>
              ))}
            </div>
            {parsed.explanation && (
              <p className="text-xs text-muted-foreground">{parsed.explanation}</p>
            )}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">{question.content}</p>
        )}
      </div>
    );
  }

  return (
    <div className="rounded-lg border bg-card p-4 shadow-sm">
      <div className="mb-2">
        <span className="rounded-full bg-green-100 px-2 py-0.5 text-[11px] font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">
          {question.questionType}
        </span>
      </div>
      <p className="text-sm text-muted-foreground">{question.content}</p>
    </div>
  );
}
