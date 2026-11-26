package gateway_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"

	libdomain "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

const gcpProjectIDKey = attribute.Key("gcp.project_id")

func Test_initTracerExporter_shouldReturnErrGoogleTraceProjectIDRequired_whenGoogleProjectIDIsMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		google *gateway.GoogleTraceConfig
	}{
		{"google_nil", nil},
		{"project_id_empty", &gateway.GoogleTraceConfig{ProjectID: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			traceConfig := gateway.TraceConfig{
				Exporter:           "google",
				SamplingPercentage: 100,
				OTLP:               nil,
				Google:             tt.google,
				Uptrace:            nil,
			}

			// when
			_, err := gateway.InitTracerExporter(t.Context(), traceConfig)

			// then
			require.ErrorIs(t, err, gateway.ErrGoogleTraceProjectIDRequired)
		})
	}
}

func Test_traceResource_shouldSetGCPProjectID_whenExporterIsGoogle(t *testing.T) {
	t.Parallel()

	// given
	traceConfig := gateway.TraceConfig{
		Exporter:           "google",
		SamplingPercentage: 100,
		OTLP:               nil,
		Google:             &gateway.GoogleTraceConfig{ProjectID: "test-project"},
		Uptrace:            nil,
	}

	// when
	res := gateway.TraceResource(traceConfig, "test-app")

	// then
	projectID, ok := res.Set().Value(gcpProjectIDKey)
	require.True(t, ok)
	assert.Equal(t, "test-project", projectID.AsString())
}

func Test_traceResource_shouldSetServiceName_whenExporterIsGoogle(t *testing.T) {
	t.Parallel()

	// given
	traceConfig := gateway.TraceConfig{
		Exporter:           "google",
		SamplingPercentage: 100,
		OTLP:               nil,
		Google:             &gateway.GoogleTraceConfig{ProjectID: "test-project"},
		Uptrace:            nil,
	}

	// when
	res := gateway.TraceResource(traceConfig, "test-app")

	// then
	serviceName, ok := res.Set().Value(semconv.ServiceNameKey)
	require.True(t, ok)
	assert.Equal(t, "test-app", serviceName.AsString())
}

func Test_traceResource_shouldNotSetGCPProjectID_whenExporterIsNotGoogle(t *testing.T) {
	t.Parallel()

	// given
	traceConfig := gateway.TraceConfig{
		Exporter:           "none",
		SamplingPercentage: 100,
		OTLP:               nil,
		Google:             &gateway.GoogleTraceConfig{ProjectID: "test-project"},
		Uptrace:            nil,
	}

	// when
	res := gateway.TraceResource(traceConfig, "test-app")

	// then
	_, ok := res.Set().Value(gcpProjectIDKey)
	assert.False(t, ok)
}

func Test_TraceConfig_shouldPassValidation_whenExporterIsGoogle(t *testing.T) {
	t.Parallel()

	// given
	traceConfig := gateway.TraceConfig{
		Exporter:           "google",
		SamplingPercentage: 100,
		OTLP:               nil,
		Google:             &gateway.GoogleTraceConfig{ProjectID: "test-project"},
		Uptrace:            nil,
	}

	// when
	err := libdomain.ValidateStruct(traceConfig)

	// then
	require.NoError(t, err)
}

func Test_traceBatchSpanProcessorOptions_shouldConfigureQueueBatchIntervalAndTimeout(t *testing.T) {
	t.Parallel()

	// given
	var opts sdktrace.BatchSpanProcessorOptions

	// when
	for _, apply := range gateway.TraceBatchSpanProcessorOptions() {
		apply(&opts)
	}

	// then
	expected := sdktrace.BatchSpanProcessorOptions{
		MaxQueueSize:       10_000,
		BatchTimeout:       10 * time.Second,
		ExportTimeout:      10 * time.Second,
		MaxExportBatchSize: 10_000,
	}
	assert.Equal(t, expected, opts)
}

func Test_initTraceSampler_shouldReturnExpectedSampler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		percentage   int
		expectedDesc string
	}{
		{"always sample when 100", 100, sdktrace.AlwaysSample().Description()},
		{"always sample when over 100", 150, sdktrace.AlwaysSample().Description()},
		{"never sample when 0", 0, sdktrace.NeverSample().Description()},
		{"never sample when negative", -10, sdktrace.NeverSample().Description()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			sampler := gateway.InitTraceSampler(tt.percentage)

			// then
			if sampler.Description() != tt.expectedDesc {
				t.Fatalf("expected %s, got %s", tt.expectedDesc, sampler.Description())
			}
		})
	}
}

func Test_initTraceSampler_shouldReturnParentBasedSampler_when50(t *testing.T) {
	t.Parallel()

	// given
	percentage := 50

	// when
	sampler := gateway.InitTraceSampler(percentage)

	// then
	desc := sampler.Description()
	if desc == sdktrace.AlwaysSample().Description() || desc == sdktrace.NeverSample().Description() {
		t.Fatalf("expected ParentBased sampler, got %s", desc)
	}
}
