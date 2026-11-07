import { describe, expect, it } from "vitest";
import { parseOrderingFormData } from "./form-data";

const VALID_ENTRIES = {
  sourceText: "私は学生です。",
  sourceLang: "ja",
  targetText: "I / am / a student [[.]]",
  targetLang: "en",
  explanation: "be 動詞",
  tags: "level:beginner, topic:be",
};

function buildFormData(entries: Record<string, string>): FormData {
  const fd = new FormData();
  for (const [key, value] of Object.entries(entries)) {
    fd.set(key, value);
  }
  return fd;
}

function thrownBy(fn: () => unknown): unknown {
  try {
    fn();
  } catch (thrown) {
    return thrown;
  }
  return undefined;
}

function parseContent(entries: Record<string, string>) {
  return JSON.parse(parseOrderingFormData(buildFormData(entries)).content);
}

describe("parseOrderingFormData", () => {
  it("should put the source text and language into the content", () => {
    // given
    const entries = VALID_ENTRIES;

    // when
    const content = parseContent(entries);

    // then
    expect(content.source).toEqual({ text: "私は学生です。", lang: "ja" });
  });

  it("should put the parsed segments and target language into the content", () => {
    // given
    const entries = VALID_ENTRIES;

    // when
    const content = parseContent(entries);

    // then
    expect(content.target).toEqual({
      lang: "en",
      segments: [
        { type: "chip", text: "I" },
        { type: "chip", text: "am" },
        { type: "chip", text: "a student" },
        { type: "fixed", text: "." },
      ],
    });
  });

  it("should put the explanation into the content", () => {
    // given
    const entries = VALID_ENTRIES;

    // when
    const content = parseContent(entries);

    // then
    expect(content.explanation).toBe("be 動詞");
  });

  it("should omit the explanation when it is blank", () => {
    // given
    const entries = { ...VALID_ENTRIES, explanation: "  " };

    // when
    const content = parseContent(entries);

    // then
    expect(content).not.toHaveProperty("explanation");
  });

  it("should split the tags by commas", () => {
    // given
    const fd = buildFormData(VALID_ENTRIES);

    // when
    const { tags } = parseOrderingFormData(fd);

    // then
    expect(tags).toEqual(["level:beginner", "topic:be"]);
  });

  it.each([
    "sourceText",
    "sourceLang",
    "targetText",
    "targetLang",
  ])("should throw 400 when %s is missing", (field) => {
    // given
    const fd = buildFormData(VALID_ENTRIES);
    fd.delete(field);

    // when
    const thrown = thrownBy(() => parseOrderingFormData(fd));

    // then
    expect(thrown).toMatchObject({ status: 400 });
  });

  it("should throw 400 when the notation is invalid", () => {
    // given
    const fd = buildFormData({ ...VALID_ENTRIES, targetText: "I / am .]]" });

    // when
    const thrown = thrownBy(() => parseOrderingFormData(fd));

    // then
    expect(thrown).toMatchObject({ status: 400 });
  });
});
