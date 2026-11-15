package controller_test

import (
	"testing"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

func Test_ServerConfig_HTTPServerTimeouts_shouldConvertSecondsToDurations_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	cfg := controller.ServerConfig{
		ReadHeaderTimeoutSec: 2,
		ReadTimeoutSec:       3,
		WriteTimeoutSec:      5,
		IdleTimeoutSec:       13,
		Shutdown:             controller.ShutdownConfig{GracePeriodSec: 11, ShutdownTimeSec: 7},
	}

	// when
	got := cfg.HTTPServerTimeouts()

	// then
	want := controller.HTTPServerTimeouts{
		ReadHeader: 2 * time.Second,
		Read:       3 * time.Second,
		Write:      5 * time.Second,
		Idle:       13 * time.Second,
		Shutdown:   7 * time.Second,
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}
