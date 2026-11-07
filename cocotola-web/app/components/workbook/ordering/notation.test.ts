import { describe, expect, it } from "vitest";
import type { OrderingSegment } from "~/components/study/question-types/ordering/answer";
import { MAX_ORDERING_SEGMENTS, MAX_ORDERING_TEXT_BYTES } from "~/lib/study/answer-payload";
import { formatOrderingNotation, parseOrderingNotation, trimSpace } from "./notation";

const chip = (text: string): OrderingSegment => ({ type: "chip", text });
const fixed = (text: string): OrderingSegment => ({ type: "fixed", text });

const EXAMPLE_NOTATION = "I / am / a student [[.]] you / are / a teacher [[.]]";
const EXAMPLE_SEGMENTS = [
  chip("I"),
  chip("am"),
  chip("a student"),
  fixed("."),
  chip("you"),
  chip("are"),
  chip("a teacher"),
  fixed("."),
];

function chipsNotation(count: number): string {
  return Array.from({ length: count }, (_, i) => `w${i}`).join(" / ");
}

describe("parseOrderingNotation", () => {
  it("should return segments when the notation is valid", () => {
    // given
    const notation = EXAMPLE_NOTATION;

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: true, segments: EXAMPLE_SEGMENTS });
  });

  it("should return a leading fixed segment when the notation starts with [[ ]]", () => {
    // given
    const notation = "[[Q:]] who / are / you";

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({
      ok: true,
      segments: [fixed("Q:"), chip("who"), chip("are"), chip("you")],
    });
  });

  it("should keep a slash inside a fixed segment", () => {
    // given
    const notation = "yes / no [[and/or]] maybe";

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({
      ok: true,
      segments: [chip("yes"), chip("no"), fixed("and/or"), chip("maybe")],
    });
  });

  it("should trim spaces around chips and fixed texts", () => {
    // given
    const notation = "  I  /  am  [[  .  ]]  ";

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: true, segments: [chip("I"), chip("am"), fixed(".")] });
  });

  it("should trim U+0085 around chips", () => {
    // given
    const notation = "\u0085I\u0085 / \u0085am\u0085";

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: true, segments: [chip("I"), chip("am")] });
  });

  it("should accept the maximum number of segments", () => {
    // given
    const notation = chipsNotation(MAX_ORDERING_SEGMENTS);

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result.ok).toBe(true);
  });

  it("should accept a chip of the maximum byte length", () => {
    // given
    const notation = `${"a".repeat(MAX_ORDERING_TEXT_BYTES)} / b`;

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result.ok).toBe(true);
  });

  it.each([
    { name: "a [[ without ]]", notation: "I / am [[." },
    { name: "a [[ followed by another [[ without ]]", notation: "I [[.]] am [[!" },
  ])("should return unmatchedOpen when the notation has $name", ({ notation }) => {
    // given / when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: false, error: "unmatchedOpen" });
  });

  it.each([
    { name: "a ]] without [[", notation: "I / am .]]" },
    { name: "a ]] after a closed fixed", notation: "I [[.]] am ]] you" },
  ])("should return unmatchedClose when the notation has $name", ({ notation }) => {
    // given / when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: false, error: "unmatchedClose" });
  });

  it.each([
    { name: "two slashes in a row", notation: "I / / am" },
    { name: "a trailing slash", notation: "I / am /" },
    { name: "a leading slash", notation: "/ I / am" },
    { name: "a slash right before a fixed", notation: "I / am / [[.]] you" },
  ])("should return emptyChip when the notation has $name", ({ notation }) => {
    // given / when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: false, error: "emptyChip" });
  });

  it("should return emptyFixed when a fixed text is blank", () => {
    // given
    const notation = "I / am [[ ]]";

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: false, error: "emptyFixed" });
  });

  it.each([
    { name: "one chip", notation: "I [[.]]" },
    { name: "no chip", notation: "[[.]]" },
    { name: "nothing", notation: "   " },
  ])("should return tooFewChips when the notation has $name", ({ notation }) => {
    // given / when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: false, error: "tooFewChips" });
  });

  it("should return tooManySegments when the segments exceed the maximum", () => {
    // given
    const notation = chipsNotation(MAX_ORDERING_SEGMENTS + 1);

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: false, error: "tooManySegments" });
  });

  it("should return tooManySegments without throwing when the notation has many fixed texts", () => {
    // given
    const notation = "[[a]]".repeat(10000);

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: false, error: "tooManySegments" });
  });

  it.each([
    { name: "ASCII", text: "a".repeat(MAX_ORDERING_TEXT_BYTES + 1) },
    { name: "multibyte", text: "あ".repeat(Math.floor(MAX_ORDERING_TEXT_BYTES / 3) + 1) },
  ])("should return textTooLong when a $name chip exceeds the maximum bytes", ({ text }) => {
    // given
    const notation = `${text} / b`;

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: false, error: "textTooLong" });
  });

  it.each([
    { name: "a chip with [", notation: "I / a[m" },
    { name: "a chip with ]", notation: "I / a]m" },
    { name: "a fixed with [", notation: "I / am [[a[b]]" },
  ])("should return forbiddenChar when the notation has $name", ({ notation }) => {
    // given / when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: false, error: "forbiddenChar" });
  });

  it("should return htmlTag when a chip contains an HTML tag", () => {
    // given
    const notation = "I / <b>am</b>";

    // when
    const result = parseOrderingNotation(notation);

    // then
    expect(result).toEqual({ ok: false, error: "htmlTag" });
  });
});

describe("formatOrderingNotation", () => {
  it("should join chips with slashes and wrap fixed texts in [[ ]]", () => {
    // given
    const segments = EXAMPLE_SEGMENTS;

    // when
    const notation = formatOrderingNotation(segments);

    // then
    expect(notation).toBe(EXAMPLE_NOTATION);
  });

  it("should return the same segments when the formatted notation is parsed", () => {
    // given
    const segments = [fixed("Q:"), chip("who"), fixed("and/or"), fixed("!"), chip("you")];

    // when
    const result = parseOrderingNotation(formatOrderingNotation(segments));

    // then
    expect(result).toEqual({ ok: true, segments });
  });
});

describe("trimSpace", () => {
  it.each([
    { name: "U+0085", text: "\u0085a\u0085" },
    { name: "ideographic space", text: "　a　" },
    { name: "tab and newline", text: "\ta\n" },
  ])("should remove $name at both ends", ({ text }) => {
    // given / when
    const trimmed = trimSpace(text);

    // then
    expect(trimmed).toBe("a");
  });
});
