import { type RefObject, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { MarkdownText } from "~/components/markdown-text";
import { getQuestionModule, isQuestionType } from "~/components/study/question-types/registry";
import { Button } from "~/components/ui/button";
import type { QuestionType } from "~/components/workbook/schemas";
import type { StudyQuestion } from "~/lib/api/study.server";
import type { AnswerDraft, QuestionModule } from "./types";

type StudyQuestionShellProps = {
  question: StudyQuestion;
  onAnswer: (draft: AnswerDraft) => void;
};

export function StudyQuestionShell({ question, onAnswer }: StudyQuestionShellProps) {
  if (!isQuestionType(question.questionType)) {
    return <UnrenderableQuestion content={question.content} />;
  }
  return (
    <ParsedQuestion
      module={getQuestionModule(question.questionType)}
      question={question}
      onAnswer={onAnswer}
    />
  );
}

function UnrenderableQuestion({ content }: { content: string }) {
  return <p className="text-sm text-muted-foreground">{content}</p>;
}

type ParsedQuestionProps<T extends QuestionType, C> = {
  module: QuestionModule<T, C>;
  question: StudyQuestion;
  onAnswer: (draft: AnswerDraft<T>) => void;
};

function ParsedQuestion<T extends QuestionType, C>({
  module,
  question,
  onAnswer,
}: ParsedQuestionProps<T, C>) {
  const content = useMemo(() => module.parse(question.content), [module, question.content]);
  if (content === null) {
    return <UnrenderableQuestion content={question.content} />;
  }
  return <QuestionFlow module={module} content={content} question={question} onAnswer={onAnswer} />;
}

type FlowState<T extends QuestionType> =
  | { phase: "answering"; draft: AnswerDraft<T> | null }
  | { phase: "revealed"; draft: AnswerDraft<T> };

function QuestionFlow<T extends QuestionType, C>({
  module,
  content,
  question,
  onAnswer,
}: ParsedQuestionProps<T, C> & { content: C }) {
  const { t } = useTranslation();
  const [state, setState] = useState<FlowState<T>>(() => ({
    phase: "answering",
    draft: module.initialDraft(content),
  }));
  const nextButtonRef = useRef<HTMLButtonElement | null>(null);

  useEffect(() => {
    if (state.phase === "revealed") nextButtonRef.current?.focus();
  }, [state.phase]);

  function handleDraftChange(draft: AnswerDraft<T> | null) {
    setState((prev) => (prev.phase === "answering" ? { phase: "answering", draft } : prev));
  }

  function handleCommit(draft: AnswerDraft<T>) {
    setState({ phase: "revealed", draft });
  }

  function handleReveal() {
    if (state.draft !== null) setState({ phase: "revealed", draft: state.draft });
  }

  const explanation = module.explanation(content, state.phase);
  const Body = module.Body;

  return (
    <div className="space-y-6">
      <Body
        content={content}
        audio={question.audio}
        phase={state.phase}
        onDraftChange={handleDraftChange}
        onCommit={handleCommit}
      />

      {explanation && <MarkdownText text={explanation} className="text-sm text-muted-foreground" />}

      {state.phase === "answering" ? (
        <div className="flex justify-end">
          <Button onClick={handleReveal} disabled={state.draft === null}>
            {t(module.revealLabelKey)}
          </Button>
        </div>
      ) : (
        <ResultFooter
          correct={state.draft.correct}
          nextButtonRef={nextButtonRef}
          onNext={() => onAnswer(state.draft)}
        />
      )}
    </div>
  );
}

function ResultFooter({
  correct,
  nextButtonRef,
  onNext,
}: {
  correct: boolean;
  nextButtonRef: RefObject<HTMLButtonElement | null>;
  onNext: () => void;
}) {
  const { t } = useTranslation();
  return (
    <div className="flex items-center justify-end gap-3">
      <span className={`text-sm font-medium ${correct ? "text-green-600" : "text-red-600"}`}>
        {correct ? t("workbooks.study.correct") : t("workbooks.study.incorrect")}
      </span>
      <Button ref={nextButtonRef} onClick={onNext}>
        {t("workbooks.study.next")}
      </Button>
    </div>
  );
}
