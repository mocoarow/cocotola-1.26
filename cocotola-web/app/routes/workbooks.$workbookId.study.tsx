import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Link,
  type ShouldRevalidateFunctionArgs,
  useFetcher,
  useLoaderData,
} from "react-router";
import { MultipleChoiceCard } from "~/components/study/multiple-choice-card";
import { ProgressBar } from "~/components/study/progress-bar";
import { StudyResult } from "~/components/study/study-result";
import { WordFillCard } from "~/components/study/word-fill-card";
import { Button } from "~/components/ui/button";
import {
  getStudyQuestions,
  recordAnswerForMultipleChoice,
  recordAnswerForWordFill,
  type StudyQuestion,
} from "~/lib/api/study.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/workbooks.$workbookId.study";

export async function loader({ request, params }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const data = await getStudyQuestions(accessToken, workbookId, 20);
  return { workbookId, questions: data.questions };
}

// Skip revalidation only for the "answer" action submit. Otherwise the loader
// reruns after every answer and prunes already-answered questions, leaving the
// component reading questions[currentIndex] after the array has shrunk past
// the local index — crashing the study screen mid-session. Navigation and any
// other revalidation triggers fall through to the default behavior so that
// re-entering the route still picks up a fresh question queue.
export function shouldRevalidate({
  formData,
  defaultShouldRevalidate,
}: ShouldRevalidateFunctionArgs) {
  if (formData?.get("intent") === "answer") return false;
  return defaultShouldRevalidate;
}

export async function action({ request, params }: Route.ActionArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const formData = await request.formData();
  const intent = formData.get("intent");

  if (intent !== "answer") return { ok: false };

  const questionId = String(formData.get("questionId") ?? "");
  const questionType = String(formData.get("questionType") ?? "");

  if (questionType === "multiple_choice") {
    const raw = String(formData.get("selectedChoiceIds") ?? "[]");
    let parsed: unknown;
    try {
      parsed = JSON.parse(raw);
    } catch {
      throw new Response("selectedChoiceIds must be valid JSON", { status: 400 });
    }
    if (!Array.isArray(parsed) || !parsed.every((v): v is string => typeof v === "string")) {
      throw new Response("selectedChoiceIds must be an array of strings", { status: 400 });
    }
    const result = await recordAnswerForMultipleChoice(accessToken, workbookId, questionId, parsed);
    return { ok: true, result };
  }

  const correct = formData.get("correct") === "true";
  const result = await recordAnswerForWordFill(accessToken, workbookId, questionId, correct);
  return { ok: true, result };
}

type Phase = "studying" | "done";

function StudySession({
  questions,
  workbookId,
}: {
  questions: StudyQuestion[];
  workbookId: string;
}) {
  const { t } = useTranslation();
  const fetcher = useFetcher();
  const [currentIndex, setCurrentIndex] = useState(0);
  const [correctCount, setCorrectCount] = useState(0);
  const [incorrectCount, setIncorrectCount] = useState(0);
  const [phase, setPhase] = useState<Phase>("studying");

  if (questions.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16">
        <p className="mb-4 text-lg text-muted-foreground">{t("workbooks.study.noQuestions")}</p>
        <Button nativeButton={false} render={<Link to={`/workbooks/${workbookId}`} />}>
          {t("workbooks.study.backToWorkbook")}
        </Button>
      </div>
    );
  }

  if (phase === "done") {
    return (
      <StudyResult
        correctCount={correctCount}
        incorrectCount={incorrectCount}
        workbookId={workbookId}
      />
    );
  }

  const question = questions[currentIndex];

  function advance(correct: boolean) {
    if (correct) {
      setCorrectCount((c) => c + 1);
    } else {
      setIncorrectCount((c) => c + 1);
    }
    if (currentIndex + 1 < questions.length) {
      setCurrentIndex((i) => i + 1);
    } else {
      setPhase("done");
    }
  }

  function handleMultipleChoiceAnswer(selectedChoiceIds: string[], correct: boolean) {
    fetcher.submit(
      {
        intent: "answer",
        questionId: question.questionId,
        questionType: "multiple_choice",
        selectedChoiceIds: JSON.stringify(selectedChoiceIds),
      },
      { method: "post" },
    );
    advance(correct);
  }

  function handleWordFillAnswer(correct: boolean) {
    fetcher.submit(
      {
        intent: "answer",
        questionId: question.questionId,
        questionType: "word_fill",
        correct: String(correct),
      },
      { method: "post" },
    );
    advance(correct);
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <ProgressBar current={currentIndex} total={questions.length} />

      <div className="rounded-xl border bg-card p-6 shadow-sm">
        {question.questionType === "multiple_choice" ? (
          <MultipleChoiceCard
            key={question.questionId}
            content={question.content}
            onAnswer={handleMultipleChoiceAnswer}
          />
        ) : (
          <WordFillCard
            key={question.questionId}
            content={question.content}
            onAnswer={handleWordFillAnswer}
          />
        )}
      </div>
    </div>
  );
}

export default function StudyPage() {
  const { workbookId, questions } = useLoaderData<typeof loader>();
  const { t } = useTranslation();

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold">{t("workbooks.study.title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {t("workbooks.study.description", { count: questions.length })}
        </p>
      </div>
      <StudySession questions={questions} workbookId={workbookId} />
    </div>
  );
}
