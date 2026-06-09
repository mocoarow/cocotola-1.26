// OpenTelemetry tracing bootstrap for the cocotola-web server.
//
// This module is loaded *before* the React Router server via the Node.js
// `--import` flag (see the `start` script in package.json) so that the HTTP
// server and outgoing `fetch` calls are instrumented from the very first
// request. Trace context (W3C `traceparent`) is automatically propagated to
// the backend services, so a single trace spans cocotola-web -> cocotola-*.
//
// Configuration is driven by the standard OpenTelemetry environment variables:
//
//   OTEL_TRACES_EXPORTER          otlp | console | none   (default: none)
//   OTEL_EXPORTER_OTLP_PROTOCOL   http/protobuf | grpc     (default: http/protobuf)
//   OTEL_EXPORTER_OTLP_ENDPOINT   e.g. http://localhost:4318 (http) / http://localhost:4317 (grpc)
//   OTEL_SERVICE_NAME             service.name             (default: cocotola-web)
//   OTEL_SERVICE_VERSION          service.version          (default: 0.0.0)
//   OTEL_TRACES_SAMPLER[_ARG]     standard sampler config  (default: parentbased_always_on)
//
// Tracing is disabled when OTEL_TRACES_EXPORTER is unset/"none" or when
// APP_ENV is "test", so local development and tests stay quiet by default.

import { NodeSDK } from "@opentelemetry/sdk-node";
import { defaultResource, resourceFromAttributes } from "@opentelemetry/resources";
import {
  ATTR_SERVICE_NAME,
  ATTR_SERVICE_VERSION,
} from "@opentelemetry/semantic-conventions";
import { HttpInstrumentation } from "@opentelemetry/instrumentation-http";
import { UndiciInstrumentation } from "@opentelemetry/instrumentation-undici";
import { ExpressInstrumentation } from "@opentelemetry/instrumentation-express";

const exporterType = (process.env.OTEL_TRACES_EXPORTER ?? "none").toLowerCase();
const appEnv = process.env.APP_ENV ?? "local";

async function createTraceExporter(type) {
  if (type === "console") {
    const { tracing } = await import("@opentelemetry/sdk-node");
    return new tracing.ConsoleSpanExporter();
  }

  // Default to the OTLP exporter. The endpoint and headers are read from the
  // standard OTEL_EXPORTER_OTLP_* environment variables by the exporter itself.
  const protocol = (process.env.OTEL_EXPORTER_OTLP_PROTOCOL ?? "http/protobuf").toLowerCase();
  if (protocol === "grpc") {
    const { OTLPTraceExporter } = await import("@opentelemetry/exporter-trace-otlp-grpc");
    return new OTLPTraceExporter();
  }

  const { OTLPTraceExporter } = await import("@opentelemetry/exporter-trace-otlp-proto");
  return new OTLPTraceExporter();
}

async function startTracing() {
  if (appEnv === "test" || exporterType === "none") {
    console.info(
      `[otel] tracing disabled (APP_ENV=${appEnv}, OTEL_TRACES_EXPORTER=${exporterType})`,
    );
    return;
  }

  const serviceName = process.env.OTEL_SERVICE_NAME ?? "cocotola-web";
  const resource = defaultResource().merge(
    resourceFromAttributes({
      [ATTR_SERVICE_NAME]: serviceName,
      [ATTR_SERVICE_VERSION]: process.env.OTEL_SERVICE_VERSION ?? "0.0.0",
    }),
  );

  const traceExporter = await createTraceExporter(exporterType);

  const sdk = new NodeSDK({
    resource,
    traceExporter,
    instrumentations: [
      new HttpInstrumentation(),
      new UndiciInstrumentation(),
      new ExpressInstrumentation(),
    ],
  });

  sdk.start();
  console.info(
    `[otel] tracing started: service=${serviceName}, exporter=${exporterType}`,
  );

  let shuttingDown = false;
  const shutdown = (signal) => {
    if (shuttingDown) {
      return;
    }
    shuttingDown = true;
    sdk
      .shutdown()
      .then(() => console.info(`[otel] tracing shut down (${signal})`))
      .catch((error) => console.error("[otel] failed to shut down tracing", error))
      .finally(() => process.exit(0));
  };

  process.once("SIGTERM", () => shutdown("SIGTERM"));
  process.once("SIGINT", () => shutdown("SIGINT"));
}

try {
  await startTracing();
} catch (error) {
  // Never let observability setup take the server down.
  console.error("[otel] failed to start tracing", error);
}
