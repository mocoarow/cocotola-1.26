package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// maxResponseBodySize limits the size of HTTP response bodies read from auth service.
const maxResponseBodySize = 1 << 20 // 1 MB

// authServiceClient is a base HTTP client for calling cocotola-auth internal APIs.
type authServiceClient struct {
	authBaseURL string
	apiKey      string
	httpClient  *http.Client
	logger      *slog.Logger
}

// newAuthServiceClient creates a new authServiceClient.
func newAuthServiceClient(authBaseURL string, apiKey string, httpClient *http.Client, loggerName string) authServiceClient {
	return authServiceClient{
		authBaseURL: authBaseURL,
		apiKey:      apiKey,
		httpClient:  httpClient,
		logger:      slog.Default().With(slog.String(liblogging.LoggerNameKey, loggerName)),
	}
}

// request performs an HTTP request to the auth service and decodes the JSON response.
func (c *authServiceClient) request(ctx context.Context, method string, reqURL string, body io.Reader, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("X-Service-Api-Key", c.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call auth service: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.ErrorContext(ctx, "close response body", slog.Any("error", err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("auth service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBodySize)).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// getMeResponse represents the response from cocotola-auth GET /api/v1/auth/me.
type getMeResponse struct {
	UserID           string `json:"userId"`
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
		if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBodySize)).Decode(&me); err != nil {
			logger.ErrorContext(ctx, "decode auth response", slog.Any("error", err))
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		c.Set(controller.ContextFieldUserID{}, me.UserID)
		c.Set(controller.ContextFieldOrganizationName{}, me.OrganizationName)
		c.Next()
	}
}

// AuthServiceOrganizationResolver returns an OrganizationIDResolver that resolves
// organization names by calling cocotola-auth's internal GET /api/v1/internal/auth/organization endpoint.
func AuthServiceOrganizationResolver(authBaseURL string, apiKey string, httpClient *http.Client) func(ctx context.Context, name string) (string, error) {
	client := newAuthServiceClient(authBaseURL, apiKey, httpClient, "OrgResolver")

	return func(ctx context.Context, name string) (string, error) {
		params := url.Values{}
		params.Set("name", name)
		reqURL := client.authBaseURL + "/api/v1/internal/auth/organization?" + params.Encode()

		var result struct {
			ID string `json:"id"`
		}
		if err := client.request(ctx, http.MethodGet, reqURL, nil, &result); err != nil {
			return "", fmt.Errorf("resolve organization %s: %w", name, err)
		}

		return result.ID, nil
	}
}

// AuthServiceMaxWorkbooksFetcher calls cocotola-auth's internal API to fetch the max workbooks setting for a user.
type AuthServiceMaxWorkbooksFetcher struct {
	authServiceClient
}

// NewAuthServiceMaxWorkbooksFetcher creates a new AuthServiceMaxWorkbooksFetcher.
func NewAuthServiceMaxWorkbooksFetcher(authBaseURL string, apiKey string, httpClient *http.Client) *AuthServiceMaxWorkbooksFetcher {
	return &AuthServiceMaxWorkbooksFetcher{
		authServiceClient: newAuthServiceClient(authBaseURL, apiKey, httpClient, "MaxWorkbooksFetcher"),
	}
}

// FetchMaxWorkbooks returns the max workbooks limit for the given user.
func (f *AuthServiceMaxWorkbooksFetcher) FetchMaxWorkbooks(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, fmt.Errorf("user id is required: %w", domain.ErrInvalidArgument)
	}

	params := url.Values{}
	params.Set("user_id", userID)
	reqURL := f.authBaseURL + "/api/v1/internal/auth/user-setting?" + params.Encode()

	var result struct {
		MaxWorkbooks int `json:"maxWorkbooks"`
	}
	if err := f.request(ctx, http.MethodGet, reqURL, nil, &result); err != nil {
		return 0, fmt.Errorf("fetch max workbooks for user %s: %w", userID, err)
	}

	if result.MaxWorkbooks <= 0 {
		return 0, fmt.Errorf("invalid max workbooks value %d from auth service", result.MaxWorkbooks)
	}

	return result.MaxWorkbooks, nil
}

// AuthServiceAuthorizationChecker calls cocotola-auth's internal API to check RBAC permissions.
type AuthServiceAuthorizationChecker struct {
	authServiceClient
}

// NewAuthServiceAuthorizationChecker creates a new AuthServiceAuthorizationChecker.
func NewAuthServiceAuthorizationChecker(authBaseURL string, apiKey string, httpClient *http.Client) *AuthServiceAuthorizationChecker {
	return &AuthServiceAuthorizationChecker{
		authServiceClient: newAuthServiceClient(authBaseURL, apiKey, httpClient, "AuthzChecker"),
	}
}

// authzCheckRequestBody is the JSON body for POST /auth/authz/check.
type authzCheckRequestBody struct {
	Org      string `json:"org"`
	User     string `json:"user"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

// IsAllowed checks if the given action on the resource is allowed
// by calling cocotola-auth's internal POST /api/v1/internal/auth/authz/check endpoint.
func (c *AuthServiceAuthorizationChecker) IsAllowed(ctx context.Context, organizationID string, operatorID string, action domain.Action, resource domain.Resource) (bool, error) {
	reqURL := c.authBaseURL + "/api/v1/internal/auth/authz/check"

	body := authzCheckRequestBody{
		Org:      organizationID,
		User:     operatorID,
		Action:   action.Value(),
		Resource: resource.Value(),
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return false, fmt.Errorf("marshal authz check request: %w", err)
	}

	var result struct {
		Allowed bool `json:"allowed"`
	}
	if err := c.request(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes), &result); err != nil {
		return false, fmt.Errorf("check authorization: %w", err)
	}

	return result.Allowed, nil
}
