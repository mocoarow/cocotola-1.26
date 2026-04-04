package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// getMeResponse represents the response from cocotola-auth GET /api/v1/auth/me.
type getMeResponse struct {
	UserID           int32  `json:"userId"`
	LoginID          string `json:"loginId"`
	OrganizationName string `json:"organizationName"`
}

// NewAuthMiddleware returns a Gin middleware that validates requests
// by forwarding the Authorization header to cocotola-auth's /api/v1/auth/me endpoint.
func NewAuthMiddleware(authBaseURL string, httpClient *http.Client) gin.HandlerFunc {
	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, "QuestionAuthMiddleware"))
	meURL := authBaseURL + "/api/v1/auth/me"

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			logger.InfoContext(ctx, "no Authorization header")
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, meURL, nil)
		if err != nil {
			logger.ErrorContext(ctx, "create request", slog.Any("error", err))
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}
		req.Header.Set("Authorization", authorization)

		resp, err := httpClient.Do(req)
		if err != nil {
			logger.ErrorContext(ctx, "call auth service", slog.Any("error", err))
			c.Status(http.StatusBadGateway)
			c.Abort()
			return
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logger.ErrorContext(ctx, "close response body", slog.Any("error", err))
			}
		}()

		if resp.StatusCode != http.StatusOK {
			logger.WarnContext(ctx, "auth service returned non-200", slog.Int("status", resp.StatusCode))
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		var me getMeResponse
		if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
			logger.ErrorContext(ctx, "decode auth response", slog.Any("error", err))
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		c.Set(controller.ContextFieldUserID{}, int(me.UserID))
		c.Set(controller.ContextFieldOrganizationName{}, me.OrganizationName)
		c.Next()
	}
}

// AuthServiceOrganizationResolver returns an OrganizationIDResolver that resolves
// organization names by calling cocotola-auth's API.
// NOTE: cocotola-auth currently returns the organization name but not the ID
// via the /auth/me endpoint. This resolver uses a simple mapping approach.
// When cocotola-auth adds a dedicated org lookup API, this should be updated.
func AuthServiceOrganizationResolver(authBaseURL string, httpClient *http.Client) func(ctx context.Context, name string) (int, error) {
	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, "OrgResolver"))

	return func(ctx context.Context, name string) (int, error) {
		// Call cocotola-auth to resolve organization name to ID.
		// Uses a hypothetical endpoint: GET /api/v1/organizations?name=<name>
		// Until this API exists, we return an error to indicate the need for implementation.
		params := url.Values{}
		params.Set("name", name)
		reqURL := authBaseURL + "/api/v1/organizations?" + params.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return 0, fmt.Errorf("create request: %w", err)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return 0, fmt.Errorf("call auth service: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logger.ErrorContext(ctx, "close response body", slog.Any("error", err))
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("auth service returned status %d", resp.StatusCode)
		}

		var result struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return 0, fmt.Errorf("decode response: %w", err)
		}

		return result.ID, nil
	}
}

// AuthServiceAuthorizationChecker calls cocotola-auth's API to check RBAC permissions.
type AuthServiceAuthorizationChecker struct {
	authBaseURL string
	httpClient  *http.Client
	logger      *slog.Logger
}

// NewAuthServiceAuthorizationChecker creates a new AuthServiceAuthorizationChecker.
func NewAuthServiceAuthorizationChecker(authBaseURL string, httpClient *http.Client) *AuthServiceAuthorizationChecker {
	return &AuthServiceAuthorizationChecker{
		authBaseURL: authBaseURL,
		httpClient:  httpClient,
		logger:      slog.Default().With(slog.String(liblogging.LoggerNameKey, "AuthzChecker")),
	}
}

// IsAllowed checks if the given action on the resource is allowed.
// NOTE: cocotola-auth does not yet have a dedicated authz check API.
// Until that API is available, this delegates to a hypothetical endpoint:
// GET /api/v1/authz/check?org=<orgID>&user=<userID>&action=<action>&resource=<resource>
func (c *AuthServiceAuthorizationChecker) IsAllowed(ctx context.Context, organizationID int, operatorID int, action domain.Action, resource domain.Resource) (bool, error) {
	params := url.Values{}
	params.Set("org", strconv.Itoa(organizationID))
	params.Set("user", strconv.Itoa(operatorID))
	params.Set("action", action.Value())
	params.Set("resource", resource.Value())
	reqURL := c.authBaseURL + "/api/v1/authz/check?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("call auth service: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.ErrorContext(ctx, "close response body", slog.Any("error", err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var result struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode response: %w", err)
	}

	return result.Allowed, nil
}
