import { describe, expect, it } from "vitest";
import { sameStringSet } from "./string-set";

describe("sameStringSet", () => {
  it("should return true when both arrays describe the same set", () => {
    // given
    const a = ["x", "y"];
    const b = ["y", "x"];

    // when/then
    expect(sameStringSet(a, b)).toBe(true);
  });

  it("should return false when the sets differ", () => {
    // given
    const a = ["x", "y"];
    const b = ["x", "z"];

    // when/then
    expect(sameStringSet(a, b)).toBe(false);
  });

  it("should ignore duplicates so partial-overlap arrays are not falsely equal", () => {
    // given: a length-only comparison would call these equal because both
    // have two entries and "x" is shared, hiding the missing "y".
    const a = ["x", "y"];
    const b = ["x", "x"];

    // when/then
    expect(sameStringSet(a, b)).toBe(false);
  });

  it("should treat a duplicated array as equal to its deduped form", () => {
    // given
    const a = ["x", "x"];
    const b = ["x"];

    // when/then
    expect(sameStringSet(a, b)).toBe(true);
  });

  it("should return true for two empty arrays", () => {
    // when/then
    expect(sameStringSet([], [])).toBe(true);
  });
});
