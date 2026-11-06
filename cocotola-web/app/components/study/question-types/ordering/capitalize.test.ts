import { describe, expect, it } from "vitest";
import type { OrderingSegment } from "./answer";
import { slotDisplayTexts } from "./capitalize";

const chip = (text: string): OrderingSegment => ({ type: "chip", text });
const fixed = (text: string): OrderingSegment => ({ type: "fixed", text });

describe("slotDisplayTexts", () => {
  it("should capitalize the chip in the first slot when the segments start with a chip", () => {
    // given
    const segments = [chip("you"), chip("are")];

    // when
    const texts = slotDisplayTexts(segments, [0, 1], "en");

    // then
    expect(texts[0]).toBe("You");
  });

  it("should not capitalize a chip in the middle of a sentence", () => {
    // given
    const segments = [chip("you"), chip("are")];

    // when
    const texts = slotDisplayTexts(segments, [0, 1], "en");

    // then
    expect(texts[1]).toBe("are");
  });

  it("should not capitalize the first slot when a fixed text precedes it", () => {
    // given
    const segments = [fixed("and"), chip("you"), chip("are")];

    // when
    const texts = slotDisplayTexts(segments, [1, 2], "en");

    // then
    expect(texts[0]).toBe("you");
  });

  it.each([
    ["period", "."],
    ["question mark", "?"],
    ["exclamation mark", "!"],
    ["period and closing quote", '."'],
    ["period and closing curly quote", ".”"],
    ["period and closing parenthesis", ".)"],
  ])("should capitalize a chip after a fixed text ending with a %s", (_label, text) => {
    // given
    const segments = [chip("I"), chip("am"), fixed(text), chip("you"), chip("are")];

    // when
    const texts = slotDisplayTexts(segments, [0, 1, 3, 4], "en");

    // then
    expect(texts[2]).toBe("You");
  });

  it("should not capitalize a chip after a fixed text ending with a comma", () => {
    // given
    const segments = [chip("yes"), fixed(","), chip("please"), chip("do")];

    // when
    const texts = slotDisplayTexts(segments, [0, 2, 3], "en");

    // then
    expect(texts[1]).toBe("please");
  });

  it("should capitalize a chip after a placed chip ending with a question mark", () => {
    // given
    const segments = [chip("ready?"), chip("go")];

    // when
    const texts = slotDisplayTexts(segments, [0, 1], "en");

    // then
    expect(texts[1]).toBe("Go");
  });

  it("should not capitalize a chip after an empty slot", () => {
    // given
    const segments = [chip("ready?"), chip("go")];

    // when
    const texts = slotDisplayTexts(segments, [null, 1], "en");

    // then
    expect(texts[1]).toBe("go");
  });

  it("should return null for an empty slot", () => {
    // given
    const segments = [chip("you"), chip("are")];

    // when
    const texts = slotDisplayTexts(segments, [0, null], "en");

    // then
    expect(texts[1]).toBeNull();
  });

  it("should capitalize with the rules of the target language", () => {
    // given
    const segments = [chip("istanbul"), chip("güzel")];

    // when
    const texts = slotDisplayTexts(segments, [0, 1], "tr");

    // then
    expect(texts[0]).toBe("İstanbul");
  });

  it("should fall back to the default upper case when the language tag is invalid", () => {
    // given
    const segments = [chip("you"), chip("are")];

    // when
    const texts = slotDisplayTexts(segments, [0, 1], "not a lang!");

    // then
    expect(texts[0]).toBe("You");
  });

  it("should capitalize a first character outside the basic multilingual plane", () => {
    // given
    const segments = [chip("𐐨𐐯"), chip("de")];

    // when
    const texts = slotDisplayTexts(segments, [0, 1], "en");

    // then
    expect(texts[0]).toBe("𐐀𐐯");
  });
});
