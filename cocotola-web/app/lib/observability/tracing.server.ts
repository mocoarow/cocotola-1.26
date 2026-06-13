import { type Attributes, type Span, SpanStatusCode, trace } from "@opentelemetry/api";

/**
 * Tracer used for application-level (manual) spans.
 *
 * The HTTP server and outgoing `fetch` calls are instrumented automatically by
 * the OpenTelemetry Node SDK (see `instrumentation.mjs`). Use {@link withSpan}
 * to add richer, business-level spans around server-side operations such as
 * backend API calls.
 */
export const tracer = trace.getTracer("cocotola-web");

/**
 * Runs `fn` inside an active span named `name`, recording errors and ending the
 * span automatically. The span is a no-op when tracing is disabled, so this is
 * safe to call unconditionally from server code (loaders, actions, API helpers).
 */
export async function withSpan<T>(
  name: string,
  fn: (span: Span) => Promise<T>,
  attributes?: Attributes,
): Promise<T> {
  return tracer.startActiveSpan(name, async (span) => {
    if (attributes) {
      span.setAttributes(attributes);
    }
    try {
      const result = await fn(span);
      span.setStatus({ code: SpanStatusCode.OK });
      return result;
    } catch (error) {
      span.setStatus({
        code: SpanStatusCode.ERROR,
        message: error instanceof Error ? error.message : String(error),
      });
      if (error instanceof Error) {
        span.recordException(error);
      }
      throw error;
    } finally {
      span.end();
    }
  });
}
