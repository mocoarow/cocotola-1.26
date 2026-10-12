package controller_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

func freePort(t *testing.T) int {
	t.Helper()

	lc := net.ListenConfig{}
	l, err := lc.Listen(context.Background(), "tcp", net.JoinHostPort(loopbackHost, "0"))
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

const (
	loopbackHost       = "127.0.0.1"
	serverStartTimeout = 5 * time.Second
	serverPollInterval = 5 * time.Millisecond
)

func waitForServer(t *testing.T, port int, errCh <-chan error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), serverStartTimeout)
	defer cancel()

	addr := net.JoinHostPort(loopbackHost, strconv.Itoa(port))
	dialer := net.Dialer{}
	ticker := time.NewTicker(serverPollInterval)
	defer ticker.Stop()

	for {
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			if err := conn.Close(); err != nil {
				t.Logf("close connection: %v", err)
			}

			return
		}

		select {
		case err := <-errCh:
			t.Fatalf("server exited before accepting connections: %v", err)
		case <-ctx.Done():
			t.Fatalf("server on %s did not accept connections within %v", addr, serverStartTimeout)
		case <-ticker.C:
		}
	}
}

func getStatusCode(t *testing.T, port int, path string) int {
	t.Helper()

	url := fmt.Sprintf("http://%s%s", net.JoinHostPort(loopbackHost, strconv.Itoa(port)), path)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Logf("close response body: %v", err)
	}

	return resp.StatusCode
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

	waitForServer(t, port, errCh)

	// when
	statusCode := getStatusCode(t, port, "/healthcheck")

	// then
	if statusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", statusCode)
	}

	// cleanup
	cancel()
	<-errCh
}

func Test_MetricsServerProcess_shouldShutdownGracefully_whenContextCanceled(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := freePort(t)
	errCh := make(chan error, 1)

	go func() {
		errCh <- controller.MetricsServerProcess(ctx, port, time.Second, time.Second)
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
