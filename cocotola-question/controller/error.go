package controller

import (
	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
)

// NewErrorResponse creates an ErrorResponse with the given error code and message.
func NewErrorResponse(code string, message string) *api.ErrorResponse {
	return &api.ErrorResponse{
		Code:    code,
		Message: message,
	}
}
