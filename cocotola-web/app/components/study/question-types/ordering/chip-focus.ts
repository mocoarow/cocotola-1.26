import { type KeyboardEvent, useEffect, useRef, useState } from "react";
import {
  type AreaItem,
  type Areas,
  type ChipArea,
  type ChipFocus,
  entryItem,
  type FocusMemory,
  keyAction,
} from "./keyboard";

export type ChipBinding = {
  ref: (el: HTMLButtonElement | null) => void;
  tabIndex: number;
  onFocus: () => void;
  onKeyDown: (event: KeyboardEvent<HTMLButtonElement>) => void;
};

export type BindChip = (chip: number) => ChipBinding;

const buttonKey = ({ area, chip }: ChipFocus) => `${area}-${chip}`;

const sameItem = (a: AreaItem | undefined, b: AreaItem) =>
  a?.chip === b.chip && a.position === b.position;

// Roving tabindex over the pool and the answer area, plus focus that follows a chip after it moves.
export function useChipFocus(areas: Areas) {
  const buttons = useRef(new Map<string, HTMLButtonElement>());
  const [target, setTarget] = useState<{ to: ChipFocus } | null>(null);
  const [memory, setMemory] = useState<FocusMemory>({});

  useEffect(() => {
    if (target !== null) buttons.current.get(buttonKey(target.to))?.focus();
  }, [target]);

  const focus = (area: ChipArea, chip: number) => setTarget({ to: { area, chip } });

  const remember = (area: ChipArea, chip: number) => {
    const item = areas[area].find((i) => i.chip === chip);
    if (item === undefined) return;
    setMemory((m) => (sameItem(m[area], item) ? m : { ...m, [area]: item }));
  };

  const handleKeyDown = (from: ChipFocus, event: KeyboardEvent<HTMLButtonElement>) => {
    const action = keyAction(event, from, areas, memory);
    if (action.kind === "ignore") return;
    event.preventDefault();
    if (action.kind === "move") focus(action.to.area, action.to.chip);
  };

  const bind =
    (area: ChipArea): BindChip =>
    (chip) => {
      const key = buttonKey({ area, chip });
      return {
        ref: (el) => {
          if (el) buttons.current.set(key, el);
          else buttons.current.delete(key);
        },
        tabIndex: entryItem(areas[area], memory[area])?.chip === chip ? 0 : -1,
        onFocus: () => remember(area, chip),
        onKeyDown: (event) => handleKeyDown({ area, chip }, event),
      };
    };

  return { bindPool: bind("pool"), bindSlot: bind("slot"), focus };
}
