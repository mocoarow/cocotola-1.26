package question_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

func Test_FailAudioInput_shouldMirrorMaxAudioLastErrorLength_whenReasonTagIsRead(t *testing.T) {
	t.Parallel()

	// given
	field, ok := reflect.TypeFor[question.FailAudioInput]().FieldByName("Reason")
	require.True(t, ok)

	// when
	tag := field.Tag.Get("validate")

	// then
	assert.Equal(t, fmt.Sprintf("max=%d", domainquestion.MaxAudioLastErrorLength), tag)
}
