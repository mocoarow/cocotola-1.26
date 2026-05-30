import { BugIcon, Trash2Icon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useFetcher, useLoaderData } from "react-router";
import { useConfirm } from "~/components/confirm-dialog-provider";
import { Button } from "~/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "~/components/ui/sheet";
import { AddQuestionSection } from "~/components/workbook/add-question-section";
import { MultipleChoiceEditForm } from "~/components/workbook/multiple-choice-edit-form";
import { QuestionCard } from "~/components/workbook/question-card";
import {
  parseMultipleChoiceContent,
  parseMultipleChoiceFormData,
  parseWordFillContent,
} from "~/components/workbook/schemas";
import { WordFillEditForm } from "~/components/workbook/word-fill-edit-form";
import { WorkbookHeader } from "~/components/workbook/workbook-header";
import type { Question } from "~/lib/api/question.server";
import {
  addQuestion,
  deleteQuestion,
  listQuestions,
  updateQuestion,
} from "~/lib/api/question.server";
import { deleteStudyHistory } from "~/lib/api/study.server";
import { getWorkbook, updateWorkbook } from "~/lib/api/workbook.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/workbooks.$workbookId";

export async function loader({ request, params }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const [workbook, questions] = await Promise.all([
    getWorkbook(accessToken, workbookId),
    listQuestions(accessToken, workbookId),
  ]);
  return { workbook, questions };
}

export async function action({ request, params }: Route.ActionArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const formData = await request.formData();
  const intent = formData.get("intent");

  if (intent === "updateTitle") {
    const title = String(formData.get("title") ?? "").trim();
    if (!title) {
      return { ok: false, errorKey: "workbooks.detail.errors.titleRequired" };
    }
    if (title.length > 200) {
      return { ok: false, errorKey: "workbooks.detail.errors.titleTooLong" };
    }
    const description = String(formData.get("description") ?? "");
    const rawVisibility = String(formData.get("visibility") ?? "private");
    const visibility = rawVisibility === "public" ? "public" : "private";
    await updateWorkbook(accessToken, workbookId, { title, description, visibility });
    return { ok: true };
  }

  if (intent === "updateQuestion") {
    const questionId = formData.get("questionId");
    if (typeof questionId !== "string" || !questionId) {
      throw new Response("questionId is required", { status: 400 });
    }

    const questionType = formData.get("questionType");

    if (questionType === "multiple_choice") {
      const { content, tags } = parseMultipleChoiceFormData(formData);
      const orderIndex = Number(formData.get("orderIndex") ?? 0);

      await updateQuestion(accessToken, workbookId, questionId, {
        content,
        tags: tags.length > 0 ? tags : undefined,
        orderIndex,
      });
      return { ok: true };
    }

    if (questionType === "word_fill" || !questionType) {
      const sourceText = formData.get("sourceText");
      const sourceLang = formData.get("sourceLang");
      const targetText = formData.get("targetText");
      const targetLang = formData.get("targetLang");

      if (
        typeof sourceText !== "string" ||
        !sourceText.trim() ||
        typeof sourceLang !== "string" ||
        !sourceLang.trim() ||
        typeof targetText !== "string" ||
        !targetText.trim() ||
        typeof targetLang !== "string" ||
        !targetLang.trim()
      ) {
        throw new Response("sourceText, sourceLang, targetText, and targetLang are required", {
          status: 400,
        });
      }

      const explanation = formData.get("explanation");
      const explanation1 = formData.get("explanation1");
      const explanation2 = formData.get("explanation2");
      const content = JSON.stringify({
        source: { text: sourceText, lang: sourceLang },
        target: { text: targetText, lang: targetLang },
        ...(typeof explanation === "string" && explanation.trim() ? { explanation } : {}),
        ...(typeof explanation1 === "string" && explanation1.trim() ? { explanation1 } : {}),
        ...(typeof explanation2 === "string" && explanation2.trim() ? { explanation2 } : {}),
      });

      const tagsRaw = formData.get("tags");
      const tags =
        typeof tagsRaw === "string"
          ? tagsRaw
              .split(",")
              .map((t) => t.trim())
              .filter(Boolean)
          : [];

      const orderIndex = Number(formData.get("orderIndex") ?? 0);

      await updateQuestion(accessToken, workbookId, questionId, {
        content,
        tags: tags.length > 0 ? tags : undefined,
        orderIndex,
      });
      return { ok: true };
    }

    throw new Response(`Unknown questionType: ${String(questionType)}`, { status: 400 });
  }

  if (intent === "deleteQuestion") {
    const questionId = formData.get("questionId");
    if (typeof questionId !== "string" || !questionId) {
      throw new Response("questionId is required", { status: 400 });
    }
    await deleteQuestion(accessToken, workbookId, questionId);
    return { ok: true };
  }

  if (intent === "clearStudyHistory") {
    await deleteStudyHistory(accessToken, workbookId);
    return { ok: true, cleared: true };
  }

  if (intent === "addWordFill") {
    const sourceText = formData.get("sourceText");
    const sourceLang = formData.get("sourceLang");
    const targetText = formData.get("targetText");
    const targetLang = formData.get("targetLang");

    if (
      typeof sourceText !== "string" ||
      !sourceText.trim() ||
      typeof sourceLang !== "string" ||
      !sourceLang.trim() ||
      typeof targetText !== "string" ||
      !targetText.trim() ||
      typeof targetLang !== "string" ||
      !targetLang.trim()
    ) {
      throw new Response("sourceText, sourceLang, targetText, and targetLang are required", {
        status: 400,
      });
    }

    const explanation = formData.get("explanation");

    const content = JSON.stringify({
      source: { text: sourceText, lang: sourceLang },
      target: { text: targetText, lang: targetLang },
      ...(typeof explanation === "string" && explanation.trim() ? { explanation } : {}),
    });

    const tagsRaw = formData.get("tags");
    const tags =
      typeof tagsRaw === "string"
        ? tagsRaw
            .split(",")
            .map((t) => t.trim())
            .filter(Boolean)
        : [];

    await addQuestion(accessToken, workbookId, {
      questionType: "word_fill",
      content,
      tags: tags.length > 0 ? tags : undefined,
      orderIndex: 0,
    });
    return { ok: true, added: true };
  }

  if (intent === "addMultipleChoice") {
    const { content, tags } = parseMultipleChoiceFormData(formData);

    await addQuestion(accessToken, workbookId, {
      questionType: "multiple_choice",
      content,
      tags: tags.length > 0 ? tags : undefined,
      orderIndex: 0,
    });
    return { ok: true, added: true };
  }

  throw new Response(`Unknown intent: ${String(intent)}`, { status: 400 });
}

