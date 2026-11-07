import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { OrderingSegment } from "~/components/study/question-types/ordering/answer";
import { OrderingPreview } from "./ordering-preview";

const chip = (text: string): OrderingSegment => ({ type: "chip", text });
const fixed = (text: string): OrderingSegment => ({ type: "fixed", text });

function previewTexts(segments: OrderingSegment[]): (string | null)[] {
  render(<OrderingPreview segments={segments} lang="en" label="preview" />);
  const list = screen.getByRole("list", { name: "preview" });
  return within(list)
    .getAllByRole("listitem")
    .map((item) => item.textContent);
}

describe("OrderingPreview", () => {
  it("should capitalize the first chip", () => {
    // given
    const segments = [chip("i"), chip("am")];

    // when
    const texts = previewTexts(segments);

    // then
    expect(texts[0]).toBe("I");
  });

  it("should capitalize a chip right after a sentence end", () => {
    // given
    const segments = [chip("i"), chip("am"), fixed("."), chip("you"), chip("are")];

    // when
    const texts = previewTexts(segments);

    // then
    expect(texts[3]).toBe("You");
  });

  it("should keep a chip in the middle of a sentence as written", () => {
    // given
    const segments = [chip("i"), fixed(","), chip("you")];

    // when
    const texts = previewTexts(segments);

    // then
    expect(texts[2]).toBe("you");
  });
});
