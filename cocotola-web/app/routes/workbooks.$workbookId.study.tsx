import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Link,
  type ShouldRevalidateFunctionArgs,
  useFetcher,
  useLoaderData,
  useRouteLoaderData,
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
import { getWorkbook } from "~/lib/api/workbook.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/workbooks.$workbookId.study";
import type { loader as workbooksLayoutLoader } from "./workbooks";

export async function loader({ request, params }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const url = new URL(request.url);
  const practice = url.searchParams.get("practice") === "true";
  const [workbook, data] = await Promise.all([
    getWorkbook(accessToken, workbookId),
    getStudyQuestions(accessToken, workbookId, 20, practice),
  ]);
  return {
    workbookId,
    workbookOwnerId: workbook.ownerId,
    questions: data.questions,
    practice,
  };
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

  // Practice mode: don't persist answers — the user is past the day's
  // SRS-due queue and just wants to keep solving without affecting their
  // spaced-repetition counters. Trust the request URL (set by the loader
  // when the page was opened) over a client-supplied form field so a
  // tampered submit body cannot bypass persistence in normal mode.
  const practice = new URL(request.url).searchParams.get("practice") === "true";
  if (practice) {
    return { ok: true, practice: true };
  }

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
  workbookId,
  questions,
  practice,
  backUrl,
  backLabel,
}: {
  workbookId: string;
  questions: StudyQuestion[];
  practice: boolean;
  backUrl: string;
  backLabel: string;
}) {
  const { t } = useTranslation();
  const fetcher = useFetcher();
  const [queue, setQueue] = useState<StudyQuestion[]>(() => questions);
  const [correctCount, setCorrectCount] = useState(0);
  const [incorrectCount, setIncorrectCount] = useState(0);
  const [attemptCounts, setAttemptCounts] = useState<Record<string, number>>({});

  // Structural empty-state guard: derived from the loader prop, not the local
  // queue. The queue can also reach length 0 (after the last correct answer)
  // but that case must render the result screen, not this empty state.
  if (questions.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16">
        <p className="mb-4 text-lg text-muted-foreground">{t("workbooks.study.noQuestions")}</p>
        <div className="flex flex-col items-center gap-3 sm:flex-row">
          {!practice && (
            <Button
              variant="outline"
              nativeButton={false}
              render={<Link to={`/workbooks/${workbookId}/study?practice=true`} />}
            >
              {t("workbooks.study.practiceCta")}
            </Button>
          )}
          <Button nativeButton={false} render={<Link to={backUrl} />}>
            {backLabel}
          </Button>
        </div>
      </div>
    );
  }

  const phase: Phase = queue.length === 0 ? "done" : "studying";

  if (phase === "done") {
    return (
      <StudyResult
        correctCount={correctCount}
        incorrectCount={incorrectCount}
        backUrl={backUrl}
        backLabel={backLabel}
      />
    );
  }

  const question = queue[0];

  function advance(correct: boolean) {
    setAttemptCounts((prev) => ({
      ...prev,
      [question.questionId]: (prev[question.questionId] ?? 0) + 1,
    }));
    if (correct) {
      setCorrectCount((c) => c + 1);
      setQueue((q) => q.slice(1));
    } else {
      setIncorrectCount((c) => c + 1);
      setQueue((q) => {
        const [head, ...rest] = q;
        return [...rest, head];
      });
    }
  }

  // Preserve the practice flag in the action URL itself so the server can
  // verify the mode independently of client-supplied form fields.
  const actionPath = practice
    ? `/workbooks/${workbookId}/study?practice=true`
    : `/workbooks/${workbookId}/study`;

  function handleMultipleChoiceAnswer(selectedChoiceIds: string[], correct: boolean) {
    fetcher.submit(
      {
        intent: "answer",
        questionId: question.questionId,
        questionType: "multiple_choice",
        selectedChoiceIds: JSON.stringify(selectedChoiceIds),
      },
      { method: "post", action: actionPath },
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
      { method: "post", action: actionPath },
    );
    advance(correct);
  }

  const cardKey = `${question.questionId}-${attemptCounts[question.questionId] ?? 0}`;

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <ProgressBar current={correctCount} total={questions.length} />

      <div className="rounded-xl border bg-card p-6 shadow-sm">
        {question.questionType === "multiple_choice" ? (
          <MultipleChoiceCard
            key={cardKey}
            content={question.content}
            onAnswer={handleMultipleChoiceAnswer}
          />
        ) : (
          <WordFillCard key={cardKey} content={question.content} onAnswer={handleWordFillAnswer} />
        )}
      </div>
    </div>
  );
}

export default function StudyPage() {
  const { workbookId, workbookOwnerId, questions, practice } = useLoaderData<typeof loader>();
  const layoutData = useRouteLoaderData<typeof workbooksLayoutLoader>("routes/workbooks");
  const { t } = useTranslation();

  const isOwner = layoutData?.user?.userId === workbookOwnerId;
  const backUrl = isOwner ? `/workbooks/${workbookId}` : "/workbooks/public";
  const backLabel = isOwner
    ? t("workbooks.study.backToWorkbook")
    : t("workbooks.study.backToPublic");

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold">
          {practice ? t("workbooks.study.practiceTitle") : t("workbooks.study.title")}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {t("workbooks.study.description", { count: questions.length })}
        </p>
        {practice && (
          <div
            role="status"
            className="mt-3 rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-700 dark:bg-amber-950/40 dark:text-amber-200"
          >
            {t("workbooks.study.practiceBanner")}
          </div>
        )}
      </div>
      <StudySession
        // Force remount when toggling between normal and practice modes.
        // StudySession seeds its `queue` from `questions` via useState, which
        // only runs once per mount — without a key change, navigating from
        // the empty-state Continue-practicing CTA reuses the old (zero-length)
        // queue and immediately renders the "Session Complete! 0%" result
        // screen against the freshly loaded 10 questions.
        key={practice ? "practice" : "normal"}
        workbookId={workbookId}
        questions={questions}
        practice={practice}
        backUrl={backUrl}
        backLabel={backLabel}
      />
    </div>
  );
}
