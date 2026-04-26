package middleware

import (
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("github.com/mocoarow/cocotola-1.26/cocotola-question/controller/middleware") //nolint:gochecknoglobals
)
