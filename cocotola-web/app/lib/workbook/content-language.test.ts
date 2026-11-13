import { describe, expect, it } from "vitest";
import { DEFAULT_CONTENT_LANGUAGE, toContentLanguage } from "./content-language";

describe("toContentLanguage", () => {
  it("should return the value when it is a content language", () => {
    // given
    const value = "ko";

    // when
    const result = toContentLanguage(value);

    // then
    expect(result).toBe("ko");
  });

  it.each([
    { name: "a region-tagged code", value: "en-US" },
    { name: "an unknown code", value: "xx" },
    { name: "undefined", value: undefined },
  ])("should return the default language when the value is $name", ({ value }) => {
    // given / when
    const result = toContentLanguage(value);

    // then
    expect(result).toBe(DEFAULT_CONTENT_LANGUAGE);
  });
});
