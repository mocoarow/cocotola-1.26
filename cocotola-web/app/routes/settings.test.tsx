import { act, render, screen, waitFor } from "@testing-library/react";
import i18nInstance from "i18next";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import "~/i18n/config";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

// Cover every concrete export the route module imports, even ones the
// current tests do not exercise, so future tests (or any reorganization
// that pulls a new helper into route scope) do not hit `undefined`.
vi.mock("~/lib/api/user-setting.server", () => ({
  getUserPreferences: vi.fn(),
  updateUserLanguage: vi.fn(),
  updateUserDailyGoal: vi.fn(),
  updateUserTimezone: vi.fn(),
}));

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useLoaderData: vi.fn(),
    useActionData: vi.fn(),
    useNavigation: vi.fn(() => ({ state: "idle", formData: undefined })),
    useFetcher: vi.fn(),
    Form: ({ children }: { children: ReactNode }) => <form>{children}</form>,
  };
});

import { useActionData, useFetcher, useLoaderData } from "react-router";
import {
  updateUserDailyGoal,
  updateUserLanguage,
  updateUserTimezone,
} from "~/lib/api/user-setting.server";
import { requireAuth } from "~/lib/auth/require-auth.server";

import SettingsPage, { action } from "./settings";

const mockedUseLoaderData = vi.mocked(useLoaderData);
const mockedUseActionData = vi.mocked(useActionData);
const mockedUseFetcher = vi.mocked(useFetcher);

function makeLoaderData() {
  return {
    loginId: "user@example.com",
    organizationName: "Acme Corp",
    language: "en",
    dailyGoal: 25,
    timezone: "America/Los_Angeles",
  };
}

type FetcherShape = {
  state: "idle" | "submitting" | "loading";
  submit: ReturnType<typeof vi.fn>;
  data: unknown;
  Form: (props: { children: ReactNode }) => ReactNode;
};

function idleFetcher(data: unknown = undefined): FetcherShape {
  return {
    state: "idle",
    submit: vi.fn(),
    data,
    Form: ({ children }) => <form>{children}</form>,
  };
}

describe("SettingsPage", () => {
  beforeEach(() => {
    // Default fetcher: idle with no data. Individual tests override the
    // return value to simulate language-save success/failure.
    mockedUseFetcher.mockReturnValue(idleFetcher() as unknown as ReturnType<typeof useFetcher>);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("should_renderHeading_when_loaderDataIsAvailable", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue(undefined);

    // when
    render(<SettingsPage />);

    // then: role-based assertion keeps the test locale-stable
    expect(screen.getByRole("heading", { level: 1 })).toBeInTheDocument();
  });

  it("should_renderAccountInfo_when_loaderProvidesIdentity", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue(undefined);

    // when
    render(<SettingsPage />);

    // then
    expect(screen.getByTestId("settings-login-id").textContent).toBe("user@example.com");
    expect(screen.getByTestId("settings-organization").textContent).toBe("Acme Corp");
  });

  it("should_renderLanguageSelector_when_loaderProvidesLanguage", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue(undefined);

    // when
    render(<SettingsPage />);

    // then
    expect(screen.getByTestId("settings-language-select")).toBeInTheDocument();
  });

  it("should_renderStudyPreferencesForm_when_loaderProvidesValues", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue(undefined);

    // when
    render(<SettingsPage />);

    // then
    const goalInput = screen.getByLabelText(/Daily goal/i) as HTMLInputElement;
    expect(goalInput.value).toBe("25");
    const tzInput = screen.getByLabelText(/Timezone/i) as HTMLInputElement;
    expect(tzInput.value).toBe("America/Los_Angeles");
  });

  it("should_showSavedConfirmation_when_actionReportsStudySaved", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue({
      ok: true,
      intent: "saveStudyPreferences",
      savedAt: "2026-06-14T10:00:00Z",
    });

    // when
    render(<SettingsPage />);

    // then
    expect(screen.getByTestId("settings-saved")).toBeInTheDocument();
  });

  it("should_showPartialFailureMessage_when_actionReportsFailedFields", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue({
      ok: false,
      intent: "saveStudyPreferences",
      failedFields: ["timezone"],
    });

    // when
    render(<SettingsPage />);

    // then
    const banner = screen.getByTestId("settings-partial-failure");
    expect(banner).toBeInTheDocument();
    expect(banner.textContent).toMatch(/timezone/);
  });

  it("should_notShowStudySavedBanner_when_actionDataIsLanguageOnly", () => {
    // given: language-only saves return intent="saveLanguage" — the study
    // form's "saved" confirmation must not appear, otherwise the user sees
    // a false-positive "preferences saved" message that includes goal/tz.
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue({
      ok: true,
      intent: "saveLanguage",
      language: "ja",
    });

    // when
    render(<SettingsPage />);

    // then
    expect(screen.queryByTestId("settings-saved")).not.toBeInTheDocument();
  });
});

