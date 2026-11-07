import { describe, expect, it } from "vitest";
import { applyDrop, chipDragId, POOL_DROP_ID, slotDropId } from "./drop";

const CHIPS = [1, 2, 4, 5];

describe("applyDrop", () => {
  it("should put a pool chip into an empty slot", () => {
    // given
    const slots = [null, null, null, null];

    // when
    const next = applyDrop(slots, CHIPS, chipDragId(4), slotDropId(2));

    // then
    expect(next).toEqual([null, null, 4, null]);
  });

  it("should return the occupant to the pool when a pool chip is dropped onto a filled slot", () => {
    // given
    const slots = [1, null, null, null];

    // when
    const next = applyDrop(slots, CHIPS, chipDragId(4), slotDropId(0));

    // then
    expect(next).toEqual([4, null, null, null]);
  });

  it("should swap chips when a placed chip is dropped onto a filled slot", () => {
    // given
    const slots = [1, 2, 4, 5];

    // when
    const next = applyDrop(slots, CHIPS, chipDragId(1), slotDropId(3));

    // then
    expect(next).toEqual([5, 2, 4, 1]);
  });

  it("should move a placed chip to an empty slot in another region", () => {
    // given
    const slots = [1, null, null, null];

    // when
    const next = applyDrop(slots, CHIPS, chipDragId(1), slotDropId(3));

    // then
    expect(next).toEqual([null, null, null, 1]);
  });

  it("should return a placed chip to the pool when it is dropped onto the pool", () => {
    // given
    const slots = [1, 2, null, null];

    // when
    const next = applyDrop(slots, CHIPS, chipDragId(2), POOL_DROP_ID);

    // then
    expect(next).toEqual([1, null, null, null]);
  });

  it("should not mutate the given slots", () => {
    // given
    const slots = [1, 2, null, null];

    // when
    applyDrop(slots, CHIPS, chipDragId(1), slotDropId(1));

    // then
    expect(slots).toEqual([1, 2, null, null]);
  });
});

describe("applyDrop without a change", () => {
  it("should keep the slots when the chip is dropped outside any target", () => {
    // given
    const slots = [1, null, null, null];

    // when
    const next = applyDrop(slots, CHIPS, chipDragId(1), null);

    // then
    expect(next).toBe(slots);
  });

  it("should keep the slots when a placed chip is dropped onto its own slot", () => {
    // given
    const slots = [1, null, null, null];

    // when
    const next = applyDrop(slots, CHIPS, chipDragId(1), slotDropId(0));

    // then
    expect(next).toBe(slots);
  });

  it("should keep the slots when a pool chip is dropped onto the pool", () => {
    // given
    const slots = [1, null, null, null];

    // when
    const next = applyDrop(slots, CHIPS, chipDragId(2), POOL_DROP_ID);

    // then
    expect(next).toBe(slots);
  });

  it.each([
    { name: "a malformed drag id", activeId: "chip-x", overId: slotDropId(0) },
    { name: "a numeric drag id", activeId: 1, overId: slotDropId(0) },
    { name: "a chip that is not in the question", activeId: chipDragId(3), overId: slotDropId(0) },
    { name: "a malformed drop id", activeId: chipDragId(1), overId: "slot-" },
    { name: "a slot out of range", activeId: chipDragId(1), overId: slotDropId(4) },
    { name: "a drop id of another kind", activeId: chipDragId(1), overId: chipDragId(2) },
  ])("should keep the slots when given $name", ({ activeId, overId }) => {
    // given
    const slots = [null, null, null, null];

    // when
    const next = applyDrop(slots, CHIPS, activeId, overId);

    // then
    expect(next).toBe(slots);
  });
});
