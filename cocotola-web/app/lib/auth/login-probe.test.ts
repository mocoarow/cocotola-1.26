import { afterEach, describe, expect, it, vi } from "vitest";
import { buildLoginProbeCookie, parseLoginProbe, writeLoginProbe } from "./login-probe";

const probe = { pageLoadedAtMs: 1000, clickedAtMs: 2000 };

describe("buildLoginProbeCookie", () => {
  it("should build a lax root cookie expiring ten minutes after the click", () => {
    // given (probe)

    // when
    const cookie = buildLoginProbeCookie(probe);

    // then
    expect(cookie).toEqual({
      name: "cocotola_login_probe",
      value: "1000.2000",
      path: "/",
      sameSite: "lax",
      expires: 602_000,
    });
  });
});

describe("writeLoginProbe", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("should set the probe cookie through the Cookie Store API", async () => {
    // given
    const set = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal("cookieStore", { set });

    // when
    await writeLoginProbe(probe);

    // then
    expect(set).toHaveBeenCalledWith(buildLoginProbeCookie(probe));
  });

  it("should resolve when the Cookie Store API is unavailable", async () => {
    // given
    vi.stubGlobal("cookieStore", undefined);

    // when / then
    await expect(writeLoginProbe(probe)).resolves.toBeUndefined();
  });

  it("should resolve when setting the cookie fails", async () => {
    // given
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.stubGlobal("cookieStore", { set: vi.fn().mockRejectedValue(new Error("denied")) });

    // when / then
    await expect(writeLoginProbe(probe)).resolves.toBeUndefined();
  });
});

describe("parseLoginProbe", () => {
  it("should return timestamps when value is well-formed", () => {
    // given
    const value = "1000.2000";

    // when
    const result = parseLoginProbe(value);

    // then
    expect(result).toEqual(probe);
  });

  it.each([
    { name: "empty", value: "" },
    { name: "missing separator", value: "1000" },
    { name: "non-numeric", value: "abc.2000" },
    { name: "extra part", value: "1000.2000.3000" },
  ])("should return null when value is $name", ({ value }) => {
    // given (value)

    // when
    const result = parseLoginProbe(value);

    // then
    expect(result).toBeNull();
  });
});
