import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Link,
  type ShouldRevalidateFunctionArgs,
  useFetcher,
  useLoaderData,
  useNavigate,
  useRouteLoaderData,
} from "react-router";
import { MultipleChoiceCard } from "~/components/study/multiple-choice-card";
import { ProgressBar } from "~/components/study/progress-bar";
import { StudyResult } from "~/components/study/study-result";
import { WordFillCard } from "~/components/study/word-fill-card";
import { Button } from "~/components/ui/button";
import { useStudyResume } from "~/hooks/use-study-resume";
import {
  getStudyQuestions,
  recordAnswerForMultipleChoice,
  recordAnswerForWordFill,
  type StudyQuestion,
} from "~/lib/api/study.server";
import { getWorkbook } from "~/lib/api/workbook.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import {
  appendAnsweredId,
  clampExcludeIds,
  clearStudySession,
  studyRemountKey,
} from "~/lib/study-session";
import type { Route } from "./+types/workbooks.$workbookId.study";
import type { loader as workbooksLayoutLoader } from "./workbooks";

// Default session size when the user lands without choosing one (e.g. the
// study URL shared from another tab). The dialog on the workbook list always
// supplies an explicit `?limit=N` so this only applies to direct navigation.
const DEFAULT_STUDY_LIMIT = 20;

export async function loader({ request, params }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const url = new URL(request.url);
  const practice = url.searchParams.get("practice") === "true";
  const limit = parseLimit(url.searchParams.get("limit"));
  // Defense in depth: drop entries the server would 400 (empty, too long,
  // over the count cap) before we issue the API call. The server validates
  // again at the trust boundary.
  const excludeIds = clampExcludeIds(url.searchParams.getAll("excludeIds"));
  const [workbook, data] = await Promise.all([
    getWorkbook(accessToken, workbookId),
    getStudyQuestions(accessToken, workbookId, limit, practice, excludeIds),
  ]);
  return {
    workbookId,
    workbookOwnerId: workbook.ownerId,
    questions: data.questions,
    practice,
    excludeIds,
  };
}

function parseLimit(raw: string | null): number {
  if (raw === null) return DEFAULT_STUDY_LIMIT;
  const parsed = Number.parseInt(raw, 10);
  if (!Number.isFinite(parsed) || parsed <= 0) return DEFAULT_STUDY_LIMIT;
  return Math.min(100, parsed);
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

function StudyResumePlaceholder() {
  const { t } = useTranslation();
  return (
    <div
      role="status"
      aria-live="polite"
      data-testid="study-resume-pending"
      className="mx-auto flex max-w-2xl flex-col items-center justify-center py-16 text-sm text-muted-foreground"
    >
      {t("workbooks.study.resuming")}
    </div>
  );
}

function StudySession({
  userId,
  workbookId,
  questions,
  practice,
  loaderExcludeIds,
  backUrl,
  backLabel,
}: {
  userId: string | null;
  workbookId: string;
  questions: StudyQuestion[];
  practice: boolean;
  loaderExcludeIds: string[];
  backUrl: string;
  backLabel: string;
}) {
  const { t } = useTranslation();
  const fetcher = useFetcher();
  const navigate = useNavigate();
  const [queue, setQueue] = useState<StudyQuestion[]>(() => questions);
  const [correctCount, setCorrectCount] = useState(0);
  const [incorrectCount, setIncorrectCount] = useState(0);
  const [attemptCounts, setAttemptCounts] = useState<Record<string, number>>({});
  // "redirecting" suppresses the answer card during the navigate→remount
  // window so the user cannot click a question that the resume orchestrator
  // is about to remove. Without this gate, a fast click on the pre-resume
  // first card fires fetcher.submit and double-updates the SRS state.
  const resumeStatus = useStudyResume({
    userId,
    workbookId,
    practice,
    loaderExcludeIds,
    navigate,
  });

  // Resume gate: hide the answer card (and its click handlers) while the
  // navigate-driven loader rerun is in flight, so the user cannot
  // double-submit an already-answered question.
  if (resumeStatus === "redirecting") {
    return <StudyResumePlaceholder />;
  }

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
      // Persist completion so a F5 mid-session resumes without re-showing
      // questions the user has finished. Incorrect attempts are not stored:
      // they get re-queued in memory and the user will retry within this
      // session; if they reload before retrying, the question reappears
      // fresh (matching the pre-resume behavior). Skip the persistence side
      // when userId is unknown so we never write under an ambiguous scope.
      const scope = userId === null ? null : { userId, workbookId, practice };
      if (scope !== null) {
        appendAnsweredId(scope, question.questionId, Date.now());
      }
      // Read the queue length here, BEFORE setQueue, so the "session done"
      // decision lives outside the setState updater. The updater itself must
      // stay pure — StrictMode runs it twice — even though clearStudySession
      // happens to be idempotent today.
      const finishesSession = queue.length === 1;
      setCorrectCount((c) => c + 1);
      setQueue((q) => q.slice(1));
      if (scope !== null && finishesSession) {
        clearStudySession(scope);
      }
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
          <WordFillCard
            key={cardKey}
            content={question.content}
            audio={question.audio}
            onAnswer={handleWordFillAnswer}
          />
        )}
      </div>
    </div>
  );
}

export default function StudyPage() {
  const { workbookId, workbookOwnerId, questions, practice, excludeIds } =
    useLoaderData<typeof loader>();
  const layoutData = useRouteLoaderData<typeof workbooksLayoutLoader>("routes/workbooks");
  const { t } = useTranslation();

  // Normalise undefined / "" to null at the boundary so downstream code only
  // has to check for null when deciding whether the session helpers may run.
  const rawUserId = layoutData?.user?.userId;
  const currentUserId = rawUserId === undefined || rawUserId === "" ? null : rawUserId;
  const isOwner = currentUserId === workbookOwnerId;
  const backUrl = isOwner ? `/workbooks/${workbookId}` : "/workbooks/public";
  const backLabel = isOwner
    ? t("workbooks.study.backToWorkbook")
    : t("workbooks.study.backToPublic");

  // Recompute only when the inputs change. excludeIds is a fresh array
  // reference on every loader rerun so the memo key naturally invalidates at
  // the same rate the underlying value changes.
  const remountKey = useMemo(
    () => studyRemountKey(currentUserId, practice, excludeIds),
    [currentUserId, practice, excludeIds],
  );

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
        // The key folds in both the mode and the loader's excludeIds so that:
        //   1. Toggling normal <-> practice forces a remount (otherwise the
        //      StudySession's useState seed runs once and the Continue-
        //      practicing CTA inherits the old empty queue).
        //   2. A resume-driven navigate (mount effect appends excludeIds and
        //      triggers a loader rerun) also forces a remount so the queue is
        //      re-seeded from the post-exclusion question list. Without the
        //      excludeIds segment the queue retained the pre-exclusion items
        //      and re-showed questions the user already answered.
        key={remountKey}
        userId={currentUserId}
        workbookId={workbookId}
        questions={questions}
        practice={practice}
        loaderExcludeIds={excludeIds}
        backUrl={backUrl}
        backLabel={backLabel}
      />
    </div>
  );
}
