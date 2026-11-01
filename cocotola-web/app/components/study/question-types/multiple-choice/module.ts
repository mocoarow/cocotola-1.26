import type { QuestionModule } from "~/components/study/question-shell/types";
import { parseMultipleChoiceContent } from "~/components/workbook/schemas";
import type { MultipleChoiceContent } from "./answer";
import { MultipleChoiceBody } from "./body";

export const multipleChoiceModule: QuestionModule<"multiple_choice", MultipleChoiceContent> = {
  parse: parseMultipleChoiceContent,
  initialDraft: () => null,
  revealLabelKey: "workbooks.study.check",
  explanation: (content, phase) =>
    phase === "revealed" && content.explanation ? content.explanation : undefined,
  Body: MultipleChoiceBody,
};
