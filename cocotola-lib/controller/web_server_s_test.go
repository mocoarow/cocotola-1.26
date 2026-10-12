package controller_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

func Test_WebServerProcess_shouldServeRequests_whenStarted(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := freePort(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	errCh := make(chan error, 1)

	go func() {
		errCh <- controller.WebServerProcess(ctx, mux, port, time.Second, time.Second)
	}()

	waitForServer(t, port, errCh)

	// when
	statusCode := getStatusCode(t, port, "/ping")

	// then
	if statusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", statusCode)
	}

	// cleanup
	cancel()
	<-errCh
}

func Test_WebServerProcess_shouldShutdownGracefully_whenContextCanceled(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := freePort(t)
	mux := http.NewServeMux()

	errCh := make(chan error, 1)

	go func() {
		errCh <- controller.WebServerProcess(ctx, mux, port, time.Second, time.Second)
	}()

	waitForServer(t, port, errCh)

	// when
	cancel()
	err := <-errCh

	// then
	if err != nil {
		t.Fatalf("expected nil error after graceful shutdown, got %v", err)
	}
}
