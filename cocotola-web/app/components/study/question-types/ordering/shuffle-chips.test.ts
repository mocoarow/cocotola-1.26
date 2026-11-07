import { afterEach, describe, expect, it, vi } from "vitest";
import type { OrderingSegment } from "./answer";
import { shuffleChips } from "./shuffle-chips";

const chip = (text: string): OrderingSegment => ({ type: "chip", text });
const fixed = (text: string): OrderingSegment => ({ type: "fixed", text });

// Math.random() just below 1 makes every Fisher-Yates step keep its element in place.
const KEEP_IN_PLACE = 0.999;

function mockRandomSequence(values: readonly number[]) {
  let call = 0;
  vi.spyOn(Math, "random").mockImplementation(() => values[Math.min(call++, values.length - 1)]);
}

describe("shuffleChips", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should return the segment indices of every chip", () => {
    // given
    vi.spyOn(Math, "random").mockReturnValue(0);
    const segments = [chip("I"), chip("am"), fixed("."), chip("you"), chip("are")];

    // when
    const order = shuffleChips(segments);

    // then
    expect([...order].sort((a, b) => a - b)).toEqual([0, 1, 3, 4]);
  });

  it("should reshuffle when the shuffled texts match the correct order", () => {
    // given
    mockRandomSequence([KEEP_IN_PLACE, KEEP_IN_PLACE, 0, 0]);
    const segments = [chip("I"), chip("am"), chip("here")];

    // when
    const order = shuffleChips(segments);

    // then
    expect(order).toEqual([1, 2, 0]);
  });

  it("should reshuffle when only chips with the same text trade places", () => {
    // given
    mockRandomSequence([0, KEEP_IN_PLACE, 0, 0]);
    const segments = [chip("a"), chip("b"), chip("a")];

    // when
    const order = shuffleChips(segments);

    // then
    expect(order.map((i) => segments[i].text)).not.toEqual(["a", "b", "a"]);
  });

  it("should not match the correct order even when every shuffle keeps it", () => {
    // given
    vi.spyOn(Math, "random").mockReturnValue(KEEP_IN_PLACE);
    const segments = [chip("I"), chip("am"), chip("here")];

    // when
    const order = shuffleChips(segments);

    // then
    expect(order.map((i) => segments[i].text)).not.toEqual(["I", "am", "here"]);
  });

  it("should accept the correct order when every chip has the same text", () => {
    // given
    vi.spyOn(Math, "random").mockReturnValue(KEEP_IN_PLACE);
    const segments = [chip("ha"), chip("ha")];

    // when
    const order = shuffleChips(segments);

    // then
    expect(order).toEqual([0, 1]);
  });
});
