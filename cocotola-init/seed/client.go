package seed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

const (
	// serviceAuthHeader names the inbound authentication header used by the
	// internal cocotola-question endpoints. It is intentionally not called
	// "ApiKey" to avoid gosec G101 false positives on the constant name.
	serviceAuthHeader    = "X-Service-Api-Key"
	organizationIDHeader = "X-Organization-Id"
	contentTypeJSON      = "application/json"
	maxBodyBytes         = 1 << 20
	maxErrorBodyBytes    = 512
)

// QuestionAPIClient is a minimal HTTP client for cocotola-question's
// /api/v1/internal/* endpoints.
type QuestionAPIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewQuestionAPIClient constructs a QuestionAPIClient.
func NewQuestionAPIClient(baseURL, apiKey string, httpClient *http.Client) *QuestionAPIClient {
	return &QuestionAPIClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// WorkbookListItem is the subset of the ListWorkbooks response the seeder cares about.
type WorkbookListItem struct {
	WorkbookID  string `json:"workbookId"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type listWorkbooksResponse struct {
	Workbooks []WorkbookListItem `json:"workbooks"`
}

// ListWorkbooks calls GET /api/v1/internal/workbook?spaceId=... and returns
// the workbooks present in the space for the given organization.
func (c *QuestionAPIClient) ListWorkbooks(ctx context.Context, organizationID, spaceID string) ([]WorkbookListItem, error) {
	q := url.Values{}
	q.Set("spaceId", spaceID)
	reqURL := c.baseURL + "/api/v1/internal/workbook?" + q.Encode()

	var resp listWorkbooksResponse
	if err := c.do(ctx, http.MethodGet, reqURL, organizationID, nil, &resp); err != nil {
		return nil, fmt.Errorf("list workbooks (space %s): %w", spaceID, err)
	}
	return resp.Workbooks, nil
}

// CreateWorkbookRequest mirrors cocotola-question's CreateWorkbookRequest.
// Visibility is ignored server-side (the SpaceType determines it), so we
// always send "public" for clarity.
type CreateWorkbookRequest struct {
	SpaceID     string `json:"spaceId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

// WorkbookResponse mirrors cocotola-question's WorkbookResponse.
type WorkbookResponse struct {
	WorkbookID string `json:"workbookId"`
}

// CreateWorkbook calls POST /api/v1/internal/workbook.
func (c *QuestionAPIClient) CreateWorkbook(ctx context.Context, organizationID string, body CreateWorkbookRequest) (string, error) {
	reqURL := c.baseURL + "/api/v1/internal/workbook"
	var resp WorkbookResponse
	if err := c.do(ctx, http.MethodPost, reqURL, organizationID, body, &resp); err != nil {
		return "", fmt.Errorf("create workbook: %w", err)
	}
	return resp.WorkbookID, nil
}

// QuestionListItem is the subset of the ListQuestions response the seeder cares about.
type QuestionListItem struct {
	QuestionID string   `json:"questionId"`
	Tags       []string `json:"tags"`
}

type listQuestionsResponse struct {
	Questions []QuestionListItem `json:"questions"`
}

// ListQuestions calls GET /api/v1/internal/workbook/{workbookId}/question.
func (c *QuestionAPIClient) ListQuestions(ctx context.Context, organizationID, workbookID string) ([]QuestionListItem, error) {
	reqURL := c.baseURL + "/api/v1/internal/workbook/" + url.PathEscape(workbookID) + "/question"

	var resp listQuestionsResponse
	if err := c.do(ctx, http.MethodGet, reqURL, organizationID, nil, &resp); err != nil {
		return nil, fmt.Errorf("list questions (workbook %s): %w", workbookID, err)
	}
	return resp.Questions, nil
}

// AddQuestionRequest mirrors cocotola-question's AddQuestionRequest.
type AddQuestionRequest struct {
	QuestionType string   `json:"questionType"`
	Content      string   `json:"content"`
	Tags         []string `json:"tags,omitempty"`
	OrderIndex   int32    `json:"orderIndex"`
}

// AddQuestion calls POST /api/v1/internal/workbook/{workbookId}/question.
func (c *QuestionAPIClient) AddQuestion(ctx context.Context, organizationID, workbookID string, body AddQuestionRequest) error {
	reqURL := c.baseURL + "/api/v1/internal/workbook/" + url.PathEscape(workbookID) + "/question"
	if err := c.do(ctx, http.MethodPost, reqURL, organizationID, body, nil); err != nil {
		return fmt.Errorf("add question (workbook %s): %w", workbookID, err)
	}
	return nil
}

// do issues an authenticated HTTP request and decodes the JSON response into out
// (when out is non-nil). Non-2xx responses are surfaced as wrapped errors.
func (c *QuestionAPIClient) do(ctx context.Context, method, reqURL, organizationID string, body, out any) error {
	req, err := c.newRequest(ctx, method, reqURL, organizationID, body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		drainBody(ctx, resp.Body)
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return statusError(resp)
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes)).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// newRequest builds an authenticated *http.Request, JSON-encoding body when non-nil.
func (c *QuestionAPIClient) newRequest(ctx context.Context, method, reqURL, organizationID string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set(serviceAuthHeader, c.apiKey)
	req.Header.Set(organizationIDHeader, organizationID)
	if body != nil {
		req.Header.Set("Content-Type", contentTypeJSON)
	}
	return req, nil
}

// drainBody reads any leftover bytes so the TCP connection can be returned
// to the keep-alive pool. The caller is responsible for closing the body.
func drainBody(ctx context.Context, body io.Reader) {
	if _, copyErr := io.Copy(io.Discard, body); copyErr != nil {
		slog.WarnContext(ctx, "drain response body", "error", copyErr)
	}
}

// statusError formats a non-2xx response into a single error, including a
// truncated body snippet to aid debugging.
func statusError(resp *http.Response) error {
	errBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
	if readErr != nil {
		return fmt.Errorf("status %d: read error body: %w", resp.StatusCode, readErr)
	}
	return fmt.Errorf("status %d: %s", resp.StatusCode, string(errBody))
}
