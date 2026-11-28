package gateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/gateway"
)

const (
	testBucket      = "test-bucket"
	testContentType = "audio/ogg"
)

var defaultLoggerMu sync.Mutex

func lockDefaultLogger(t *testing.T) {
	t.Helper()

	defaultLoggerMu.Lock()
	prev := slog.Default()
	prevOut, prevFlags := log.Writer(), log.Flags()
	t.Cleanup(func() {
		slog.SetDefault(prev)
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
		defaultLoggerMu.Unlock()
	})
}

type uploadedObject struct {
	Name         string `json:"name"`
	ContentType  string `json:"contentType"`
	CacheControl string `json:"cacheControl"`
}

func readUploadedObject(r *http.Request) (uploadedObject, error) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return uploadedObject{}, fmt.Errorf("parse content type: %w", err)
	}
	if mediaType != "multipart/related" {
		return uploadedObject{}, fmt.Errorf("unexpected media type %s", mediaType)
	}

	part, err := multipart.NewReader(r.Body, params["boundary"]).NextPart()
	if err != nil {
		return uploadedObject{}, fmt.Errorf("read metadata part: %w", err)
	}

	var obj uploadedObject
	if err := json.NewDecoder(part).Decode(&obj); err != nil {
		return uploadedObject{}, fmt.Errorf("decode metadata: %w", err)
	}
	if _, err := io.Copy(io.Discard, r.Body); err != nil {
		return uploadedObject{}, fmt.Errorf("drain body: %w", err)
	}

	return obj, nil
}

// The returned handler runs on the server goroutine, so it must use t.Errorf, not t.Fatal.
func recordUpload(t *testing.T, uploads chan<- uploadedObject) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		obj, err := readUploadedObject(r)
		if err != nil {
			t.Errorf("read uploaded object: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)

			return
		}

		uploads <- obj
		writeJSON(t, w, obj)
	})
}

func forbidden() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	})
}

func noRequestExpected(t *testing.T) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Errorf("unexpected request to storage server")
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	})
}

func newTestGCSClient(t *testing.T, handler http.Handler) *gateway.GCSClient {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := storage.NewClient(context.Background(),
		option.WithEndpoint(srv.URL),
		option.WithoutAuthentication())
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close storage client: %v", err)
		}
	})

	return gateway.NewGCSClientForTest(client, testBucket)
}

func receiveUpload(t *testing.T, uploads <-chan uploadedObject) uploadedObject {
	t.Helper()

	select {
	case obj := <-uploads:
		return obj
	default:
		t.Fatal("no upload received")

		return uploadedObject{}
	}
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}

func hasLogRecord(t *testing.T, buf *bytes.Buffer, msg, objectPath string) bool {
	t.Helper()

	dec := json.NewDecoder(buf)
	for dec.More() {
		var rec struct {
			Msg        string `json:"msg"`
			ObjectPath string `json:"objectPath"`
		}
		require.NoError(t, dec.Decode(&rec))
		if rec.Msg == msg && rec.ObjectPath == objectPath {
			return true
		}
	}

	return false
}

func Test_GCSClient_Upload_shouldReturnWrittenBytes_whenUploadSucceeds(t *testing.T) {
	t.Parallel()

	// given
	uploads := make(chan uploadedObject, 1)
	client := newTestGCSClient(t, recordUpload(t, uploads))
	data := []byte("opus-audio-bytes")

	// when
	n, err := client.Upload(context.Background(), "audio/written-bytes.opus", testContentType, data)

	// then
	require.NoError(t, err)
	assert.Equal(t, int64(len(data)), n)
}

func Test_GCSClient_Upload_shouldWriteToObjectPath_whenUploadSucceeds(t *testing.T) {
	t.Parallel()

	// given
	uploads := make(chan uploadedObject, 1)
	client := newTestGCSClient(t, recordUpload(t, uploads))
	objectPath := "audio/wb-1/q-1/source.opus"

	// when
	_, err := client.Upload(context.Background(), objectPath, testContentType, []byte("opus-audio-bytes"))

	// then
	require.NoError(t, err)
	assert.Equal(t, objectPath, receiveUpload(t, uploads).Name)
}

func Test_GCSClient_Upload_shouldSetContentType_whenUploadSucceeds(t *testing.T) {
	t.Parallel()

	// given
	uploads := make(chan uploadedObject, 1)
	client := newTestGCSClient(t, recordUpload(t, uploads))

	// when
	_, err := client.Upload(context.Background(), "audio/content-type.opus", testContentType, []byte("opus-audio-bytes"))

	// then
	require.NoError(t, err)
	assert.Equal(t, testContentType, receiveUpload(t, uploads).ContentType)
}

func Test_GCSClient_Upload_shouldSetImmutableCacheControl_whenUploadSucceeds(t *testing.T) {
	t.Parallel()

	// given
	uploads := make(chan uploadedObject, 1)
	client := newTestGCSClient(t, recordUpload(t, uploads))

	// when
	_, err := client.Upload(context.Background(), "audio/cache-control.opus", testContentType, []byte("opus-audio-bytes"))

	// then
	require.NoError(t, err)
	assert.Equal(t, "public, max-age=31536000, immutable", receiveUpload(t, uploads).CacheControl)
}

func Test_GCSClient_Upload_shouldReturnContextCanceled_whenCopyFails(t *testing.T) {
	t.Parallel()

	// given
	lockDefaultLogger(t)
	client := newTestGCSClient(t, noRequestExpected(t))

	// when
	_, err := client.Upload(canceledContext(), "audio/copy-fails.opus", testContentType, []byte("opus-audio-bytes"))

	// then
	require.ErrorIs(t, err, context.Canceled)
}

func Test_GCSClient_Upload_shouldReturnServerError_whenCloseFails(t *testing.T) {
	t.Parallel()

	// given
	client := newTestGCSClient(t, forbidden())

	// when
	_, err := client.Upload(context.Background(), "audio/close-fails.opus", testContentType, []byte("opus-audio-bytes"))

	// then
	var apiErr *googleapi.Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusForbidden, apiErr.Code)
}

func Test_GCSClient_Upload_shouldLogCloseFailure_whenCopyAndCloseFail(t *testing.T) {
	t.Parallel()

	// given
	lockDefaultLogger(t)
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	client := newTestGCSClient(t, noRequestExpected(t))
	objectPath := "audio/copy-and-close-fail.opus"

	// when
	_, _ = client.Upload(canceledContext(), objectPath, testContentType, []byte("opus-audio-bytes"))

	// then
	assert.True(t, hasLogRecord(t, &buf, "close object writer after copy failure", objectPath))
}
