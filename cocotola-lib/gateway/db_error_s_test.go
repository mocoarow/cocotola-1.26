package gateway_test

import (
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

func newIncorrectIntegerMySQLError() *mysql.MySQLError {
	return &mysql.MySQLError{
		Number:   1366,
		SQLState: [5]byte{'H', 'Y', '0', '0', '0'},
		Message:  "Incorrect integer value: '" + testLoginID + "' for column 'id' at row 1",
	}
}

func Test_redactDriverError_shouldReturnMessageWithoutDriverMessage_whenErrorIsDriverError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "postgres", err: newInvalidUUIDPgError(), want: "postgres error (SQLSTATE 22P02)"},
		{name: "mysql", err: newIncorrectIntegerMySQLError(), want: "mysql error 1366"},
		{name: "wrapped postgres", err: fmt.Errorf("find: %w", newInvalidUUIDPgError()), want: "postgres error (SQLSTATE 22P02)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			err := tt.err

			// when
			redacted := gateway.RedactDriverError(err)

			// then
			assert.Equal(t, tt.want, redacted.Error())
		})
	}
}

func Test_redactDriverError_shouldKeepOriginalErrorInChain_whenErrorIsDriverError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{name: "postgres", err: newInvalidUUIDPgError()},
		{name: "mysql", err: newIncorrectIntegerMySQLError()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			err := tt.err

			// when
			redacted := gateway.RedactDriverError(err)

			// then
			assert.ErrorIs(t, redacted, err)
		})
	}
}

func Test_redactDriverError_shouldReturnSameError_whenErrorIsNotDriverError(t *testing.T) {
	t.Parallel()

	// given
	err := gorm.ErrDuplicatedKey

	// when
	redacted := gateway.RedactDriverError(err)

	// then
	assert.Equal(t, err, redacted)
}
