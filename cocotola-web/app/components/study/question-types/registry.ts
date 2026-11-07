import type { QuestionModule } from "~/components/study/question-shell/types";
import type { QuestionType } from "~/components/workbook/schemas";
import type { MultipleChoiceContent } from "./multiple-choice/answer";
import { multipleChoiceModule } from "./multiple-choice/module";
import type { OrderingContent } from "./ordering/answer";
import { orderingModule } from "./ordering/module";
import type { WordFillContent } from "./word-fill/answer";
import { wordFillModule } from "./word-fill/module";

type ContentMap = {
  multiple_choice: MultipleChoiceContent;
  word_fill: WordFillContent;
  ordering: OrderingContent;
};

const registry: { [T in QuestionType]: QuestionModule<T, ContentMap[T]> } = {
  multiple_choice: multipleChoiceModule,
  word_fill: wordFillModule,
  ordering: orderingModule,
};

export function isQuestionType(value: string): value is QuestionType {
  return Object.hasOwn(registry, value);
}

export function getQuestionModule<T extends QuestionType>(
  type: T,
): QuestionModule<T, ContentMap[T]> {
  return registry[type];
}
