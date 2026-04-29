import { describe, expect, it } from "vitest";
import { formatDate } from "./date";

describe("formatDate", () => {
  it("should format en-US dates", () => {
    // given
    const dateStr = "2026-01-15T12:00:00Z";

    // when
    const result = formatDate(dateStr, "en");

    // then
    expect(result).toMatch(/Jan/);
    expect(result).toMatch(/2026/);
  });

  it("should format ja-JP dates", () => {
    // given
    const dateStr = "2026-01-15T12:00:00Z";

    // when
    const result = formatDate(dateStr, "ja");

    // then
    expect(result).toMatch(/2026/);
    expect(result).toMatch(/1月/);
  });

  it("should format ko-KR dates", () => {
    // given
    const dateStr = "2026-01-15T12:00:00Z";

    // when
    const result = formatDate(dateStr, "ko");

    // then
    expect(result).toMatch(/2026/);
  });

  it("should fall back to the raw locale when not in the map", () => {
    // given
    const dateStr = "2026-01-15T12:00:00Z";

    // when
    const result = formatDate(dateStr, "fr");

    // then
    expect(result).toMatch(/2026/);
  });
});
