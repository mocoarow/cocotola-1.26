import type { Slots } from "./answer";

export type ChipArea = "pool" | "slot";

// position: the index in the pool, or the slot number in the answer area.
export type AreaItem = { chip: number; position: number };

export type Areas = Readonly<Record<ChipArea, readonly AreaItem[]> & { slotCount: number }>;

export type FocusMemory = Readonly<Partial<Record<ChipArea, AreaItem>>>;

export type ChipFocus = { area: ChipArea; chip: number };

export type KeyInput = {
  key: string;
  shiftKey: boolean;
  ctrlKey: boolean;
  altKey: boolean;
  metaKey: boolean;
};

export type SlotEdit =
  | { kind: "return"; slot: number }
  | { kind: "place"; chip: number; slot: number };

export type KeyAction =
  | { kind: "ignore" }
  | { kind: "stay" }
  | { kind: "move"; to: ChipFocus }
  | { kind: "edit"; edit: SlotEdit };

const AREA_SWITCHES: ReadonlyMap<string, { from: ChipArea; to: ChipArea }> = new Map([
  ["ArrowUp", { from: "pool", to: "slot" }],
  ["ArrowDown", { from: "slot", to: "pool" }],
]);

const SLOT_STEPS: ReadonlyMap<string, number> = new Map([
  ["ArrowLeft", -1],
  ["ArrowRight", 1],
]);

export function chipAreas(poolChips: readonly number[], slots: Slots): Areas {
  return {
    pool: poolChips.map((chip, position) => ({ chip, position })),
    slot: slots.flatMap((chip, position) => (chip === null ? [] : [{ chip, position }])),
    slotCount: slots.length,
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

function returnLast(areas: Areas): KeyAction {
  const last = areas.slot.at(-1);
  if (last === undefined) return { kind: "stay" };
  return { kind: "edit", edit: { kind: "return", slot: last.position } };
}

function shiftInSlots(step: number, from: ChipFocus, areas: Areas): KeyAction {
  const item = areas.slot.find((i) => i.chip === from.chip);
  if (item === undefined) return { kind: "stay" };
  const to = item.position + step;
  if (to < 0 || to >= areas.slotCount) return { kind: "stay" };
  return { kind: "edit", edit: { kind: "place", chip: from.chip, slot: to } };
}

export function keyAction(
  input: KeyInput,
  from: ChipFocus,
  areas: Areas,
  memory: FocusMemory,
): KeyAction {
  if (input.ctrlKey || input.altKey || input.metaKey) return { kind: "ignore" };
  if (input.key === "Backspace") return returnLast(areas);
  const step = SLOT_STEPS.get(input.key);
  if (input.shiftKey && from.area === "slot" && step !== undefined)
    return shiftInSlots(step, from, areas);
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
