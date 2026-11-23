import { BugIcon, Trash2Icon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { data, Link, useFetcher, useLoaderData } from "react-router";
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
import { parseOrderingFormData } from "~/components/workbook/ordering/form-data";
import { OrderingEditForm } from "~/components/workbook/ordering/ordering-edit-form";
import { QuestionCard } from "~/components/workbook/question-card";
import {
  QUESTION_INVALID_ERROR_KEY,
  type QuestionActionData,
} from "~/components/workbook/question-form-error";
import {
  parseMultipleChoiceContent,
  parseMultipleChoiceFormData,
  parseOrderingContent,
  parseTagsFormData,
  parseWordFillContent,
  type QuestionType,
} from "~/components/workbook/schemas";
import { WordFillEditForm } from "~/components/workbook/word-fill-edit-form";
import { WorkbookHeader } from "~/components/workbook/workbook-header";
import { type BackendClient, createBackendClient } from "~/lib/api/client.server";
import type { Question } from "~/lib/api/question.server";
import {
  addQuestion,
  deleteQuestion,
  listQuestions,
  QuestionValidationError,
  updateQuestion,
} from "~/lib/api/question.server";
import { deleteStudyHistory } from "~/lib/api/study.server";
import { getWorkbook, updateWorkbook, WorkbookValidationError } from "~/lib/api/workbook.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/workbooks.$workbookId";

export async function loader({ request, params, context }: Route.LoaderArgs) {
  const { accessToken } = requireAuth(context);
  const client = createBackendClient({ request, accessToken });
  const { workbookId } = params;
  const [workbook, questions] = await Promise.all([
    getWorkbook(client, workbookId),
    listQuestions(client, workbookId),
  ]);
  return { workbook, questions };
}

async function saveQuestion<T>(save: () => Promise<unknown>, saved: T) {
  try {
    await save();
  } catch (error) {
    if (error instanceof QuestionValidationError) {
      return data({ ok: false, errorKey: QUESTION_INVALID_ERROR_KEY }, { status: 400 });
    }
    throw error;
  }
  return saved;
}

async function handleUpdateTitleIntent(
  client: BackendClient,
  workbookId: string,
  formData: FormData,
) {
  const title = String(formData.get("title") ?? "").trim();
  if (!title) {
    return { ok: false, errorKey: "workbooks.detail.errors.titleRequired" };
  }
  if (title.length > 200) {
    return { ok: false, errorKey: "workbooks.detail.errors.titleTooLong" };
  }
  const language = String(formData.get("language") ?? "");
  if (!language) {
    return { ok: false, errorKey: "workbooks.detail.errors.invalid" };
  }
  const description = String(formData.get("description") ?? "");
  const rawVisibility = String(formData.get("visibility") ?? "private");
  const visibility = rawVisibility === "public" ? "public" : "private";
  try {
    await updateWorkbook(client, workbookId, { title, description, visibility, language });
  } catch (error) {
    if (error instanceof WorkbookValidationError) {
      return { ok: false, errorKey: "workbooks.detail.errors.invalid" };
    }
    throw error;
  }
  return { ok: true };
}

function optionalTags(tags: string[]) {
  return tags.length > 0 ? tags : undefined;
}

function requireQuestionId(formData: FormData) {
  const questionId = formData.get("questionId");
  if (typeof questionId !== "string" || !questionId) {
    throw new Response("questionId is required", { status: 400 });
  }
  return questionId;
}

function nonBlankText(formData: FormData, key: string) {
  const value = formData.get(key);
  return typeof value === "string" && value.trim() ? value : undefined;
}

function optionalTextEntry(formData: FormData, key: string) {
  const value = nonBlankText(formData, key);
  return value === undefined ? {} : { [key]: value };
}

function parseWordFillPair(formData: FormData) {
  const sourceText = nonBlankText(formData, "sourceText");
  const sourceLang = nonBlankText(formData, "sourceLang");
  const targetText = nonBlankText(formData, "targetText");
  const targetLang = nonBlankText(formData, "targetLang");
  if (!(sourceText && sourceLang && targetText && targetLang)) {
    throw new Response("sourceText, sourceLang, targetText, and targetLang are required", {
      status: 400,
    });
  }
  return {
    source: { text: sourceText, lang: sourceLang },
    target: { text: targetText, lang: targetLang },
  };
}

function parseWordFillFormData(formData: FormData, explanationKeys: string[]) {
  const pair = parseWordFillPair(formData);
  const explanations = explanationKeys.map((key) => optionalTextEntry(formData, key));
  const content = JSON.stringify(Object.assign({}, pair, ...explanations));
  return { content, tags: parseTagsFormData(formData) };
}

function parseUpdateQuestionFormData(formData: FormData) {
  const questionType = formData.get("questionType");
  if (questionType === "multiple_choice") {
    return parseMultipleChoiceFormData(formData);
  }
  if (questionType === "ordering") {
    return parseOrderingFormData(formData);
  }
  if (questionType === "word_fill" || !questionType) {
    return parseWordFillFormData(formData, ["explanation", "explanation1", "explanation2"]);
  }
  throw new Response(`Unknown questionType: ${String(questionType)}`, { status: 400 });
}

async function handleUpdateQuestionIntent(
  client: BackendClient,
  workbookId: string,
  formData: FormData,
) {
  const questionId = requireQuestionId(formData);
  const { content, tags } = parseUpdateQuestionFormData(formData);
  const orderIndex = Number(formData.get("orderIndex") ?? 0);
  return saveQuestion(
    () =>
      updateQuestion(client, workbookId, questionId, {
        content,
        tags: optionalTags(tags),
        orderIndex,
      }),
    { ok: true },
  );
}

async function handleDeleteQuestionIntent(
  client: BackendClient,
  workbookId: string,
  formData: FormData,
) {
  const questionId = requireQuestionId(formData);
  await deleteQuestion(client, workbookId, questionId);
  return { ok: true };
}

async function handleClearStudyHistoryIntent(client: BackendClient, workbookId: string) {
  await deleteStudyHistory(client, workbookId);
  return { ok: true, cleared: true };
}

async function handleAddQuestionIntent(
  client: BackendClient,
  workbookId: string,
  questionType: QuestionType,
  { content, tags }: { content: string; tags: string[] },
) {
  return saveQuestion(
    () =>
      addQuestion(client, workbookId, {
        questionType,
        content,
        tags: optionalTags(tags),
        orderIndex: 0,
      }),
    { ok: true, added: true },
  );
}

export async function action({ request, params, context }: Route.ActionArgs) {
  const { accessToken } = requireAuth(context);
  const client = createBackendClient({ request, accessToken });
  const { workbookId } = params;
  const formData = await request.formData();
  const intent = formData.get("intent");

  if (intent === "updateTitle") {
    return handleUpdateTitleIntent(client, workbookId, formData);
  }
  if (intent === "updateQuestion") {
    return handleUpdateQuestionIntent(client, workbookId, formData);
  }
  if (intent === "deleteQuestion") {
    return handleDeleteQuestionIntent(client, workbookId, formData);
  }
  if (intent === "clearStudyHistory") {
    return handleClearStudyHistoryIntent(client, workbookId);
  }
  if (intent === "addWordFill") {
    return handleAddQuestionIntent(
      client,
      workbookId,
      "word_fill",
      parseWordFillFormData(formData, ["explanation"]),
    );
  }
  if (intent === "addMultipleChoice") {
    return handleAddQuestionIntent(
      client,
      workbookId,
      "multiple_choice",
      parseMultipleChoiceFormData(formData),
    );
  }
  if (intent === "addOrdering") {
    return handleAddQuestionIntent(client, workbookId, "ordering", parseOrderingFormData(formData));
  }
  throw new Response(`Unknown intent: ${String(intent)}`, { status: 400 });
}

export default function WorkbookDetail() {
  const { workbook, questions } = useLoaderData<typeof loader>();
  const { t } = useTranslation();
  const [editingQuestion, setEditingQuestion] = useState<Question | null>(null);
  const [editOpenCount, setEditOpenCount] = useState(0);
  const [editFetcherKey, setEditFetcherKey] = useState("edit-question-initial");
  const editFetcher = useFetcher<QuestionActionData>({ key: editFetcherKey });
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

  function openEditor(question: Question) {
    const openCount = editOpenCount + 1;
    setEditOpenCount(openCount);
    setEditFetcherKey(`edit-question-${question.questionId}-${openCount}`);
    setEditingQuestion(question);
  }

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
      : editingQuestion?.questionType === "ordering"
        ? t("workbooks.addQuestion.titleOrdering")
        : t("workbooks.addQuestion.titleMultipleChoice");

  const isClearing = clearHistoryFetcher.state !== "idle";

  return (
    <div>
      <WorkbookHeader
        title={workbook.title}
        description={workbook.description}
        visibility={workbook.visibility}
        language={workbook.language}
      />

      <div className="mb-4 flex flex-wrap items-center gap-2">
        <Button variant="outline" size="sm" onClick={handleClearStudyHistory} disabled={isClearing}>
          <Trash2Icon data-icon="inline-start" className="size-3.5 text-destructive" />
          <span>{t("workbooks.detail.clearStudyHistory")}</span>
        </Button>
        {import.meta.env.DEV && (
          <Button
            variant="ghost"
            size="sm"
            nativeButton={false}
            render={<Link to={`/workbooks/${workbook.workbookId}/debug-study`} />}
          >
            <BugIcon data-icon="inline-start" className="size-3.5" />
            <span>{t("workbooks.debugStudy.title")}</span>
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
            <QuestionCard key={question.questionId} question={question} onEdit={openEditor} />
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
          {editingQuestion?.questionType === "ordering" && (
            <OrderingEditForm
              question={editingQuestion}
              parsed={parseOrderingContent(editingQuestion.content)}
              fetcher={editFetcher}
              onCancel={() => setEditingQuestion(null)}
            />
          )}
        </SheetContent>
      </Sheet>
    </div>
  );
}