export default function WorkbookDetail() {
  const { workbook, questions } = useLoaderData<typeof loader>();
  const { t } = useTranslation();
  const [editingQuestion, setEditingQuestion] = useState<Question | null>(null);
  const editFetcher = useFetcher<{ ok: boolean }>();
  const clearHistoryFetcher = useFetcher<{ ok: boolean; cleared?: boolean }>();
  const confirm = useConfirm();
  const [showCleared, setShowCleared] = useState(false);

  const saved = useRef(false);
  if (editFetcher.data?.ok && !saved.current) {
    saved.current = true;
    setEditingQuestion(null);
  }
  if (editFetcher.state === "submitting") {
    saved.current = false;
  }

  useEffect(() => {
    if (clearHistoryFetcher.data?.cleared) {
      setShowCleared(true);
      const timer = setTimeout(() => setShowCleared(false), 3000);
      return () => clearTimeout(timer);
    }
  }, [clearHistoryFetcher.data]);

  async function handleClearStudyHistory() {
    const confirmed = await confirm({
      title: t("workbooks.detail.clearStudyHistoryConfirmTitle"),
      description: t("workbooks.detail.clearStudyHistoryConfirm"),
      confirmLabel: t("workbooks.detail.clearStudyHistory"),
    });
    if (confirmed) {
      clearHistoryFetcher.submit({ intent: "clearStudyHistory" }, { method: "post" });
    }
  }

  const sheetTitle =
    editingQuestion?.questionType === "word_fill"
      ? t("workbooks.addQuestion.titleWordFill")
      : t("workbooks.addQuestion.titleMultipleChoice");

  const isClearing = clearHistoryFetcher.state !== "idle";

  return (
    <div>
      <WorkbookHeader
        title={workbook.title}
        description={workbook.description}
        visibility={workbook.visibility}
      />

      <div className="mb-4 flex flex-wrap items-center gap-2">
        <Button variant="outline" size="sm" onClick={handleClearStudyHistory} disabled={isClearing}>
          <Trash2Icon data-icon="inline-start" className="size-3.5 text-destructive" />
          <span>{t("workbooks.detail.clearStudyHistory")}</span>
        </Button>
        {import.meta.env.DEV && (
          <Button variant="ghost" size="sm" asChild>
            <Link to={`/workbooks/${workbook.workbookId}/debug-study`}>
              <BugIcon data-icon="inline-start" className="size-3.5" />
              <span>{t("workbooks.debugStudy.title")}</span>
            </Link>
          </Button>
        )}
        {showCleared && (
          <span
            role="status"
            className="text-sm text-muted-foreground"
            data-testid="clear-study-history-status"
          >
            {t("workbooks.detail.clearStudyHistoryDone")}
          </span>
        )}
      </div>

      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-lg font-semibold">{t("workbooks.detail.questionsTitle")}</h2>
        <AddQuestionSection />
      </div>

      {questions.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16">
          <p className="text-lg font-medium text-muted-foreground">
            {t("workbooks.detail.emptyTitle")}
          </p>
          <p className="mt-1 text-sm text-muted-foreground/70">
            {t("workbooks.detail.emptyDescription")}
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {questions.map((question) => (
            <QuestionCard
              key={question.questionId}
              question={question}
              onEdit={setEditingQuestion}
            />
          ))}
        </div>
      )}

      <Sheet
        open={editingQuestion !== null}
        onOpenChange={(open) => {
          if (!open) setEditingQuestion(null);
        }}
      >
        <SheetContent side="right">
          <SheetHeader>
            <SheetTitle>{sheetTitle}</SheetTitle>
            <SheetDescription>{t("workbooks.detail.editQuestion")}</SheetDescription>
          </SheetHeader>
          {editingQuestion?.questionType === "word_fill" && (
            <WordFillEditForm
              question={editingQuestion}
              parsed={parseWordFillContent(editingQuestion.content) ?? {}}
              fetcher={editFetcher}
              onCancel={() => setEditingQuestion(null)}
            />
          )}
          {editingQuestion?.questionType === "multiple_choice" && (
            <MultipleChoiceEditForm
              question={editingQuestion}
              parsed={parseMultipleChoiceContent(editingQuestion.content) ?? {}}
              fetcher={editFetcher}
              onCancel={() => setEditingQuestion(null)}
            />
          )}
        </SheetContent>
      </Sheet>
    </div>
  );
}
