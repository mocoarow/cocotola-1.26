package controller_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

func listenTCP(t *testing.T, address string) (net.Listener, int) {
	t.Helper()

	lc := net.ListenConfig{}
	l, err := lc.Listen(t.Context(), "tcp", address)
	if err != nil {
		t.Fatalf("listen %s: %v", address, err)
	}

	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		closeListener(t, l)
		t.Fatal("expected *net.TCPAddr")
	}

	return l, addr.Port
}

func closeListener(t *testing.T, l net.Listener) {
	t.Helper()

	if err := l.Close(); err != nil {
		t.Logf("close listener: %v", err)
	}
}

func freePort(t *testing.T) int {
	t.Helper()

	l, port := listenTCP(t, net.JoinHostPort(loopbackHost, "0"))
	closeListener(t, l)

	return port
}

const (
	loopbackHost       = "127.0.0.1"
	serverStartTimeout = 5 * time.Second
	serverPollInterval = 5 * time.Millisecond
	shortServerTimeout = 100 * time.Millisecond
	clientWaitTimeout  = 3 * time.Second
)

var testServerTimeouts = controller.HTTPServerTimeouts{
	ReadHeader: time.Second,
	Read:       time.Second,
	Write:      time.Second,
	Idle:       time.Second,
	Shutdown:   time.Second,
}

// sendRawRequest writes request on a new connection and reads until the server closes it
// or clientWaitTimeout elapses, in which case the error wraps os.ErrDeadlineExceeded.
func sendRawRequest(t *testing.T, port int, request string) ([]byte, error) {
	t.Helper()

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(t.Context(), "tcp", net.JoinHostPort(loopbackHost, strconv.Itoa(port)))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Logf("close connection: %v", err)
		}
	}()

	if err := conn.SetDeadline(time.Now().Add(clientWaitTimeout)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	if _, err := io.WriteString(conn, request); err != nil {
		t.Fatalf("write request: %v", err)
	}

	resp, err := io.ReadAll(conn)
	if err != nil {
		return resp, fmt.Errorf("read response: %w", err)
	}

	return resp, nil
}

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
		errCh <- controller.MetricsServerProcess(ctx, port, testServerTimeouts)
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
		errCh <- controller.MetricsServerProcess(ctx, port, testServerTimeouts)
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

func Test_MetricsServerProcess_shouldCloseIdleConnection_whenIdleTimeoutElapses(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	port := freePort(t)
	timeouts := controller.HTTPServerTimeouts{
		ReadHeader: time.Minute,
		Read:       time.Minute,
		Write:      time.Minute,
		Idle:       shortServerTimeout,
		Shutdown:   time.Second,
	}
	errCh := make(chan error, 1)

	go func() {
		errCh <- controller.MetricsServerProcess(ctx, port, timeouts)
	}()

	waitForServer(t, port, errCh)

	// when
	_, err := sendRawRequest(t, port, "GET /healthcheck HTTP/1.1\r\nHost: localhost\r\n\r\n")

	// then
	if errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("expected the server to close the idle keep-alive connection within %v", clientWaitTimeout)
	}

	// cleanup
	cancel()
	<-errCh
}
