package workbook_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

func Test_WorkbookInput_shouldMirrorDomainVisibilities_whenVisibilityTagIsRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input reflect.Type
	}{
		{name: "CreateWorkbookInput", input: reflect.TypeFor[workbook.CreateWorkbookInput]()},
		{name: "UpdateWorkbookInput", input: reflect.TypeFor[workbook.UpdateWorkbookInput]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			field, ok := tt.input.FieldByName("Visibility")
			require.True(t, ok)

			// when
			tag := field.Tag.Get("validate")

			// then
			expected := fmt.Sprintf("required,oneof=%s %s", domainworkbook.VisibilityPrivate().Value(), domainworkbook.VisibilityPublic().Value())
			assert.Equal(t, expected, tag)
		})
	}
}

func Test_WorkbookInput_shouldMirrorDomainMaxSeedKeyLength_whenSeedKeyTagIsRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input reflect.Type
	}{
		{name: "CreateWorkbookInput", input: reflect.TypeFor[workbook.CreateWorkbookInput]()},
		{name: "UpdateWorkbookInput", input: reflect.TypeFor[workbook.UpdateWorkbookInput]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			field, ok := tt.input.FieldByName("SeedKey")
			require.True(t, ok)

			// when
			tag := field.Tag.Get("validate")

			// then
			assert.Equal(t, fmt.Sprintf("omitempty,max=%d", domainworkbook.MaxSeedKeyLength), tag)
		})
	}
}