// The language commit path is the M2 fix: i18n.changeLanguage must run
// only after the server confirms the save, so a failed/redirected/missing
// payload leaves the UI on the pre-change language and stays consistent
// with the persisted preference.
describe("SettingsPage language commit gate", () => {
  let changeLanguageSpy: ReturnType<typeof vi.spyOn>;
  let originalLanguage: string;

  beforeEach(async () => {
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue(undefined);

    // Reset i18n to a known starting language so each test starts from
    // the same baseline regardless of order. Wrap in act() because
    // changeLanguage triggers a sync emit on subscribed react-i18next
    // hooks elsewhere in the test environment.
    originalLanguage = i18nInstance.language;
    await act(async () => {
      await i18nInstance.changeLanguage("en");
    });

    // spyOn AFTER the reset so the baseline call does not pollute the
    // call count we assert on.
    changeLanguageSpy = vi.spyOn(i18nInstance, "changeLanguage");
  });

  afterEach(async () => {
    changeLanguageSpy.mockRestore();
    await act(async () => {
      await i18nInstance.changeLanguage(originalLanguage);
    });
    vi.clearAllMocks();
  });

  it("should_commitI18nChange_when_languageFetcherReturnsSuccess", async () => {
    // given: fetcher reports a successful language save with the new code
    mockedUseFetcher.mockReturnValue(
      idleFetcher({
        ok: true,
        intent: "saveLanguage",
        language: "ja",
      }) as unknown as ReturnType<typeof useFetcher>,
    );

    // when
    render(<SettingsPage />);

    // then: the effect commits the language change exactly once
    await waitFor(() => {
      expect(changeLanguageSpy).toHaveBeenCalledWith("ja");
    });
    expect(changeLanguageSpy).toHaveBeenCalledTimes(1);
  });

  it("should_notCommitI18nChange_when_languageFetcherDataIsUndefined", async () => {
    // given: initial mount, no submission yet
    mockedUseFetcher.mockReturnValue(idleFetcher() as unknown as ReturnType<typeof useFetcher>);

    // when
    render(<SettingsPage />);

    // then: the effect must not commit anything until data arrives
    // (waitFor + a microtask flush gives the effect a chance to fire)
    await waitFor(() => {
      expect(screen.getByTestId("settings-language-select")).toBeInTheDocument();
    });
    expect(changeLanguageSpy).not.toHaveBeenCalled();
  });

  it("should_notCommitI18nChange_when_languageFetcherReturnsStudyIntentShape", async () => {
    // given: a study-intent payload landing on the language fetcher
    // (defensive: this should never happen via the real client flow,
    // but the gate keeps us safe from any future cross-wired fetcher).
    mockedUseFetcher.mockReturnValue(
      idleFetcher({
        ok: true,
        intent: "saveStudyPreferences",
        savedAt: "2026-06-14T10:00:00Z",
      }) as unknown as ReturnType<typeof useFetcher>,
    );

    // when
    render(<SettingsPage />);

    // then: discriminator rejects non-language payloads
    await waitFor(() => {
      expect(screen.getByTestId("settings-language-select")).toBeInTheDocument();
    });
    expect(changeLanguageSpy).not.toHaveBeenCalled();
  });

  it("should_notCommitI18nChange_when_languageFetcherReturnsUnsupportedCode", async () => {
    // given: tampered payload claiming a code that is not in the allow-list
    mockedUseFetcher.mockReturnValue(
      idleFetcher({
        ok: true,
        intent: "saveLanguage",
        language: "xx",
      }) as unknown as ReturnType<typeof useFetcher>,
    );

    // when
    render(<SettingsPage />);

    // then: isLanguageSuccess rejects unsupported codes even if the
    // wire shape matches
    await waitFor(() => {
      expect(screen.getByTestId("settings-language-select")).toBeInTheDocument();
    });
    expect(changeLanguageSpy).not.toHaveBeenCalled();
  });

  it("should_notCommitI18nChange_when_fetcherIsStillSubmitting", async () => {
    // given: a submitting fetcher — the effect early-returns on non-idle
    // state so the change does not race ahead of the server response.
    mockedUseFetcher.mockReturnValue({
      state: "submitting",
      submit: vi.fn(),
      data: { ok: true, intent: "saveLanguage", language: "ja" },
      Form: ({ children }: { children: ReactNode }) => <form>{children}</form>,
    } as unknown as ReturnType<typeof useFetcher>);

    // when
    render(<SettingsPage />);

    // then
    await waitFor(() => {
      expect(screen.getByTestId("settings-language-select")).toBeInTheDocument();
    });
    expect(changeLanguageSpy).not.toHaveBeenCalled();
  });
});

