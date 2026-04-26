package space

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// FindSpaceUsecase defines the use case method required by the FindSpaceHandler.
type FindSpaceUsecase interface {
	FindSpace(ctx context.Context, input *spaceservice.FindSpaceInput) (*spaceservice.FindSpaceOutput, error)
}

// FindSpaceHandler handles the GET /internal/auth/space/:spaceId endpoint.
type FindSpaceHandler struct {
	usecase FindSpaceUsecase
	logger  *slog.Logger
}

// NewFindSpaceHandler returns a new FindSpaceHandler.
func NewFindSpaceHandler(usecase FindSpaceUsecase) *FindSpaceHandler {
	return &FindSpaceHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "FindSpaceHandler")),
	}
}

// FindSpace handles GET /internal/auth/space/:spaceId.
func (h *FindSpaceHandler) FindSpace(c *gin.Context) {
	ctx := c.Request.Context()

	spaceIDParam := c.Param("spaceId")
	spaceID, err := domain.ParseSpaceID(spaceIDParam)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid space id", slog.String("space_id", spaceIDParam), slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_space_id", "space id must be a valid UUID"))
		return
	}

	input, err := spaceservice.NewFindSpaceInput(spaceID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid find space input", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	output, err := h.usecase.FindSpace(ctx, input)
	if err != nil {
		if errors.Is(err, domain.ErrSpaceNotFound) {
			h.logger.InfoContext(ctx, "space not found", slog.String("space_id", spaceIDParam))
			c.JSON(http.StatusNotFound, controller.NewErrorResponse("space_not_found", "space not found"))
			return
		}
		h.logger.ErrorContext(ctx, "find space", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, api.FindSpaceResponse{
		SpaceID:        output.Item.SpaceID.UUID(),
		OrganizationID: output.Item.OrganizationID.UUID(),
		OwnerID:        output.Item.OwnerID.UUID(),
		KeyName:        output.Item.KeyName,
		Name:           output.Item.Name,
		SpaceType:      api.FindSpaceResponseSpaceType(output.Item.SpaceType),
		Deleted:        output.Item.Deleted,
	})
}
