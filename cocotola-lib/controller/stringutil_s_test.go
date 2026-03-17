package controller_test

import (
	"reflect"
	"testing"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

func TestSplitCommaSeparated_shouldReturnExpectedSlice_whenInputVaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "comma separated",
			input: "a,b",
			want:  []string{"a", "b"},
		},
		{
			name:  "spaces are trimmed",
			input: " a , b ",
			want:  []string{"a", "b"},
		},
		{
			name:  "single value",
			input: "only",
			want:  []string{"only"},
		},
		{
			name:  "wildcard remains single",
			input: "*",
			want:  []string{"*"},
		},
		{
			name:  "trailing comma ignores empty",
			input: "a,",
			want:  []string{"a"},
		},
		{
			name:  "empty input returns empty slice",
			input: "",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := controller.SplitCommaSeparated(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("SplitCommaSeparated(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}
