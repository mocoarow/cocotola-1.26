import type { QuestionModule } from "~/components/study/question-shell/types";
import { parseOrderingContent } from "~/components/workbook/schemas";
import { chipSegmentIndices, type OrderingContent, toOrderingDraft } from "./answer";
import { OrderingBody } from "./body";
import { emptySlots } from "./placement";

export const orderingModule: QuestionModule<"ordering", OrderingContent> = {
  parse: parseOrderingContent,
  initialDraft: ({ target: { segments } }) =>
    toOrderingDraft(segments, emptySlots(chipSegmentIndices(segments).length)),
  revealLabelKey: "workbooks.study.check",
  explanation: (content, phase) =>
    phase === "revealed" && content.explanation ? content.explanation : undefined,
  Body: OrderingBody,
};
