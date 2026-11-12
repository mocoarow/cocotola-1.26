import type { DndContextProps, DragEndEvent } from "@dnd-kit/core";
import { act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createElement } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { chipDragId, slotDropId } from "./drop";
import {
  answerArea,
  buttonTexts,
  makeContent,
  renderQuestion,
  THREE_CHIPS,
  tapChips,
} from "./test-helpers";

const dnd = vi.hoisted(() => ({ onDragEnd: undefined as DndContextProps["onDragEnd"] }));

vi.mock("@dnd-kit/core", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@dnd-kit/core")>();
  return {
    ...actual,
    DndContext: (props: DndContextProps) => {
      dnd.onDragEnd = props.onDragEnd;
      return createElement(actual.DndContext, props);
    },
  };
});

const ARE = THREE_CHIPS.findIndex((c) => c.text === "are");

function drop(chip: number, slot: number) {
  const onDragEnd = dnd.onDragEnd;
  if (onDragEnd === undefined) throw new Error("DndContext has no onDragEnd");
  act(() =>
    onDragEnd({
      active: { id: chipDragId(chip) },
      over: { id: slotDropId(slot) },
    } as DragEndEvent),
  );
}

describe("ordering question keyboard after a drag", () => {
  beforeEach(() => {
    vi.spyOn(Math, "random").mockReturnValue(0);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should return the chip in the last slot with Backspace", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));
    await tapChips(user, ["are", "kind"]);
    drop(ARE, 2);

    // when
    await user.keyboard("{Backspace}");

    // then
    expect(buttonTexts(answerArea())).toEqual(["kind"]);
  });
});
