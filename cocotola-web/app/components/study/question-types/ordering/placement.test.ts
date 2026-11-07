import { describe, expect, it } from "vitest";
import { emptySlots, placeChip, placeInFirstEmpty, returnChip } from "./placement";

describe("emptySlots", () => {
  it("should create the given number of empty slots", () => {
    // given / when
    const slots = emptySlots(3);

    // then
    expect(slots).toEqual([null, null, null]);
  });
});

describe("placeInFirstEmpty", () => {
  it("should put a pool chip into the first empty slot", () => {
    // given
    const slots = [5, null, null];

    // when
    const next = placeInFirstEmpty(slots, 7);

    // then
    expect(next).toEqual([5, 7, null]);
  });

  it("should fill an empty slot left before filled ones", () => {
    // given
    const slots = [null, 5, null];

    // when
    const next = placeInFirstEmpty(slots, 7);

    // then
    expect(next).toEqual([7, 5, null]);
  });

  it("should keep the slots when every slot is filled", () => {
    // given
    const slots = [5, 7];

    // when
    const next = placeInFirstEmpty(slots, 9);

    // then
    expect(next).toBe(slots);
  });

  it("should keep the slots when the chip is already placed", () => {
    // given
    const slots = [5, null];

    // when
    const next = placeInFirstEmpty(slots, 5);

    // then
    expect(next).toBe(slots);
  });

  it("should not mutate the given slots", () => {
    // given
    const slots = [null, null];

    // when
    placeInFirstEmpty(slots, 5);

    // then
    expect(slots).toEqual([null, null]);
  });
});

describe("placeChip", () => {
  it("should put a pool chip into an empty slot", () => {
    // given
    const slots = [null, null, null];

    // when
    const next = placeChip(slots, 7, 2);

    // then
    expect(next).toEqual([null, null, 7]);
  });

  it("should return the occupant to the pool when a pool chip is put into a filled slot", () => {
    // given
    const slots = [5, null];

    // when
    const next = placeChip(slots, 7, 0);

    // then
    expect(next).toEqual([7, null]);
  });

  it("should move a placed chip to an empty slot in another region", () => {
    // given
    const slots = [5, null, null];

    // when
    const next = placeChip(slots, 5, 2);

    // then
    expect(next).toEqual([null, null, 5]);
  });

  it("should swap chips when a placed chip is moved onto a filled slot", () => {
    // given
    const slots = [5, 7, 9];

    // when
    const next = placeChip(slots, 5, 2);

    // then
    expect(next).toEqual([9, 7, 5]);
  });

  it("should keep the slots when a chip is dropped onto its own slot", () => {
    // given
    const slots = [5, 7];

    // when
    const next = placeChip(slots, 5, 0);

    // then
    expect(next).toBe(slots);
  });
});

describe("returnChip", () => {
  it("should empty the slot so the chip goes back to the pool", () => {
    // given
    const slots = [5, 7];

    // when
    const next = returnChip(slots, 1);

    // then
    expect(next).toEqual([5, null]);
  });
});
