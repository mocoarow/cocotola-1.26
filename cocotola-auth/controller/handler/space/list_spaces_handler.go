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

// ListSpacesUsecase defines the use case method required by the ListSpacesHandler.
type ListSpacesUsecase interface {
	ListSpaces(ctx context.Context, input *spaceservice.ListSpacesInput) (*spaceservice.ListSpacesOutput, error)
}

// ListSpacesHandler handles the GET /space endpoint.
type ListSpacesHandler struct {
	usecase ListSpacesUsecase
	logger  *slog.Logger
}

// NewListSpacesHandler returns a new ListSpacesHandler.
func NewListSpacesHandler(usecase ListSpacesUsecase) *ListSpacesHandler {
	return &ListSpacesHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "ListSpacesHandler")),
	}
}

// ListSpaces handles GET /space.
func (h *ListSpacesHandler) ListSpaces(c *gin.Context) {
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

	input, err := spaceservice.NewListSpacesInput(userID, organizationName)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid list spaces input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.ListSpaces(ctx, input)
	if err != nil {
		handleSpaceError(ctx, h.logger, c, "list spaces", err)
		return
	}

	items := make([]api.SpaceItem, len(output.Spaces))
	for i, s := range output.Spaces {
		items[i] = api.SpaceItem{
			SpaceID:        s.SpaceID.UUID(),
			OrganizationID: s.OrganizationID.UUID(),
			OwnerID:        s.OwnerID.UUID(),
			KeyName:        s.KeyName,
			Name:           s.Name,
			SpaceType:      api.SpaceItemSpaceType(s.SpaceType),
			Deleted:        s.Deleted,
		}
	}

	c.JSON(http.StatusOK, api.ListSpacesResponse{
		Spaces: items,
	})
}
