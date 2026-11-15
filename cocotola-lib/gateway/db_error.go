package gateway

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type redactedDriverError struct {
	summary string
	err     error
}

func (e *redactedDriverError) Error() string {
	return e.summary
}

func (e *redactedDriverError) Unwrap() error {
	return e.err
}

// redactDriverError replaces the message of a driver error, which can contain bound values,
// with its error code. The original error stays reachable through errors.Is and errors.As.
func redactDriverError(err error) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return &redactedDriverError{summary: "postgres error (SQLSTATE " + pgErr.Code + ")", err: err}
	}
	if mysqlErr, ok := errors.AsType[*mysql.MySQLError](err); ok {
		return &redactedDriverError{summary: fmt.Sprintf("mysql error %d", mysqlErr.Number), err: err}
	}
	return err
}

type gormCallbackRegisterer interface {
	Register(name string, fn func(*gorm.DB)) error
}

func registerDriverErrorRedaction(db *gorm.DB) error {
	cb := db.Callback()
	hooks := []struct {
		operation  string
		registerer gormCallbackRegisterer
	}{
		{operation: "create_before_tx_end", registerer: cb.Create().Before("gorm:commit_or_rollback_transaction")},
		{operation: "create", registerer: cb.Create().Before("otel:after:create")},
		{operation: "select", registerer: cb.Query().Before("otel:after:select")},
		{operation: "update_before_tx_end", registerer: cb.Update().Before("gorm:commit_or_rollback_transaction")},
		{operation: "update", registerer: cb.Update().Before("otel:after:update")},
		{operation: "delete_before_tx_end", registerer: cb.Delete().Before("gorm:commit_or_rollback_transaction")},
		{operation: "delete", registerer: cb.Delete().Before("otel:after:delete")},
		{operation: "row", registerer: cb.Row().Before("otel:after:row")},
		{operation: "raw", registerer: cb.Raw().Before("otel:after:raw")},
	}
	for _, h := range hooks {
		if err := h.registerer.Register("cocotola:redact_driver_error:"+h.operation, redactTxError); err != nil {
			return fmt.Errorf("%s: %w", h.operation, err)
		}
	}
	return nil
}

func redactTxError(tx *gorm.DB) {
	if tx.Error != nil {
		tx.Error = redactDriverError(tx.Error)
	}
}
