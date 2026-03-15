package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/mocoarow/slogotel"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

const (
	logMaxQueueSize      = 10_000
	logExportMaxBatch    = 10_000
	logExportIntervalSec = 10
	logExportTimeoutSec  = 10
)

// OTLPLogConfig holds OTLP HTTP log exporter settings.
type OTLPLogConfig struct {
	Endpoint string `yaml:"endpoint" validate:"required"`
	Insecure bool   `yaml:"insecure"`
}

// UptraceLogConfig holds Uptrace log exporter settings.
type UptraceLogConfig struct {
	Endpoint string `yaml:"endpoint" validate:"required"`
	DSN      string `yaml:"dsn" validate:"required"`
}

// LogConfig holds log level, platform, and exporter settings.
type LogConfig struct {
	Level    string            `yaml:"level"`
	Platform string            `yaml:"platform"`
	Levels   map[string]string `yaml:"levels"`
	Exporter string            `yaml:"exporter" validate:"oneof=none otlphttp uptracehttp"`
	OTLP     *OTLPLogConfig    `yaml:"otlp"`
	Uptrace  *UptraceLogConfig `yaml:"uptrace"`
}

// InitLog sets up the global slog default based on the configured exporter.
func InitLog(ctx context.Context, logConfig *LogConfig, appName string) (func(), error) {
	if logConfig.Exporter == "none" {
		defaultLogLevel := stringToLogLevel(logConfig.Level)
		jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ //nolint:exhaustruct
			Level: defaultLogLevel,
		})
		handler := slogotel.New(jsonHandler)
		slog.SetDefault(slog.New(handler))
		return func() {
			// No-op
		}, nil
	}

	return InitLogProvider(ctx, logConfig, appName)
}

const logShutdownTimeout = 5 * time.Second

func initLogExporter(ctx context.Context, logConfig *LogConfig) (sdklog.Exporter, error) { //nolint:ireturn // returns interface required by OpenTelemetry SDK
	switch logConfig.Exporter {
	case "otlphttp":
		return initLogExporterOTLPHTTP(ctx, logConfig)
	case "uptracehttp":
		return initLogExporterUptraceHTTP(ctx, logConfig)
	default:
		return nil, fmt.Errorf("invalid log exporter: %s", logConfig.Exporter)
	}
}

// InitLogProvider creates an OpenTelemetry log provider and sets it as the global slog default.
func InitLogProvider(ctx context.Context, logConfig *LogConfig, appName string) (func(), error) {
	exp, err := initLogExporter(ctx, logConfig)
	if err != nil {
		return nil, fmt.Errorf("init log exporter: %w", err)
	}

	bp := sdklog.NewBatchProcessor(exp,
		sdklog.WithMaxQueueSize(logMaxQueueSize),
		sdklog.WithExportMaxBatchSize(logExportMaxBatch),
		sdklog.WithExportInterval(logExportIntervalSec*time.Second),
		sdklog.WithExportTimeout(logExportTimeoutSec*time.Second),
	)

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(bp),
		sdklog.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(appName),
		)),
	)

	global.SetLoggerProvider(lp)

	defaultLogLevel := stringToLogLevel(logConfig.Level)
	otelHandler := otelslog.NewHandler(appName, otelslog.WithLoggerProvider(lp))
	filteredHandler := &levelFilterHandler{
		handler:  otelHandler,
		minLevel: defaultLogLevel,
	}

	slog.SetDefault(slog.New(filteredHandler))

	return func() {
		shutdownBaseCtx := context.Background()
		if ctx != nil {
			shutdownBaseCtx = context.WithoutCancel(ctx)
		}
		shutdownCtx, cancel := context.WithTimeout(shutdownBaseCtx, logShutdownTimeout)
		defer cancel()

		if err := lp.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown log provider", slog.String("error", err.Error()))
		}
	}, nil
}

type levelFilterHandler struct {
	handler  slog.Handler
	minLevel slog.Level
}

func (h *levelFilterHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

// Handle processes the record if the level is enabled.
// The Enabled check is intentionally kept as a defensive guard because this handler
// may be wrapped by WithAttrs/WithGroup, where Handle could be called directly.
func (h *levelFilterHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}
	return h.handler.Handle(ctx, record) //nolint:wrapcheck
}

func (h *levelFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelFilterHandler{
		handler:  h.handler.WithAttrs(attrs),
		minLevel: h.minLevel,
	}
}

func (h *levelFilterHandler) WithGroup(name string) slog.Handler {
	return &levelFilterHandler{
		handler:  h.handler.WithGroup(name),
		minLevel: h.minLevel,
	}
}

func stringToLogLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		slog.Warn("unsupported log level, falling back to warn", slog.String("level", value))

		return slog.LevelWarn
	}
}
