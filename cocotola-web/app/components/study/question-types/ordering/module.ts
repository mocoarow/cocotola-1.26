import type { QuestionModule } from "~/components/study/question-shell/types";
import { parseOrderingContent } from "~/components/workbook/schemas";
import type { OrderingContent } from "./answer";
import { OrderingBody } from "./body";

export const orderingModule: QuestionModule<"ordering", OrderingContent> = {
  parse: parseOrderingContent,
  initialDraft: () => null,
  revealLabelKey: "workbooks.study.check",
  explanation: (content, phase) =>
    phase === "revealed" && content.explanation ? content.explanation : undefined,
  Body: OrderingBody,
};
