package config_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/audio"
)

func Test_AudioGeneratorConfig_MaxPerRunValidatorTagMatchesMaxBatchLimit(t *testing.T) {
	t.Parallel()

	// given
	field, ok := reflect.TypeFor[config.AudioGeneratorConfig]().FieldByName("MaxPerRun")
	require.True(t, ok, "AudioGeneratorConfig.MaxPerRun field must exist")
	tag := field.Tag.Get("validate")

	// when
	expected := fmt.Sprintf("required,gte=1,lte=%d", audio.MaxBatchLimit)

	// then
	assert.Equal(t, expected, tag, "validate tag must mirror audio.MaxBatchLimit")
}
