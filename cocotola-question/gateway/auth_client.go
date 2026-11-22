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

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

const (
	// maxResponseBodySize limits the size of HTTP response bodies read from auth service.
	maxResponseBodySize = 1 << 20 // 1 MB
	// maxErrorBodySize limits the size of error response bodies read from auth service.
	maxErrorBodySize = 512

	serviceAuthHeader = "X-Service-Api-Key"
)

// authServiceClient is a base HTTP client for calling cocotola-auth APIs.
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

// do sends a request to the auth service and passes the response to handle.
func (c *authServiceClient) do(ctx context.Context, method string, reqURL string, body io.Reader, handle func(*http.Response) error) error {
	header := http.Header{}
	if c.apiKey != "" {
		header.Set(serviceAuthHeader, c.apiKey)
	}
	if body != nil {
		header.Set("Content-Type", "application/json")
	}

	return c.send(ctx, method, reqURL, header, body, handle)
}

func (c *authServiceClient) send(ctx context.Context, method string, reqURL string, header http.Header, body io.Reader, handle func(*http.Response) error) error {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header = header

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call auth service: %w: %w", domain.ErrAuthServiceUnavailable, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.ErrorContext(ctx, "close response body", slog.Any("error", err))
		}
	}()

	return handle(resp)
}

// request performs an HTTP request to the auth service and decodes the JSON response.
func (c *authServiceClient) request(ctx context.Context, method string, reqURL string, body io.Reader, result any) error {
	return c.do(ctx, method, reqURL, body, func(resp *http.Response) error {
		if resp.StatusCode != http.StatusOK {
			return statusError(resp)
		}

		return decodeJSON(resp, result)
	})
}

func statusError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
	if err != nil {
		return fmt.Errorf("auth service returned status %d (read body: %w)", resp.StatusCode, err)
	}

	return fmt.Errorf("auth service returned status %d: %s", resp.StatusCode, string(body))
}

func decodeJSON(resp *http.Response, result any) error {
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBodySize)).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// AuthServiceOrganizationResolver resolves organization names by calling
// cocotola-auth's internal GET /api/v1/internal/auth/organization endpoint.
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

// AuthServiceSpaceTypeFetcher calls cocotola-auth's internal API to look up a space record by ID.
type AuthServiceSpaceTypeFetcher struct {
	authServiceClient
}

// NewAuthServiceSpaceTypeFetcher creates a new AuthServiceSpaceTypeFetcher.
func NewAuthServiceSpaceTypeFetcher(authBaseURL string, apiKey string, httpClient *http.Client) *AuthServiceSpaceTypeFetcher {
	return &AuthServiceSpaceTypeFetcher{
		authServiceClient: newAuthServiceClient(authBaseURL, apiKey, httpClient, "SpaceTypeFetcher"),
	}
}

// FetchSpaceType returns the SpaceType of the given space.
// Returns domain.ErrSpaceNotFound when cocotola-auth responds with 404.
func (f *AuthServiceSpaceTypeFetcher) FetchSpaceType(ctx context.Context, spaceID string) (domain.SpaceType, error) {
	if spaceID == "" {
		return domain.SpaceType{}, fmt.Errorf("space id is required: %w", domain.ErrInvalidArgument)
	}

	reqURL := f.authBaseURL + "/api/v1/internal/auth/space/" + url.PathEscape(spaceID)

	var result struct {
		SpaceType string `json:"spaceType"`
	}
	err := f.do(ctx, http.MethodGet, reqURL, nil, func(resp *http.Response) error {
		switch resp.StatusCode {
		case http.StatusOK:
			return decodeJSON(resp, &result)
		case http.StatusNotFound:
			return domain.ErrSpaceNotFound
		default:
			return statusError(resp)
		}
	})
	if err != nil {
		return domain.SpaceType{}, fmt.Errorf("get auth space %s: %w", spaceID, err)
	}

	spaceType, err := domain.NewSpaceType(result.SpaceType)
	if err != nil {
		return domain.SpaceType{}, fmt.Errorf("space %s: %w", spaceID, err)
	}
	return spaceType, nil
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
		return 0, fmt.Errorf("unexpected max workbooks %d for user %s from auth service: %w", result.MaxWorkbooks, userID, domain.ErrInvalidMaxWorkbooks)
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

// AuthServicePolicyAdder calls cocotola-auth's internal API to add RBAC policies.
type AuthServicePolicyAdder struct {
	authServiceClient
}

// NewAuthServicePolicyAdder creates a new AuthServicePolicyAdder.
func NewAuthServicePolicyAdder(authBaseURL string, apiKey string, httpClient *http.Client) *AuthServicePolicyAdder {
	return &AuthServicePolicyAdder{
		authServiceClient: newAuthServiceClient(authBaseURL, apiKey, httpClient, "PolicyAdder"),
	}
}

// addPolicyRequestBody is the JSON body for POST /auth/authz/policy.
type addPolicyRequestBody struct {
	Org      string `json:"org"`
	User     string `json:"user"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
	Effect   string `json:"effect"`
}

// AddPolicyForUser adds a per-user RBAC policy via cocotola-auth's internal API.
func (c *AuthServicePolicyAdder) AddPolicyForUser(ctx context.Context, organizationID string, userID string, action domain.Action, resource domain.Resource, effect domain.Effect) error {
	reqURL := c.authBaseURL + "/api/v1/internal/auth/authz/policy"

	body := addPolicyRequestBody{
		Org:      organizationID,
		User:     userID,
		Action:   action.Value(),
		Resource: resource.Value(),
		Effect:   effect.Value(),
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal add policy request: %w", err)
	}

	err = c.do(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes), func(resp *http.Response) error {
		if resp.StatusCode != http.StatusNoContent {
			return statusError(resp)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("add policy for user %s: %w", userID, err)
	}

	return nil
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
