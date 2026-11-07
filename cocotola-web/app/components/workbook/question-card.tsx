import { CircleCheckIcon, CircleIcon, PencilIcon, Trash2Icon } from "lucide-react";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useFetcher } from "react-router";
import { useConfirm } from "~/components/confirm-dialog-provider";
import { MarkdownText } from "~/components/markdown-text";
import { Button } from "~/components/ui/button";
import type { Question } from "~/lib/api/question.server";
import { OrderingPreview } from "./ordering/ordering-preview";
import { parseMultipleChoiceContent, parseOrderingContent, parseWordFillContent } from "./schemas";

type BadgeStyle = { labelKey: string; className: string };

function TypeBadge({ badge }: { badge: BadgeStyle }) {
  const { t } = useTranslation();
  return <span className={badge.className}>{t(badge.labelKey)}</span>;
}

function TagList({ tags }: { tags?: string[] }) {
  return tags?.map((tag) => (
    <span
      key={tag}
      className="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] text-gray-600 dark:bg-gray-800 dark:text-gray-400"
    >
      {tag}
    </span>
  ));
}

function ActionButtons({
  question,
  onEdit,
}: {
  question: Question;
  onEdit: (question: Question) => void;
}) {
  const { t } = useTranslation();
  const deleteFetcher = useFetcher();
  const confirm = useConfirm();

  const handleDelete = async () => {
    const confirmed = await confirm({
      title: t("workbooks.detail.deleteQuestionConfirmTitle"),
      description: t("workbooks.detail.deleteQuestionConfirm"),
      confirmLabel: t("common.delete"),
    });
    if (confirmed) {
      deleteFetcher.submit(
        { intent: "deleteQuestion", questionId: question.questionId },
        { method: "post" },
      );
    }
  };

  return (
    <div className="flex items-center gap-1">
      <Button size="icon-sm" variant="ghost" onClick={() => onEdit(question)}>
        <PencilIcon className="size-3.5" />
        <span className="sr-only">{t("workbooks.detail.editQuestion")}</span>
      </Button>
      <Button size="icon-sm" variant="ghost" onClick={handleDelete}>
        <Trash2Icon className="size-3.5 text-destructive" />
        <span className="sr-only">{t("workbooks.detail.deleteQuestion")}</span>
      </Button>
    </div>
  );
}

function CardShell({
  badge,
  question,
  onEdit,
  children,
}: {
  badge: BadgeStyle;
  question: Question;
  onEdit: (question: Question) => void;
  children: ReactNode;
}) {
  return (
    <div className="rounded-lg border bg-card p-4 shadow-sm">
      <div className="mb-2 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <TypeBadge badge={badge} />
          <TagList tags={question.tags} />
        </div>
        <ActionButtons question={question} onEdit={onEdit} />
      </div>
      {children}
    </div>
  );
}

function RawContent({ content }: { content: string }) {
  return <p className="text-sm text-muted-foreground">{content}</p>;
}

function WordFillCardBody({ content }: { content: string }) {
  const parsed = parseWordFillContent(content);
  if (!parsed) return <RawContent content={content} />;

  const explanations = Array.from(
    new Set(
      [parsed.explanation, parsed.explanation1, parsed.explanation2].filter(Boolean) as string[],
    ),
  );
  return (
    <div className="space-y-1 text-sm">
      <p>
        <span className="font-medium text-muted-foreground">[{parsed.source?.lang}]</span>{" "}
        {parsed.source?.text}
      </p>
      <p>
        <span className="font-medium text-muted-foreground">[{parsed.target?.lang}]</span>{" "}
        {parsed.target?.text}
      </p>
      {explanations.map((text) => (
        <MarkdownText key={text} text={text} className="text-xs text-muted-foreground" />
      ))}
    </div>
  );
}

function MultipleChoiceCardBody({ content }: { content: string }) {
  const parsed = parseMultipleChoiceContent(content);
  if (!parsed) return <RawContent content={content} />;

  return (
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
        <MarkdownText text={parsed.explanation} className="text-xs text-muted-foreground" />
      )}
    </div>
  );
}

function OrderingCardBody({ content }: { content: string }) {
  const { t } = useTranslation();
  const parsed = parseOrderingContent(content);
  if (!parsed) return <RawContent content={content} />;

  return (
    <div className="space-y-1 text-sm">
      <p>
        <span className="font-medium text-muted-foreground">[{parsed.source.lang}]</span>{" "}
        {parsed.source.text}
      </p>
      <div className="flex items-center gap-1.5">
        <span className="font-medium text-muted-foreground">[{parsed.target.lang}]</span>
        <OrderingPreview
          segments={parsed.target.segments}
          lang={parsed.target.lang}
          label={t("workbooks.study.ordering.correctOrder")}
        />
      </div>
      {parsed.explanation && (
        <MarkdownText text={parsed.explanation} className="text-xs text-muted-foreground" />
      )}
    </div>
  );
}

type CardKind = {
  badge: BadgeStyle;
  Body: (props: { content: string }) => ReactNode;
};

const CARD_KINDS = new Map<string, CardKind>([
  [
    "word_fill",
    {
      badge: {
        labelKey: "workbooks.addQuestion.wordFill",
        className:
          "rounded-full bg-blue-100 px-2 py-0.5 text-[11px] font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
      },
      Body: WordFillCardBody,
    },
  ],
  [
    "multiple_choice",
    {
      badge: {
        labelKey: "workbooks.addQuestion.multipleChoice",
        className:
          "rounded-full bg-purple-100 px-2 py-0.5 text-[11px] font-medium text-purple-700 dark:bg-purple-900/30 dark:text-purple-400",
      },
      Body: MultipleChoiceCardBody,
    },
  ],
  [
    "ordering",
    {
      badge: {
        labelKey: "workbooks.addQuestion.ordering",
        className:
          "rounded-full bg-amber-100 px-2 py-0.5 text-[11px] font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-400",
      },
      Body: OrderingCardBody,
    },
  ],
]);

function UnknownQuestionCard({ question }: { question: Question }) {
  return (
    <div className="rounded-lg border bg-card p-4 shadow-sm">
      <div className="mb-2">
        <span className="rounded-full bg-green-100 px-2 py-0.5 text-[11px] font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">
          {question.questionType}
        </span>
      </div>
      <RawContent content={question.content} />
    </div>
  );
}

export function QuestionCard({
  question,
  onEdit,
}: {
  question: Question;
  onEdit: (question: Question) => void;
}) {
  const kind = CARD_KINDS.get(question.questionType);
  if (!kind) return <UnknownQuestionCard question={question} />;

  return (
    <CardShell badge={kind.badge} question={question} onEdit={onEdit}>
      <kind.Body content={question.content} />
    </CardShell>
  );
}
