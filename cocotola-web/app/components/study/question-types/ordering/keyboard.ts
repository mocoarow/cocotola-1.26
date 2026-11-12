import type { Slots } from "./answer";

export type ChipArea = "pool" | "slot";

// position: the index in the pool, or the slot number in the answer area.
export type AreaItem = { chip: number; position: number };

export type Areas = Readonly<Record<ChipArea, readonly AreaItem[]>>;

export type FocusMemory = Readonly<Partial<Record<ChipArea, AreaItem>>>;

export type ChipFocus = { area: ChipArea; chip: number };

export type KeyInput = {
  key: string;
  shiftKey: boolean;
  ctrlKey: boolean;
  altKey: boolean;
  metaKey: boolean;
};

export type KeyAction = { kind: "ignore" } | { kind: "stay" } | { kind: "move"; to: ChipFocus };

const AREA_SWITCHES: ReadonlyMap<string, { from: ChipArea; to: ChipArea }> = new Map([
  ["ArrowUp", { from: "pool", to: "slot" }],
  ["ArrowDown", { from: "slot", to: "pool" }],
]);

export function chipAreas(poolChips: readonly number[], slots: Slots): Areas {
  return {
    pool: poolChips.map((chip, position) => ({ chip, position })),
    slot: slots.flatMap((chip, position) => (chip === null ? [] : [{ chip, position }])),
  };
}

export function entryItem(
  items: readonly AreaItem[],
  remembered: AreaItem | undefined,
): AreaItem | undefined {
  if (remembered === undefined) return items[0];
  const same = items.find((item) => item.chip === remembered.chip);
  if (same !== undefined) return same;
  const distance = (item: AreaItem) => Math.abs(item.position - remembered.position);
  return items.reduce<AreaItem | undefined>(
    (nearest, item) =>
      nearest === undefined || distance(item) < distance(nearest) ? item : nearest,
    undefined,
  );
}

function indexWithin(key: string, index: number, last: number): number | null {
  switch (key) {
    case "ArrowLeft":
      return Math.max(index - 1, 0);
    case "ArrowRight":
      return Math.min(index + 1, last);
    case "Home":
      return 0;
    case "End":
      return last;
    default:
      return null;
  }
}

function moveTo(area: ChipArea, item: AreaItem | undefined, from: ChipFocus): KeyAction {
  if (item === undefined || (area === from.area && item.chip === from.chip))
    return { kind: "stay" };
  return { kind: "move", to: { area, chip: item.chip } };
}

export function keyAction(
  input: KeyInput,
  from: ChipFocus,
  areas: Areas,
  memory: FocusMemory,
): KeyAction {
  if (input.ctrlKey || input.altKey || input.metaKey) return { kind: "ignore" };
  const across = AREA_SWITCHES.get(input.key);
  if (across !== undefined) {
    if (across.from !== from.area) return { kind: "stay" };
    return moveTo(across.to, entryItem(areas[across.to], memory[across.to]), from);
  }
  const items = areas[from.area];
  const index = items.findIndex((item) => item.chip === from.chip);
  const next = indexWithin(input.key, index, items.length - 1);
  if (next === null) return { kind: "ignore" };
  return moveTo(from.area, items[next], from);
}
