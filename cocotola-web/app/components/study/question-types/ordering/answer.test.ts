import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { parseOrderingContent } from "~/components/workbook/schemas";
import {
  computeRegions,
  evaluateRegions,
  type OrderingSegment,
  slotCorrectness,
  toOrderingDraft,
} from "./answer";

type FixtureCase = {
  name: string;
  segments: OrderingSegment[];
  regions: string[][];
  answer: string[][];
  correct: boolean;
};

const fixturePath = resolve(
  import.meta.dirname,
  "../../../../../../cocotola-question/domain/question/testdata/ordering_cases.json",
);
const fixtureCases: FixtureCase[] = JSON.parse(readFileSync(fixturePath, "utf8"));

function contentJson(segments: unknown): string {
  return JSON.stringify({
    source: { text: "私は学生です。", lang: "ja" },
    target: { lang: "en", segments },
    explanation: "",
  });
}

const chip = (text: string): OrderingSegment => ({ type: "chip", text });
const fixed = (text: string): OrderingSegment => ({ type: "fixed", text });

describe("computeRegions", () => {
  it.each(fixtureCases)("should return the shared fixture regions: $name", (c) => {
    // given / when
    const regions = computeRegions(c.segments);

    // then
    expect(regions).toEqual(c.regions);
  });
});

describe("evaluateRegions", () => {
  it.each(fixtureCases)("should match the shared fixture judgement: $name", (c) => {
    // given / when
    const correct = evaluateRegions(c.segments, c.answer);

    // then
    expect(correct).toBe(c.correct);
  });
});

describe("parseOrderingContent", () => {
  it.each(fixtureCases)("should accept content built from the fixture segments: $name", (c) => {
    // given
    const raw = contentJson(c.segments);

    // when
    const content = parseOrderingContent(raw);

    // then
    expect(content).not.toBeNull();
  });

  it.each([
    ["not JSON", "{"],
    ["missing target", JSON.stringify({ source: { text: "x", lang: "ja" } })],
    ["unknown segment type", contentJson([chip("I"), { type: "blank", text: "am" }])],
    ["non-string segment text", contentJson([chip("I"), { type: "chip", text: 1 }])],
    ["a single chip", contentJson([chip("I"), fixed("am")])],
    ["no segments", contentJson([])],
  ])("should return null when content has %s", (_label, raw) => {
    // given / when
    const content = parseOrderingContent(raw);

    // then
    expect(content).toBeNull();
  });
});

describe("toOrderingDraft", () => {
  const segments = [chip("I"), chip("am"), fixed("."), chip("you"), chip("are")];

  it("should return null when a slot is empty", () => {
    // given
    const slots = [0, 1, 3, null];

    // when
    const draft = toOrderingDraft(segments, slots);

    // then
    expect(draft).toBeNull();
  });

  it("should send the placed chip texts by region when every slot is filled", () => {
    // given
    const slots = [1, 0, 3, 4];

    // when
    const draft = toOrderingDraft(segments, slots);

    // then
    expect(draft?.payload).toEqual({
      type: "ordering",
      regions: [
        ["am", "I"],
        ["you", "are"],
      ],
    });
  });

  it("should be correct when every chip text is in its correct slot", () => {
    // given
    const slots = [0, 1, 3, 4];

    // when
    const draft = toOrderingDraft(segments, slots);

    // then
    expect(draft?.correct).toBe(true);
  });

  it("should be incorrect when chips are swapped across regions", () => {
    // given
    const slots = [0, 4, 3, 1];

    // when
    const draft = toOrderingDraft(segments, slots);

    // then
    expect(draft?.correct).toBe(false);
  });

  it("should be correct when chips with the same text are swapped", () => {
    // given
    const duplicated = [chip("the"), chip("cat"), chip("ate"), chip("the"), chip("fish")];
    const slots = [3, 1, 2, 0, 4];

    // when
    const draft = toOrderingDraft(duplicated, slots);

    // then
    expect(draft?.correct).toBe(true);
  });
});

describe("slotCorrectness", () => {
  it("should mark each slot by whether its chip text matches the correct chip text", () => {
    // given
    const segments = [chip("I"), chip("am"), fixed("."), chip("you"), chip("are")];
    const slots = [0, 4, 3, 1];

    // when
    const result = slotCorrectness(segments, slots);

    // then
    expect(result).toEqual([true, false, true, false]);
  });

  it("should mark every slot correct when chips with the same text are swapped", () => {
    // given
    const segments = [chip("the"), chip("cat"), chip("the")];
    const slots = [2, 1, 0];

    // when
    const result = slotCorrectness(segments, slots);

    // then
    expect(result).toEqual([true, true, true]);
  });
});
