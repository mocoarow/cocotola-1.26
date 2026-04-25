import { useTranslation } from "react-i18next";
import { useLoaderData } from "react-router";
import { AddQuestionSection } from "~/components/workbook/add-question-section";
import { QuestionCard } from "~/components/workbook/question-card";
import { parseMultipleChoiceFormData } from "~/components/workbook/schemas";
import { WorkbookHeader } from "~/components/workbook/workbook-header";
import {
  addQuestion,
  deleteQuestion,
  listQuestions,
  updateQuestion,
} from "~/lib/api/question.server";
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

  return { ok: true };
}

export default function WorkbookDetail() {
  const { workbook, questions } = useLoaderData<typeof loader>();
  const { t } = useTranslation();

  return (
    <div>
      <WorkbookHeader
        title={workbook.title}
        description={workbook.description}
        visibility={workbook.visibility}
      />

      <div className="mb-6">
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
            <QuestionCard key={question.questionId} question={question} />
          ))}
        </div>
      )}
    </div>
  );
}
