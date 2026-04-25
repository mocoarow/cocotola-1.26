import { describe, expect, test } from "vitest";
import {
  parseMultipleChoiceContent,
  parseMultipleChoiceFormData,
  parseWordFillContent,
} from "./schemas";

function buildFormData(entries: Record<string, string>): FormData {
  const fd = new FormData();
  for (const [key, value] of Object.entries(entries)) {
    fd.set(key, value);
  }
  return fd;
}

describe("parseWordFillContent", () => {
  test("should parse valid word fill content", () => {
    const content = JSON.stringify({
      source: { text: "ゴミを捨てる", lang: "ja" },
      target: { text: "throw it away", lang: "en" },
      explanation: "句動詞",
    });
    const result = parseWordFillContent(content);
    expect(result).toEqual({
      source: { text: "ゴミを捨てる", lang: "ja" },
      target: { text: "throw it away", lang: "en" },
      explanation: "句動詞",
    });
  });

  test("should return null for invalid JSON", () => {
    expect(parseWordFillContent("not json")).toBeNull();
  });

  test("should return null for non-object content", () => {
    expect(parseWordFillContent('"string"')).toBeNull();
  });
});

describe("parseMultipleChoiceContent", () => {
  test("should parse valid multiple choice content", () => {
    const content = JSON.stringify({
      questionText: "What is 1+1?",
      choices: [
        { id: "1", text: "2", isCorrect: true },
        { id: "2", text: "3", isCorrect: false },
      ],
      explanation: "Basic math",
      shuffleChoices: true,
    });
    const result = parseMultipleChoiceContent(content);
    expect(result).toEqual({
      questionText: "What is 1+1?",
      choices: [
        { id: "1", text: "2", isCorrect: true },
        { id: "2", text: "3", isCorrect: false },
      ],
      explanation: "Basic math",
      shuffleChoices: true,
    });
  });

  test("should return null for invalid JSON", () => {
    expect(parseMultipleChoiceContent("bad")).toBeNull();
  });
});

describe("parseMultipleChoiceFormData", () => {
  test("should parse valid form data", () => {
    const choices = [
      { id: "1", text: "Answer A", isCorrect: true },
      { id: "2", text: "Answer B", isCorrect: false },
    ];
    const fd = buildFormData({
      questionText: "What is the capital of France?",
      choices: JSON.stringify(choices),
      explanation: "Paris is the capital",
      shuffleChoices: "true",
      tags: "level:beginner,topic:geography",
    });

    const result = parseMultipleChoiceFormData(fd);

    const parsed = JSON.parse(result.content);
    expect(parsed.questionText).toBe("What is the capital of France?");
    expect(parsed.choices).toHaveLength(2);
    expect(parsed.shuffleChoices).toBe(true);
    expect(parsed.explanation).toBe("Paris is the capital");
    expect(result.tags).toEqual(["level:beginner", "topic:geography"]);
  });

  test("should throw 400 when questionText is missing", () => {
    const fd = buildFormData({
      choices: JSON.stringify([{ id: "1", text: "A", isCorrect: true }]),
      shuffleChoices: "true",
    });

    expect(() => parseMultipleChoiceFormData(fd)).toThrow();
  });

  test("should throw 400 when choices is not valid JSON", () => {
    const fd = buildFormData({
      questionText: "Q?",
      choices: "not-json",
      shuffleChoices: "true",
    });

    expect(() => parseMultipleChoiceFormData(fd)).toThrow();
  });

  test("should throw 400 when no choice is marked correct", () => {
    const choices = [
      { id: "1", text: "A", isCorrect: false },
      { id: "2", text: "B", isCorrect: false },
    ];
    const fd = buildFormData({
      questionText: "Q?",
      choices: JSON.stringify(choices),
      shuffleChoices: "false",
    });

    expect(() => parseMultipleChoiceFormData(fd)).toThrow();
  });

  test("should throw 400 when choices array is empty", () => {
    const fd = buildFormData({
      questionText: "Q?",
      choices: JSON.stringify([]),
      shuffleChoices: "false",
    });

    expect(() => parseMultipleChoiceFormData(fd)).toThrow();
  });

  test("should handle empty tags", () => {
    const choices = [{ id: "1", text: "A", isCorrect: true }];
    const fd = buildFormData({
      questionText: "Q?",
      choices: JSON.stringify(choices),
      shuffleChoices: "true",
      tags: "",
    });

    const result = parseMultipleChoiceFormData(fd);
    expect(result.tags).toEqual([]);
  });

  test("should handle missing explanation", () => {
    const choices = [{ id: "1", text: "A", isCorrect: true }];
    const fd = buildFormData({
      questionText: "Q?",
      choices: JSON.stringify(choices),
      shuffleChoices: "true",
    });

    const result = parseMultipleChoiceFormData(fd);
    const parsed = JSON.parse(result.content);
    expect(parsed.explanation).toBe("");
  });
});
