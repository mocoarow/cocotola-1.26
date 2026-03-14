package main_test

import (
	"testing"

	main "github.com/mocoarow/cocotola-1.26/hello-world"
)

func Test_Message_shouldReturnHelloWorld(t *testing.T) {
	t.Parallel()

	// given
	expected := "Hello World"

	// when
	got := main.Message()

	// then
	if got != expected {
		t.Errorf("Message() = %q, want %q", got, expected)
	}
}