describe("settings action", () => {
  function buildActionRequest(fields: Record<string, string>): Request {
    const form = new URLSearchParams(fields);
    return new Request("http://localhost/settings", {
      method: "POST",
      body: form.toString(),
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
    });
  }

  async function callAction(request: Request) {
    return action({ request } as unknown as Parameters<typeof action>[0]);
  }

  beforeEach(() => {
    vi.mocked(requireAuth).mockResolvedValue({
      accessToken: "test-token",
      refreshToken: "test-refresh",
    });
    vi.mocked(updateUserLanguage).mockResolvedValue(undefined);
    vi.mocked(updateUserDailyGoal).mockResolvedValue(undefined);
    vi.mocked(updateUserTimezone).mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("should_returnLanguageInPayload_when_languageSaveSucceeds", async () => {
    // given
    const request = buildActionRequest({
      intent: "saveLanguage",
      language: "ja",
    });

    // when
    const result = await callAction(request);

    // then: the saved code must round-trip in the response so the
    // client's effect can commit i18n.changeLanguage without trusting
    // the unverified pending state.
    expect(result).toEqual({ ok: true, intent: "saveLanguage", language: "ja" });
    expect(vi.mocked(updateUserLanguage)).toHaveBeenCalledWith(request, "test-token", "ja");
  });

  it("should_throw400_when_languageIsUnsupported", async () => {
    // given
    const request = buildActionRequest({
      intent: "saveLanguage",
      language: "xx",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
    expect(vi.mocked(updateUserLanguage)).not.toHaveBeenCalled();
  });

  it("should_notInvokeStudyUpdaters_when_intentIsLanguage", async () => {
    // given: the M3 split means handleLanguageIntent must never reach
    // the study updaters even though they live in the same module.
    const request = buildActionRequest({
      intent: "saveLanguage",
      language: "ja",
    });

    // when
    await callAction(request);

    // then
    expect(vi.mocked(updateUserDailyGoal)).not.toHaveBeenCalled();
    expect(vi.mocked(updateUserTimezone)).not.toHaveBeenCalled();
  });

  it("should_returnOk_when_bothStudyUpdatesSucceed", async () => {
    // given
    const request = buildActionRequest({
      intent: "saveStudyPreferences",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when
    const result = await callAction(request);

    // then
    expect(result).toMatchObject({ ok: true, intent: "saveStudyPreferences" });
    expect(vi.mocked(updateUserDailyGoal)).toHaveBeenCalledWith(request, "test-token", 20);
    expect(vi.mocked(updateUserTimezone)).toHaveBeenCalledWith(request, "test-token", "Asia/Tokyo");
  });

  it("should_notInvokeLanguageUpdater_when_intentIsStudy", async () => {
    // given
    const request = buildActionRequest({
      intent: "saveStudyPreferences",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when
    await callAction(request);

    // then
    expect(vi.mocked(updateUserLanguage)).not.toHaveBeenCalled();
  });

  it("should_throw400_when_intentIsUnknown", async () => {
    // given
    const request = buildActionRequest({
      intent: "deleteEverything",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
  });

  it("should_throw400_when_dailyGoalBelowMin", async () => {
    // given
    const request = buildActionRequest({
      intent: "saveStudyPreferences",
      dailyGoal: "0",
      timezone: "Asia/Tokyo",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
    expect(vi.mocked(updateUserDailyGoal)).not.toHaveBeenCalled();
  });

  it("should_throw400_when_dailyGoalHasTrailingGarbage", async () => {
    // given
    // Number.parseInt("30abc", 10) returns 30 silently. The strict
    // /^\d+$/ guard must convert this into NaN so the range check
    // rejects the tampered body rather than accepting "30" out of it.
    const request = buildActionRequest({
      intent: "saveStudyPreferences",
      dailyGoal: "30abc",
      timezone: "Asia/Tokyo",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
    expect(vi.mocked(updateUserDailyGoal)).not.toHaveBeenCalled();
  });

  it("should_throw400_when_dailyGoalAboveMax", async () => {
    // given
    const request = buildActionRequest({
      intent: "saveStudyPreferences",
      dailyGoal: "501",
      timezone: "Asia/Tokyo",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
  });

  it("should_throw400_when_timezoneShapeInvalid", async () => {
    // given
    const request = buildActionRequest({
      intent: "saveStudyPreferences",
      dailyGoal: "20",
      timezone: "Asia Tokyo", // space → fails character-class regex
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
  });

  it("should_returnFailedTimezoneField_when_timezoneUpdaterRejects", async () => {
    // given
    vi.mocked(updateUserTimezone).mockRejectedValue(
      new Response("server unavailable", { status: 500 }),
    );
    const request = buildActionRequest({
      intent: "saveStudyPreferences",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when
    const result = await callAction(request);

    // then
    expect(result).toMatchObject({
      ok: false,
      intent: "saveStudyPreferences",
      failedFields: ["timezone"],
    });
  });

  it("should_returnBothFailedFields_when_bothUpdatersReject", async () => {
    // given
    vi.mocked(updateUserDailyGoal).mockRejectedValue(
      new Response("server unavailable", { status: 500 }),
    );
    vi.mocked(updateUserTimezone).mockRejectedValue(
      new Response("server unavailable", { status: 500 }),
    );
    const request = buildActionRequest({
      intent: "saveStudyPreferences",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when
    const result = await callAction(request);

    // then
    expect(result).toMatchObject({
      ok: false,
      intent: "saveStudyPreferences",
      failedFields: ["dailyGoal", "timezone"],
    });
  });

  it("should_propagateRedirect_when_updaterThrowsRedirectResponse", async () => {
    // given
    // redirectOnUnauthorized throws a redirect Response on 401. The
    // savePreferenceField helper must propagate any 3xx Response
    // (Location header present) so React Router can navigate, instead
    // of recording it as a field-level failure.
    const loginRedirect = new Response(null, {
      status: 302,
      headers: { Location: "/login" },
    });
    vi.mocked(updateUserDailyGoal).mockRejectedValue(loginRedirect);
    const request = buildActionRequest({
      intent: "saveStudyPreferences",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBe(loginRedirect);
    expect(vi.mocked(updateUserTimezone)).not.toHaveBeenCalled();
  });
});
