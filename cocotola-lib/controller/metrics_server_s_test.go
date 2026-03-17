package controller_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

func freePort(t *testing.T) int {
	t.Helper()

	lc := net.ListenConfig{}
	l, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("get free port: %v", err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			t.Logf("close listener: %v", err)
		}
	}()

	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatal("expected *net.TCPAddr")
	}

	return addr.Port
}

func Test_MetricsServerProcess_shouldServeHealthcheck_whenStarted(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := freePort(t)
	errCh := make(chan error, 1)

	go func() {
		errCh <- controller.MetricsServerProcess(ctx, port, time.Second, time.Second)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// when
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthcheck", port)) //nolint:noctx

	// then
	if err != nil {
		t.Fatalf("healthcheck request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// cleanup
	cancel()
	<-errCh
}

func Test_MetricsServerProcess_shouldShutdownGracefully_whenContextCanceled(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	port := freePort(t)
	errCh := make(chan error, 1)

	go func() {
		errCh <- controller.MetricsServerProcess(ctx, port, time.Second, time.Second)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// when
	cancel()
	err := <-errCh

	// then
	if err != nil {
		t.Fatalf("expected nil error after graceful shutdown, got %v", err)
	}
}
