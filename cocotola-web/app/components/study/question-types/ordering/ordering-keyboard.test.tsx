import { act, fireEvent, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  answerArea,
  buttonTexts,
  makeContent,
  pool,
  renderQuestion,
  revealedAnswer,
  THREE_CHIPS,
  tabStops,
  tapChips,
} from "./test-helpers";

const KEYBOARD_HELP_ID = "ordering-keyboard-help";

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
      act(() => within(answerArea()).getByRole("button", { name: "You" }).focus());

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
