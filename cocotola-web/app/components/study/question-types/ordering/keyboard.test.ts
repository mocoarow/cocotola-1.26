import { describe, expect, it } from "vitest";
import { type Areas, chipAreas, entryItem, keyAction } from "./keyboard";

type Modifiers = { shiftKey?: boolean; ctrlKey?: boolean; altKey?: boolean; metaKey?: boolean };

const press = (key: string, modifiers: Modifiers = {}) => ({
  key,
  shiftKey: false,
  ctrlKey: false,
  altKey: false,
  metaKey: false,
  ...modifiers,
});

// pool: chips 4, 5, 6 / slots: [1, null, 2, 3]
const AREAS: Areas = chipAreas([4, 5, 6], [1, null, 2, 3]);

describe("chipAreas", () => {
  it("should number pool chips by their order", () => {
    // given / when
    const areas = chipAreas([4, 5], [null]);

    // then
    expect(areas.pool).toEqual([
      { chip: 4, position: 0 },
      { chip: 5, position: 1 },
    ]);
  });

  it("should list only filled slots, numbered by slot", () => {
    // given / when
    const areas = chipAreas([], [1, null, 2]);

    // then
    expect(areas.slot).toEqual([
      { chip: 1, position: 0 },
      { chip: 2, position: 2 },
    ]);
  });
});

describe("entryItem", () => {
  it("should pick the first chip when the area was never focused", () => {
    // given / when
    const item = entryItem(AREAS.slot, undefined);

    // then
    expect(item).toEqual({ chip: 1, position: 0 });
  });

  it("should pick the remembered chip when it is still in the area", () => {
    // given / when
    const item = entryItem(AREAS.slot, { chip: 2, position: 2 });

    // then
    expect(item).toEqual({ chip: 2, position: 2 });
  });

  it("should pick the chip with the nearest number when the remembered chip is gone", () => {
    // given / when
    const item = entryItem(AREAS.slot, { chip: 9, position: 3 });

    // then
    expect(item).toEqual({ chip: 3, position: 3 });
  });

  it("should pick the earlier chip when two chips are equally near", () => {
    // given / when
    const item = entryItem(AREAS.slot, { chip: 9, position: 1 });

    // then
    expect(item).toEqual({ chip: 1, position: 0 });
  });

  it("should pick nothing when the area is empty", () => {
    // given / when
    const item = entryItem([], { chip: 9, position: 0 });

    // then
    expect(item).toBeUndefined();
  });
});

describe("keyAction within an area", () => {
  const tests = [
    { name: "ArrowRight to the next chip", key: "ArrowRight", from: 5, to: 6 },
    { name: "ArrowLeft to the previous chip", key: "ArrowLeft", from: 5, to: 4 },
    { name: "Home to the first chip", key: "Home", from: 6, to: 4 },
    { name: "End to the last chip", key: "End", from: 4, to: 6 },
  ];
  for (const tt of tests) {
    it(`should move with ${tt.name}`, () => {
      // given / when
      const action = keyAction(press(tt.key), { area: "pool", chip: tt.from }, AREAS, {});

      // then
      expect(action).toEqual({ kind: "move", to: { area: "pool", chip: tt.to } });
    });
  }

  it("should move over empty slots in the answer area", () => {
    // given / when
    const action = keyAction(press("ArrowRight"), { area: "slot", chip: 1 }, AREAS, {});

    // then
    expect(action).toEqual({ kind: "move", to: { area: "slot", chip: 2 } });
  });

  const shiftTests = [
    { key: "ArrowRight", to: 6 },
    { key: "ArrowLeft", to: 4 },
  ];
  for (const tt of shiftTests) {
    it(`should move like a plain arrow with Shift + ${tt.key} in the pool`, () => {
      // given / when
      const action = keyAction(
        press(tt.key, { shiftKey: true }),
        { area: "pool", chip: 5 },
        AREAS,
        {},
      );

      // then
      expect(action).toEqual({ kind: "move", to: { area: "pool", chip: tt.to } });
    });
  }
});

