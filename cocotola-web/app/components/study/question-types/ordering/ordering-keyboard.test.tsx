import { fireEvent, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  answerArea,
  buttonTexts,
  chip,
  fixed,
  focusChip,
  makeContent,
  pool,
  renderQuestion,
  revealedAnswer,
  THREE_CHIPS,
  tabStops,
  tapChips,
} from "./test-helpers";

const KEYBOARD_HELP_ID = "ordering-keyboard-help";
const SHIFT_LEFT = "{Shift>}{ArrowLeft}{/Shift}";
const SHIFT_RIGHT = "{Shift>}{ArrowRight}{/Shift}";

describe("ordering question keyboard", () => {
  beforeEach(() => {
    vi.spyOn(Math, "random").mockReturnValue(0);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("Tab stops", () => {
    it("should stop at the first pool chip only", () => {
      // given / when
      renderQuestion(makeContent(THREE_CHIPS));

      // then
      expect(tabStops(pool())).toEqual(["are"]);
    });

    it("should not stop in the answer area while it is empty", () => {
      // given / when
      renderQuestion(makeContent(THREE_CHIPS));

      // then
      expect(tabStops(answerArea())).toEqual([]);
    });

    it("should stop at one chip in the answer area once chips are placed", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));

      // when
      await tapChips(user, ["are", "kind"]);

      // then
      expect(tabStops(answerArea())).toEqual(["Are"]);
    });

    it("should not stop in the pool once every chip is placed", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));

      // when
      await tapChips(user, ["are", "kind", "you"]);

      // then
      expect(tabStops(pool())).toEqual([]);
    });

    it("should leave the pool with the next Tab", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await user.tab();

      // when
      await user.tab();

      // then
      expect(pool()).not.toContainElement(document.activeElement as HTMLElement);
    });
  });

  describe("moving within an area", () => {
    it("should place the chip chosen with an arrow by Space", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await user.tab();
      await user.keyboard("{ArrowRight}");

      // when
      await user.keyboard(" ");

      // then
      expect(buttonTexts(answerArea())).toEqual(["Kind"]);
    });

    it("should move to the last chip with End", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await user.tab();

      // when
      await user.keyboard("{End}");

      // then
      expect(within(pool()).getByRole("button", { name: "you" })).toHaveFocus();
    });

    it("should move to the first chip with Home", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await user.tab();
      await user.keyboard("{End}");

      // when
      await user.keyboard("{Home}");

      // then
      expect(within(pool()).getByRole("button", { name: "are" })).toHaveFocus();
    });

    it("should keep the focus on the first chip with ArrowLeft", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await user.tab();

      // when
      await user.keyboard("{ArrowLeft}");

      // then
      expect(within(pool()).getByRole("button", { name: "are" })).toHaveFocus();
    });

    it("should prevent the default action of an arrow at the edge", () => {
      // given
      renderQuestion(makeContent(THREE_CHIPS));

      // when
      const notPrevented = fireEvent.keyDown(within(pool()).getByRole("button", { name: "are" }), {
        key: "ArrowLeft",
      });

      // then
      expect(notPrevented).toBe(false);
    });

    it("should leave an arrow with Alt to the browser", () => {
      // given
      renderQuestion(makeContent(THREE_CHIPS));

      // when
      const notPrevented = fireEvent.keyDown(within(pool()).getByRole("button", { name: "kind" }), {
        key: "ArrowLeft",
        altKey: true,
      });

      // then
      expect(notPrevented).toBe(true);
    });

    it("should skip empty slots in the answer area", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["you", "are", "kind"]);
      await user.click(within(answerArea()).getByRole("button", { name: "are" }));
      focusChip(answerArea(), "You");

      // when
      await user.keyboard("{ArrowRight}");

      // then
      expect(within(answerArea()).getByRole("button", { name: "kind" })).toHaveFocus();
    });
  });

  describe("moving between areas", () => {
    it("should move up to the first answer chip when the answer area is new", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind"]);

      // when
      await user.keyboard("{ArrowUp}");

      // then
      expect(within(answerArea()).getByRole("button", { name: "Are" })).toHaveFocus();
    });

    it("should move down to the pool chip focused last", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are"]);
      await user.keyboard("{ArrowRight}");
      await user.keyboard("{ArrowUp}");

      // when
      await user.keyboard("{ArrowDown}");

      // then
      expect(within(pool()).getByRole("button", { name: "you" })).toHaveFocus();
    });

    it("should move up to the nearest slot chip when the one focused last was returned", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["you", "are", "kind"]);
      await user.click(within(answerArea()).getByRole("button", { name: "kind" }));

      // when
      await user.keyboard("{ArrowUp}");

      // then
      expect(within(answerArea()).getByRole("button", { name: "are" })).toHaveFocus();
    });

    it("should Tab back into the nearest slot chip when the one focused last was returned", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["you", "are", "kind"]);
      await user.click(within(answerArea()).getByRole("button", { name: "kind" }));

      // when
      await user.tab({ shift: true });

      // then
      expect(within(answerArea()).getByRole("button", { name: "are" })).toHaveFocus();
    });
  });

  describe("Backspace", () => {
    it("should return the chip in the last slot", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind"]);

      // when
      await user.keyboard("{Backspace}");

      // then
      expect(buttonTexts(answerArea())).toEqual(["Are"]);
    });

    it("should move the focus to the returned chip", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind"]);

      // when
      await user.keyboard("{Backspace}");

      // then
      expect(within(pool()).getByRole("button", { name: "kind" })).toHaveFocus();
    });

    it("should return the chip in the last slot from the answer area", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind"]);
      focusChip(answerArea(), "Are");

      // when
      await user.keyboard("{Backspace}");

      // then
      expect(buttonTexts(answerArea())).toEqual(["Are"]);
    });

    it("should return the chip in the last slot rather than the chip placed last", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["you", "are", "kind"]);
      await user.click(within(answerArea()).getByRole("button", { name: "You" }));
      await tapChips(user, ["you"]);

      // when
      await user.keyboard("{Backspace}");

      // then
      expect(buttonTexts(answerArea())).toEqual(["You", "are"]);
    });

    it("should update the answer to be sent", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind", "you"]);

      // when
      await user.keyboard("{Backspace}");

      // then
      expect(screen.getByRole("button", { name: "Check" })).toBeDisabled();
    });

    it("should leave the pool as it is when the answer area is empty", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await user.tab();

      // when
      await user.keyboard("{Backspace}");

      // then
      expect(buttonTexts(pool())).toEqual(["are", "kind", "you"]);
    });

    it("should prevent the default action when the answer area is empty", () => {
      // given
      renderQuestion(makeContent(THREE_CHIPS));

      // when
      const notPrevented = fireEvent.keyDown(within(pool()).getByRole("button", { name: "are" }), {
        key: "Backspace",
      });

      // then
      expect(notPrevented).toBe(false);
    });
  });

  describe("Shift + an arrow in the answer area", () => {
    it("should swap the chip with the next one", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind", "you"]);
      focusChip(answerArea(), "Are");

      // when
      await user.keyboard(SHIFT_RIGHT);

      // then
      expect(buttonTexts(answerArea())).toEqual(["Kind", "are", "you"]);
    });

    it("should swap the chip with the previous one", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind", "you"]);
      focusChip(answerArea(), "you");

      // when
      await user.keyboard(SHIFT_LEFT);

      // then
      expect(buttonTexts(answerArea())).toEqual(["Are", "you", "kind"]);
    });

    it("should keep the focus on the moved chip", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind", "you"]);
      focusChip(answerArea(), "Are");

      // when
      await user.keyboard(SHIFT_RIGHT);

      // then
      expect(within(answerArea()).getByRole("button", { name: "are" })).toHaveFocus();
    });

    it("should move the chip into an empty slot next to it", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind", "you"]);
      await user.click(within(answerArea()).getByRole("button", { name: "kind" }));
      focusChip(answerArea(), "you");

      // when
      await user.keyboard(SHIFT_LEFT);

      // then
      expect(within(answerArea()).getByText("Blank 3")).toBeInTheDocument();
    });

    it("should keep the focus on the chip moved into an empty slot", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind", "you"]);
      await user.click(within(answerArea()).getByRole("button", { name: "kind" }));
      focusChip(answerArea(), "you");

      // when
      await user.keyboard(SHIFT_LEFT);

      // then
      expect(within(answerArea()).getByRole("button", { name: "you" })).toHaveFocus();
    });

    it("should make the moved chip the Tab stop of the answer area", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind", "you"]);
      focusChip(answerArea(), "Are");

      // when
      await user.keyboard(SHIFT_RIGHT);

      // then
      expect(tabStops(answerArea())).toEqual(["are"]);
    });

    it("should swap across a fixed text", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent([chip("you"), chip("are"), fixed("very"), chip("kind")]));
      await tapChips(user, ["are", "kind", "you"]);
      focusChip(answerArea(), "kind");

      // when
      await user.keyboard(SHIFT_RIGHT);

      // then
      expect(buttonTexts(answerArea())).toEqual(["Are", "you", "kind"]);
    });

    it("should keep the order at the edge", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind", "you"]);
      focusChip(answerArea(), "Are");

      // when
      await user.keyboard(SHIFT_LEFT);

      // then
      expect(buttonTexts(answerArea())).toEqual(["Are", "kind", "you"]);
    });

    it("should prevent the default action at the edge", async () => {
      // given
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind", "you"]);

      // when
      const notPrevented = fireEvent.keyDown(
        within(answerArea()).getByRole("button", { name: "Are" }),
        { key: "ArrowLeft", shiftKey: true },
      );

      // then
      expect(notPrevented).toBe(false);
    });

    it("should update the answer to be sent", async () => {
      // given
      const user = userEvent.setup();
      const onAnswer = renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["are", "kind", "you"]);
      focusChip(answerArea(), "Are");
      await user.keyboard(SHIFT_RIGHT);

      // when
      await user.click(screen.getByRole("button", { name: "Check" }));
      await user.click(screen.getByRole("button", { name: "Next" }));

      // then
      expect(onAnswer).toHaveBeenCalledWith({
        correct: false,
        payload: { type: "ordering", regions: [["kind", "are", "you"]] },
      });
    });
  });

  describe("keyboard help", () => {
    it("should describe the pool with the keyboard help", () => {
      // given / when
      renderQuestion(makeContent(THREE_CHIPS));

      // then
      expect(pool()).toHaveAttribute("aria-describedby", screen.getByTestId(KEYBOARD_HELP_ID).id);
    });

    it("should describe the answer area with the keyboard help", () => {
      // given / when
      renderQuestion(makeContent(THREE_CHIPS));

      // then
      expect(answerArea()).toHaveAttribute(
        "aria-describedby",
        screen.getByTestId(KEYBOARD_HELP_ID).id,
      );
    });
  });

  describe("after Check", () => {
    async function check() {
      const user = userEvent.setup();
      renderQuestion(makeContent(THREE_CHIPS));
      await tapChips(user, ["you", "are", "kind"]);
      await user.click(screen.getByRole("button", { name: "Check" }));
    }

    it("should not expose the answer area as a toolbar", async () => {
      // given / when
      await check();

      // then
      expect(screen.queryByRole("toolbar", { name: "Your answer" })).not.toBeInTheDocument();
    });

    it("should expose the answer area as a named group", async () => {
      // given / when
      await check();

      // then
      expect(revealedAnswer()).toBeInTheDocument();
    });

    it("should drop the keyboard help from the answer area", async () => {
      // given / when
      await check();

      // then
      expect(revealedAnswer()).not.toHaveAttribute("aria-describedby");
    });

    it("should hide the keyboard help", async () => {
      // given / when
      await check();

      // then
      expect(screen.queryByTestId(KEYBOARD_HELP_ID)).not.toBeInTheDocument();
    });
  });
});
