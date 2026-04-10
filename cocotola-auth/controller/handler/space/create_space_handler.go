package space

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// CreateSpaceUsecase defines the use case method required by the CreateSpaceHandler.
type CreateSpaceUsecase interface {
	CreateSpace(ctx context.Context, input *spaceservice.CreateSpaceInput) (*spaceservice.CreateSpaceOutput, error)
}

// CreateSpaceHandler handles the POST /space endpoint.
type CreateSpaceHandler struct {
	usecase CreateSpaceUsecase
	logger  *slog.Logger
}

// NewCreateSpaceHandler returns a new CreateSpaceHandler.
func NewCreateSpaceHandler(usecase CreateSpaceUsecase) *CreateSpaceHandler {
	return &CreateSpaceHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "CreateSpaceHandler")),
	}
}

// CreateSpace handles POST /space.
func (h *CreateSpaceHandler) CreateSpace(c *gin.Context) {
	ctx := c.Request.Context()

	userID, ok := handler.GetAppUserIDFromContext(c)
	if !ok {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	organizationName := c.GetString(controller.ContextFieldOrganizationName{})
	if organizationName == "" {
		h.logger.WarnContext(ctx, "unauthorized: missing organization name")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	var req api.CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid create space request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))
		return
	}

	input, err := spaceservice.NewCreateSpaceInput(userID, organizationName, req.Name, string(req.SpaceType))
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid create space input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.CreateSpace(ctx, input)
	if err != nil {
		handleSpaceError(ctx, h.logger, c, "create space", err)
		return
	}

	c.JSON(http.StatusCreated, api.CreateSpaceResponse{
		SpaceID:        output.SpaceID.UUID(),
		OrganizationID: output.OrganizationID.UUID(),
		OwnerID:        output.OwnerID.UUID(),
		KeyName:        output.KeyName,
		Name:           output.Name,
		SpaceType:      api.CreateSpaceResponseSpaceType(output.SpaceType),
		Deleted:        output.Deleted,
	})
}
