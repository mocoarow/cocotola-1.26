package controller_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

const serveFailureAttempts = 20

// cancelOnFirstLog returns a logger that cancels ctx when the first record is written,
// so the cancellation happens after Serve has already failed.
func cancelOnFirstLog(t *testing.T, cancel context.CancelFunc) *slog.Logger {
	t.Helper()

	pr, pw := io.Pipe()
	t.Cleanup(func() {
		if err := pw.Close(); err != nil {
			t.Logf("close pipe writer: %v", err)
		}
	})

	go func() {
		buf := make([]byte, 1)
		if _, err := pr.Read(buf); err == nil {
			cancel()
		}
		if _, err := io.Copy(io.Discard, pr); err != nil {
			t.Logf("drain pipe: %v", err)
		}
	}()

	return slog.New(slog.NewTextHandler(pw, nil))
}

func Test_ServeUntilDone_shouldReturnServeError_whenContextIsCanceledAfterServeFails(t *testing.T) {
	t.Parallel()

	// given
	attempt := func() error {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		l, _ := listenTCP(t, net.JoinHostPort(loopbackHost, "0"))
		closeListener(t, l)
		server := &http.Server{Handler: http.NewServeMux(), ReadHeaderTimeout: time.Second}

		return controller.ServeUntilDone(ctx, server, l, cancelOnFirstLog(t, cancel), time.Second)
	}

	// when
	// An implementation that races the serve error against ctx.Done passes a single attempt by chance.
	errs := make([]error, 0, serveFailureAttempts)
	for range serveFailureAttempts {
		errs = append(errs, attempt())
	}

	// then
	for _, err := range errs {
		if !errors.Is(err, net.ErrClosed) {
			t.Fatalf("expected net.ErrClosed, got %v", err)
		}
	}
}
