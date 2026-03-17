package gateway_test

import (
	"testing"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

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
