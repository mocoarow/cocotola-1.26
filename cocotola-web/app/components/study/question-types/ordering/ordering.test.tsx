import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  answerArea,
  buttonTexts,
  chip,
  fixed,
  makeContent,
  pool,
  renderQuestion,
  revealedAnswer,
  THREE_CHIPS,
  tapChips,
} from "./test-helpers";

describe("ordering question", () => {
  beforeEach(() => {
    vi.spyOn(Math, "random").mockReturnValue(0);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should render the source text", () => {
    // given / when
    renderQuestion(makeContent(THREE_CHIPS));

    // then
    expect(screen.getByText("あなたは親切です")).toBeInTheDocument();
  });

  it("should offer every chip in the shuffled order", () => {
    // given / when
    renderQuestion(makeContent(THREE_CHIPS));

    // then
    expect(buttonTexts(pool())).toEqual(["are", "kind", "you"]);
  });

  it("should render fixed texts in the answer area", () => {
    // given / when
    renderQuestion(makeContent([chip("I"), chip("am"), fixed("a student.")]));

    // then
    expect(within(answerArea()).getByText("a student.")).toBeInTheDocument();
  });

  it("should put a tapped chip into the first empty slot", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));

    // when
    await tapChips(user, ["are"]);

    // then
    expect(buttonTexts(answerArea())).toEqual(["Are"]);
  });

  it("should remove a placed chip from the pool", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));

    // when
    await tapChips(user, ["are"]);

    // then
    expect(buttonTexts(pool())).toEqual(["kind", "you"]);
  });

  it("should return a tapped slot chip to the pool", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));
    await tapChips(user, ["are"]);

    // when
    await user.click(within(answerArea()).getByRole("button", { name: "Are" }));

    // then
    expect(buttonTexts(pool())).toEqual(["are", "kind", "you"]);
  });

  it("should fill the emptied slot with the next tapped chip", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));
    await tapChips(user, ["you", "are"]);
    await user.click(within(answerArea()).getByRole("button", { name: "You" }));

    // when
    await tapChips(user, ["kind"]);

    // then
    expect(buttonTexts(answerArea())).toEqual(["Kind", "are"]);
  });

  it("should place chips one after another with the keyboard", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));

    // when
    await user.tab();
    await user.keyboard("{Enter}");
    await user.keyboard("{Enter}");
    await user.keyboard(" ");

    // then
    expect(buttonTexts(answerArea())).toEqual(["Are", "kind", "you"]);
  });

  it("should focus the slot just filled when the last chip leaves the pool", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));

    // when
    await tapChips(user, ["you", "are", "kind"]);

    // then
    expect(within(answerArea()).getByRole("button", { name: "kind" })).toHaveFocus();
  });

  it("should focus the returned chip in the pool when a slot chip is tapped", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));
    await tapChips(user, ["you", "are"]);

    // when
    await user.click(within(answerArea()).getByRole("button", { name: "You" }));

    // then
    expect(within(pool()).getByRole("button", { name: "you" })).toHaveFocus();
  });

  it("should focus the new last pool chip when the last pool chip is placed", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));

    // when
    await tapChips(user, ["you"]);

    // then
    expect(within(pool()).getByRole("button", { name: "kind" })).toHaveFocus();
  });

  it("should keep Check disabled before any chip is placed", () => {
    // given / when
    renderQuestion(makeContent(THREE_CHIPS));

    // then
    expect(screen.getByRole("button", { name: "Check" })).toBeDisabled();
  });

  it("should keep Check disabled while a slot is empty", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));

    // when
    await tapChips(user, ["you", "are"]);

    // then
    expect(screen.getByRole("button", { name: "Check" })).toBeDisabled();
  });

  it("should enable Check when every slot is filled", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));

    // when
    await tapChips(user, ["you", "are", "kind"]);

    // then
    expect(screen.getByRole("button", { name: "Check" })).toBeEnabled();
  });

  it("should disable Check again when a chip is returned after filling every slot", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));
    await tapChips(user, ["you", "are", "kind"]);

    // when
    await user.click(within(answerArea()).getByRole("button", { name: "kind" }));

    // then
    expect(screen.getByRole("button", { name: "Check" })).toBeDisabled();
  });

  it("should send the stored chip texts by region and true when the order is correct", async () => {
    // given
    const user = userEvent.setup();
    const onAnswer = renderQuestion(
      makeContent([chip("you"), chip("are"), fixed("."), chip("so"), chip("kind")]),
    );
    await tapChips(user, ["you", "are", "so", "kind"]);

    // when
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then
    expect(onAnswer).toHaveBeenCalledWith({
      correct: true,
      payload: {
        type: "ordering",
        regions: [
          ["you", "are"],
          ["so", "kind"],
        ],
      },
    });
  });

  it("should send false when the order is wrong", async () => {
    // given
    const user = userEvent.setup();
    const onAnswer = renderQuestion(makeContent(THREE_CHIPS));
    await tapChips(user, ["are", "you", "kind"]);

    // when
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then
    expect(onAnswer).toHaveBeenCalledWith({
      correct: false,
      payload: { type: "ordering", regions: [["are", "you", "kind"]] },
    });
  });

  it("should announce whether each slot is correct after Check", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));
    await tapChips(user, ["are", "you", "kind"]);

    // when
    await user.click(screen.getByRole("button", { name: "Check" }));

    // then
    const slotStatuses = within(revealedAnswer())
      .getAllByText(/^(Correct position|Wrong position)$/)
      .map((el) => el.textContent);
    expect(slotStatuses).toEqual(["Wrong position", "Wrong position", "Correct position"]);
  });

  it("should stop returning chips after Check", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS));
    await tapChips(user, ["are", "you", "kind"]);

    // when
    await user.click(screen.getByRole("button", { name: "Check" }));

    // then
    expect(within(revealedAnswer()).queryAllByRole("button")).toHaveLength(0);
  });

  it("should show the correct order after Check", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent([chip("you"), chip("are"), fixed("."), chip("so"), chip("kind")]));
    await tapChips(user, ["so", "kind", "you", "are"]);

    // when
    await user.click(screen.getByRole("button", { name: "Check" }));

    // then
    const tokens = within(screen.getByRole("list", { name: "Correct order" }))
      .getAllByRole("listitem")
      .map((item) => item.textContent);
    expect(tokens).toEqual(["You", "are", ".", "So", "kind"]);
  });

  it("should not show the correct order before Check", () => {
    // given / when
    renderQuestion(makeContent(THREE_CHIPS));

    // then
    expect(screen.queryByRole("list", { name: "Correct order" })).not.toBeInTheDocument();
  });

  it("should show the explanation after Check", async () => {
    // given
    const user = userEvent.setup();
    renderQuestion(makeContent(THREE_CHIPS, "Subject comes first"));
    await tapChips(user, ["you", "are", "kind"]);

    // when
    await user.click(screen.getByRole("button", { name: "Check" }));

    // then
    expect(screen.getByText("Subject comes first")).toBeInTheDocument();
  });

  it("should not show the explanation before Check", () => {
    // given / when
    renderQuestion(makeContent(THREE_CHIPS, "Subject comes first"));

    // then
    expect(screen.queryByText("Subject comes first")).not.toBeInTheDocument();
  });

  it("should offer Skip when the content has fewer than two chips", () => {
    // given / when
    renderQuestion(makeContent([chip("Hello"), fixed(".")]));

    // then
    expect(screen.getByRole("button", { name: "Skip" })).toBeInTheDocument();
  });
});