describe("keyAction at an edge", () => {
  const tests = [
    { key: "ArrowRight", from: 6 },
    { key: "ArrowLeft", from: 4 },
    { key: "Home", from: 4 },
    { key: "End", from: 6 },
  ];
  for (const tt of tests) {
    it(`should stay with ${tt.key} on chip ${tt.from}`, () => {
      // given / when
      const action = keyAction(press(tt.key), { area: "pool", chip: tt.from }, AREAS, {});

      // then
      expect(action).toEqual({ kind: "stay" });
    });
  }
});

describe("keyAction between areas", () => {
  it("should move up from the pool to the first slot chip when the answer area is new", () => {
    // given / when
    const action = keyAction(press("ArrowUp"), { area: "pool", chip: 5 }, AREAS, {});

    // then
    expect(action).toEqual({ kind: "move", to: { area: "slot", chip: 1 } });
  });

  it("should move up to the remembered slot chip", () => {
    // given
    const memory = { slot: { chip: 3, position: 3 } };

    // when
    const action = keyAction(press("ArrowUp"), { area: "pool", chip: 5 }, AREAS, memory);

    // then
    expect(action).toEqual({ kind: "move", to: { area: "slot", chip: 3 } });
  });

  it("should move down to the nearest pool chip when the remembered one is gone", () => {
    // given
    const memory = { pool: { chip: 9, position: 2 } };

    // when
    const action = keyAction(press("ArrowDown"), { area: "slot", chip: 1 }, AREAS, memory);

    // then
    expect(action).toEqual({ kind: "move", to: { area: "pool", chip: 6 } });
  });

  it("should stay with ArrowUp when the answer area is empty", () => {
    // given
    const areas = chipAreas([4, 5], [null, null]);

    // when
    const action = keyAction(press("ArrowUp"), { area: "pool", chip: 4 }, areas, {});

    // then
    expect(action).toEqual({ kind: "stay" });
  });

  it("should stay with ArrowDown when the pool is empty", () => {
    // given
    const areas = chipAreas([], [1, 2]);

    // when
    const action = keyAction(press("ArrowDown"), { area: "slot", chip: 1 }, areas, {});

    // then
    expect(action).toEqual({ kind: "stay" });
  });

  it("should stay with ArrowUp in the answer area", () => {
    // given / when
    const action = keyAction(press("ArrowUp"), { area: "slot", chip: 1 }, AREAS, {});

    // then
    expect(action).toEqual({ kind: "stay" });
  });

  it("should stay with ArrowDown in the pool", () => {
    // given / when
    const action = keyAction(press("ArrowDown"), { area: "pool", chip: 5 }, AREAS, {});

    // then
    expect(action).toEqual({ kind: "stay" });
  });
});

describe("keyAction for other keys", () => {
  const tests = [{ key: "Enter" }, { key: " " }, { key: "Tab" }, { key: "a" }];
  for (const tt of tests) {
    it(`should ignore ${JSON.stringify(tt.key)}`, () => {
      // given / when
      const action = keyAction(press(tt.key), { area: "pool", chip: 5 }, AREAS, {});

      // then
      expect(action).toEqual({ kind: "ignore" });
    });
  }

  const modifierTests = [
    { name: "Alt+ArrowLeft", key: "ArrowLeft", modifiers: { altKey: true } },
    { name: "Meta+ArrowLeft", key: "ArrowLeft", modifiers: { metaKey: true } },
    { name: "Ctrl+Home", key: "Home", modifiers: { ctrlKey: true } },
  ];
  for (const tt of modifierTests) {
    it(`should ignore ${tt.name}`, () => {
      // given / when
      const action = keyAction(press(tt.key, tt.modifiers), { area: "pool", chip: 5 }, AREAS, {});

      // then
      expect(action).toEqual({ kind: "ignore" });
    });
  }
});
