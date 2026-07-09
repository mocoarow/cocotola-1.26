//go:build small

package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/config"
)

func Test_buildSeeder_shouldReturnError_whenBaseURLIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	qcfg := config.QuestionClientConfig{BaseURL: ""}
	csvCfg := config.CSVSeedConfig{BucketName: ""}

	// when
	seeder, err := buildSeeder(ctx, "local", qcfg, csvCfg, nil)

	// then
	require.ErrorIs(t, err, ErrQuestionBaseURLRequired)
	assert.Nil(t, seeder)
}

func Test_buildSeeder_shouldApplyDefaultTimeout_whenTimeoutSecIsZero(t *testing.T) {
	// Mutates package-level newHTTPClientFn — must not be parallel.

	// given
	var capturedTimeout time.Duration
	orig := newHTTPClientFn
	newHTTPClientFn = func(ctx context.Context, appEnv, audience string, timeout time.Duration) (*http.Client, error) {
		capturedTimeout = timeout
		return orig(ctx, appEnv, audience, timeout)
	}
	defer func() { newHTTPClientFn = orig }()

	ctx := context.Background()
	qcfg := config.QuestionClientConfig{BaseURL: "http://localhost", TimeoutSec: 0}
	csvCfg := config.CSVSeedConfig{BucketName: ""}

	// when
	seeder, err := buildSeeder(ctx, "local", qcfg, csvCfg, nil)

	// then
	require.NoError(t, err)
	require.NotNil(t, seeder)
	assert.Equal(t, 10*time.Second, capturedTimeout)
}

func Test_loadCSVSeeds_shouldReturnEmptySlice_whenBucketNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	csvCfg := config.CSVSeedConfig{BucketName: ""}

	// when
	seeds, err := loadCSVSeeds(ctx, csvCfg)

	// then
	require.NoError(t, err)
	assert.NotNil(t, seeds)
	assert.Empty(t, seeds)
}
