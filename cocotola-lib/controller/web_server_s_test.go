package controller_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"syscall"
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
		errCh <- controller.WebServerProcess(ctx, mux, port, testServerTimeouts)
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
		errCh <- controller.WebServerProcess(ctx, mux, port, testServerTimeouts)
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

func Test_WebServerProcess_shouldReturnError_whenPortInUse(t *testing.T) {
	t.Parallel()

	// given
	l, port := listenTCP(t, ":0")
	defer closeListener(t, l)

	ctx, cancel := context.WithTimeout(t.Context(), serverStartTimeout)
	defer cancel()

	// when
	err := controller.WebServerProcess(ctx, http.NewServeMux(), port, testServerTimeouts)

	// then
	if !errors.Is(err, syscall.EADDRINUSE) {
		t.Fatalf("expected EADDRINUSE, got %v", err)
	}
}

func Test_WebServerProcess_shouldCloseConnection_whenRequestBodyIsNotSentWithinReadTimeout(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	port := freePort(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	timeouts := controller.HTTPServerTimeouts{
		ReadHeader: time.Minute,
		Read:       shortServerTimeout,
		Write:      time.Minute,
		Idle:       time.Minute,
		Shutdown:   time.Second,
	}
	errCh := make(chan error, 1)

	go func() {
		errCh <- controller.WebServerProcess(ctx, mux, port, timeouts)
	}()

	waitForServer(t, port, errCh)

	// when
	_, err := sendRawRequest(t, port,
		"POST /upload HTTP/1.1\r\nHost: localhost\r\nContent-Length: 10\r\nConnection: close\r\n\r\nabc")

	// then
	if errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("expected the server to close the connection within %v", clientWaitTimeout)
	}

	// cleanup
	cancel()
	<-errCh
}

func Test_WebServerProcess_shouldNotDeliverResponse_whenHandlerWritesAfterWriteTimeout(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	port := freePort(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(3 * shortServerTimeout):
		case <-r.Context().Done():
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	timeouts := controller.HTTPServerTimeouts{
		ReadHeader: time.Minute,
		Read:       time.Minute,
		Write:      shortServerTimeout,
		Idle:       time.Minute,
		Shutdown:   time.Second,
	}
	errCh := make(chan error, 1)

	go func() {
		errCh <- controller.WebServerProcess(ctx, mux, port, timeouts)
	}()

	waitForServer(t, port, errCh)

	// when
	resp, err := sendRawRequest(t, port, "GET /slow HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")

	// then
	if len(resp) != 0 {
		t.Fatalf("expected no response after the write timeout, got %q (err: %v)", resp, err)
	}

	// cleanup
	cancel()
	<-errCh
}

const canceledStartAttempts = 20

func Test_WebServerProcess_shouldReturnError_whenPortInUseAndContextAlreadyCanceled(t *testing.T) {
	t.Parallel()

	// given
	l, port := listenTCP(t, ":0")
	defer closeListener(t, l)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	// when
	// An implementation that races the startup error against ctx.Done passes a single attempt by chance.
	errs := make([]error, 0, canceledStartAttempts)
	for range canceledStartAttempts {
		errs = append(errs, controller.WebServerProcess(ctx, http.NewServeMux(), port, testServerTimeouts))
	}

	// then
	for _, err := range errs {
		if !errors.Is(err, syscall.EADDRINUSE) {
			t.Fatalf("expected EADDRINUSE, got %v", err)
		}
	}
}
