import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { type DashboardResponse, getDashboard } from "./dashboard.server";
import { catchError, jsonResponse, newClient, unauthorizedResponse } from "./test-helpers";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

const QUESTION_BASE_URL = "http://localhost:8090";

const TODAY = "2026-11-23";
const TIMEZONE = "America/Los_Angeles";

function buildDashboardResponse(): DashboardResponse {
  return {
    from: "2026-11-17",
    to: TODAY,
    days: [
      { date: "2026-11-22", answeredCount: 12, correctCount: 9 },
      { date: TODAY, answeredCount: 4, correctCount: 3 },
    ],
    currentStreak: 2,
    longestStreak: 15,
    todayCount: 4,
    todayCorrect: 3,
    activeDays: 2,
    totalAnswered: 16,
    totalCorrect: 12,
  };
}

function requestedUrl(): string {
  return fetchMock.mock.calls[0][0] as string;
}

function requestedHeaders(): Record<string, string> {
  return (fetchMock.mock.calls[0][1] as RequestInit).headers as Record<string, string>;
}

beforeEach(() => {
  vi.stubEnv("QUESTION_BASE_URL", QUESTION_BASE_URL);
  vi.spyOn(console, "info").mockImplementation(() => {});
  vi.spyOn(console, "error").mockImplementation(() => {});
});

afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
  fetchMock.mockReset();
});

describe("getDashboard", () => {
  it("should return the dashboard on success", async () => {
    // given
    const response = buildDashboardResponse();
    fetchMock.mockResolvedValue(jsonResponse(response));

    // when
    const result = await getDashboard(newClient(), TODAY, TIMEZONE);

    // then
    expect(result).toEqual(response);
  });

  it("should send GET request with the access token and the local date headers", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildDashboardResponse()));

    // when
    await getDashboard(newClient(), TODAY, TIMEZONE, 30);

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/study/dashboard?days=30`, {
      method: "GET",
      headers: {
        Authorization: "Bearer test-token",
        "X-Local-Date": TODAY,
        "X-Local-Timezone": TIMEZONE,
      },
    });
  });

  it("should send the local date header even when it is empty", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildDashboardResponse()));

    // when
    await getDashboard(newClient(), "", TIMEZONE);

    // then
    expect(requestedHeaders()["X-Local-Date"]).toBe("");
  });

  it("should request 365 days when days is omitted", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildDashboardResponse()));

    // when
    await getDashboard(newClient(), TODAY, TIMEZONE);

    // then
    expect(requestedUrl()).toBe(`${QUESTION_BASE_URL}/api/v1/study/dashboard?days=365`);
  });

  it.each([
    [3, 7],
    [1000, 730],
    [30.9, 30],
    [Number.NaN, 365],
  ])("should clamp days=%s to %s", async (days, expected) => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildDashboardResponse()));

    // when
    await getDashboard(newClient(), TODAY, TIMEZONE, days);

    // then
    expect(requestedUrl()).toBe(`${QUESTION_BASE_URL}/api/v1/study/dashboard?days=${expected}`);
  });

  it("should accept a dashboard with no days", async () => {
    // given
    const response = { ...buildDashboardResponse(), days: [] };
    fetchMock.mockResolvedValue(jsonResponse(response));

    // when
    const result = await getDashboard(newClient(), TODAY, TIMEZONE);

    // then
    expect(result).toEqual(response);
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(getDashboard(newClient(), TODAY, TIMEZONE)).rejects.toMatchObject({
      status: 500,
    });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(getDashboard(newClient(), TODAY, TIMEZONE));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when the response lacks currentStreak", async () => {
    // given
    const { currentStreak: _, ...withoutStreak } = buildDashboardResponse();
    fetchMock.mockResolvedValue(jsonResponse(withoutStreak));

    // when / then
    await expect(getDashboard(newClient(), TODAY, TIMEZONE)).rejects.toMatchObject({
      status: 502,
    });
  });

  it("should throw 502 Response when a day in the response has a non-numeric answeredCount", async () => {
    // given
    const response = {
      ...buildDashboardResponse(),
      days: [{ date: TODAY, answeredCount: "4", correctCount: 3 }],
    };
    fetchMock.mockResolvedValue(jsonResponse(response));

    // when / then
    await expect(getDashboard(newClient(), TODAY, TIMEZONE)).rejects.toMatchObject({
      status: 502,
    });
  });

  it("should throw 502 Response when the response has a non-integer currentStreak", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ ...buildDashboardResponse(), currentStreak: 1.5 }));

    // when / then
    await expect(getDashboard(newClient(), TODAY, TIMEZONE)).rejects.toMatchObject({
      status: 502,
    });
  });

  it("should throw 502 Response when the response has a negative todayCount", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ ...buildDashboardResponse(), todayCount: -1 }));

    // when / then
    await expect(getDashboard(newClient(), TODAY, TIMEZONE)).rejects.toMatchObject({
      status: 502,
    });
  });
});
