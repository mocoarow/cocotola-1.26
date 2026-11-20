import { SpanStatusCode, trace } from "@opentelemetry/api";
import { tracing } from "@opentelemetry/sdk-node";
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";
import { fetchWithIdToken } from "./fetch.server";

const AUDIENCE = "http://localhost:8090";
const BACKEND_URL = `${AUDIENCE}/api/v1/workbook`;

const exporter = new tracing.InMemorySpanExporter();
const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

function finishedSpan(): tracing.ReadableSpan {
  const spans = exporter.getFinishedSpans();
  expect(spans).toHaveLength(1);
  return spans[0];
}

describe("fetchWithIdToken", () => {
  beforeAll(() => {
    trace.setGlobalTracerProvider(
      new tracing.BasicTracerProvider({
        spanProcessors: [new tracing.SimpleSpanProcessor(exporter)],
      }),
    );
  });

  beforeEach(() => {
    vi.stubEnv("APP_ENV", "test");
  });

  afterEach(() => {
    exporter.reset();
    fetchMock.mockReset();
    vi.unstubAllEnvs();
  });

  it("should mark the span as error when the backend returns 5xx", async () => {
    // given
    fetchMock.mockResolvedValue({ status: 500 });

    // when
    await fetchWithIdToken(AUDIENCE, BACKEND_URL);

    // then
    expect(finishedSpan().status.code).toBe(SpanStatusCode.ERROR);
  });

  it("should leave the span status unset when the backend returns 4xx", async () => {
    // given
    fetchMock.mockResolvedValue({ status: 499 });

    // when
    await fetchWithIdToken(AUDIENCE, BACKEND_URL);

    // then
    expect(finishedSpan().status.code).toBe(SpanStatusCode.UNSET);
  });

  it("should leave the span status unset when the backend returns 2xx", async () => {
    // given
    fetchMock.mockResolvedValue({ status: 200 });

    // when
    await fetchWithIdToken(AUDIENCE, BACKEND_URL);

    // then
    expect(finishedSpan().status.code).toBe(SpanStatusCode.UNSET);
  });

  it("should record the response status code on the span", async () => {
    // given
    fetchMock.mockResolvedValue({ status: 201 });

    // when
    await fetchWithIdToken(AUDIENCE, BACKEND_URL, { method: "POST" });

    // then
    expect(finishedSpan().attributes["http.response.status_code"]).toBe(201);
  });

  it("should mark the span as error when fetch throws", async () => {
    // given
    fetchMock.mockRejectedValue(new Error("connection refused"));

    // when
    await expect(fetchWithIdToken(AUDIENCE, BACKEND_URL)).rejects.toThrow("connection refused");

    // then
    expect(finishedSpan().status.code).toBe(SpanStatusCode.ERROR);
  });
});
