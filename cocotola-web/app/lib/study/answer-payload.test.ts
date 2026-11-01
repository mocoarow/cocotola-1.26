import { describe, expect, it } from "vitest";
import {
  MAX_CHOICE_ID_LENGTH,
  MAX_SELECTED_CHOICE_IDS_COUNT,
  parseAnswerPayload,
} from "./answer-payload";

describe("parseAnswerPayload", () => {
  it("should return multiple_choice payload when selectedChoiceIds is valid", () => {
    // given
    const raw = JSON.stringify({ type: "multiple_choice", selectedChoiceIds: ["c1", "c2"] });

    // when
    const result = parseAnswerPayload(raw);

    // then
    expect(result).toEqual({ type: "multiple_choice", selectedChoiceIds: ["c1", "c2"] });
  });

  it("should return word_fill payload when correct is boolean", () => {
    // given
    const raw = JSON.stringify({ type: "word_fill", correct: true });

    // when
    const result = parseAnswerPayload(raw);

    // then
    expect(result).toEqual({ type: "word_fill", correct: true });
  });

  it.each([
    ["not JSON", "{"],
    ["unknown type", JSON.stringify({ type: "ordering", order: ["a"] })],
    ["missing type", JSON.stringify({ correct: true })],
    ["word_fill without correct", JSON.stringify({ type: "word_fill" })],
    ["word_fill with string correct", JSON.stringify({ type: "word_fill", correct: "true" })],
    [
      "multiple_choice with non-string id",
      JSON.stringify({ type: "multiple_choice", selectedChoiceIds: [1] }),
    ],
    [
      "multiple_choice over count limit",
      JSON.stringify({
        type: "multiple_choice",
        selectedChoiceIds: Array.from(
          { length: MAX_SELECTED_CHOICE_IDS_COUNT + 1 },
          (_, i) => `c${i}`,
        ),
      }),
    ],
    [
      "multiple_choice over id length limit",
      JSON.stringify({
        type: "multiple_choice",
        selectedChoiceIds: ["x".repeat(MAX_CHOICE_ID_LENGTH + 1)],
      }),
    ],
  ])("should return null when payload is %s", (_label, raw) => {
    // given / when
    const result = parseAnswerPayload(raw);

    // then
    expect(result).toBeNull();
  });

  it("should accept payload at the count and length limits", () => {
    // given
    const raw = JSON.stringify({
      type: "multiple_choice",
      selectedChoiceIds: Array.from({ length: MAX_SELECTED_CHOICE_IDS_COUNT }, () =>
        "x".repeat(MAX_CHOICE_ID_LENGTH),
      ),
    });

    // when
    const result = parseAnswerPayload(raw);

    // then
    expect(result).not.toBeNull();
  });
});
