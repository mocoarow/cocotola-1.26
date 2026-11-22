package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

const (
	// serviceAuthHeader names the inbound authentication header used by the
	// internal cocotola-question endpoints. It is intentionally not called
	// "ApiKey" to avoid gosec G101 false positives on the constant name.
	serviceAuthHeader         = "X-Service-Api-Key"
	organizationIDHeader      = "X-Organization-Id"
	contentTypeJSON           = "application/json"
	maxBodyBytes              = 1 << 20
	maxListQuestionsBodyBytes = 32 << 20
	maxErrorBodyBytes         = 512
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

type workbookListItemResponse struct {
	WorkbookID  string `json:"workbookId"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type listWorkbooksResponse struct {
	Workbooks []workbookListItemResponse `json:"workbooks"`
}

// ListWorkbooks calls GET /api/v1/internal/workbook?spaceId=... and returns
// the workbooks present in the space for the given organization.
func (c *QuestionAPIClient) ListWorkbooks(ctx context.Context, organizationID, spaceID string) ([]seed.WorkbookListItem, error) {
	q := url.Values{}
	q.Set("spaceId", spaceID)
	reqURL := c.baseURL + "/api/v1/internal/workbook?" + q.Encode()

	var resp listWorkbooksResponse
	if err := c.do(ctx, http.MethodGet, reqURL, organizationID, nil, &resp, maxBodyBytes); err != nil {
		return nil, fmt.Errorf("list workbooks (space %s): %w", spaceID, err)
	}
	items := make([]seed.WorkbookListItem, 0, len(resp.Workbooks))
	for _, w := range resp.Workbooks {
		items = append(items, seed.WorkbookListItem{WorkbookID: w.WorkbookID, Title: w.Title, Description: w.Description})
	}
	return items, nil
}

// workbookVisibilityPublic is sent on every create because cocotola-question
// requires the field, although it overwrites the value with the space type's.
const workbookVisibilityPublic = "public"

type createWorkbookRequest struct {
	SpaceID     string `json:"spaceId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	Language    string `json:"language"`
}

type workbookResponse struct {
	WorkbookID string `json:"workbookId"`
}

// CreateWorkbook calls POST /api/v1/internal/workbook.
func (c *QuestionAPIClient) CreateWorkbook(ctx context.Context, organizationID string, input seed.CreateWorkbookInput) (string, error) {
	reqURL := c.baseURL + "/api/v1/internal/workbook"
	body := createWorkbookRequest{
		SpaceID:     input.SpaceID,
		Title:       input.Title,
		Description: input.Description,
		Visibility:  workbookVisibilityPublic,
		Language:    input.Language,
	}
	var resp workbookResponse
	if err := c.do(ctx, http.MethodPost, reqURL, organizationID, body, &resp, maxBodyBytes); err != nil {
		return "", fmt.Errorf("create workbook: %w", err)
	}
	return resp.WorkbookID, nil
}

type questionListItemResponse struct {
	QuestionID string   `json:"questionId"`
	Tags       []string `json:"tags"`
}

type listQuestionsResponse struct {
	Questions []questionListItemResponse `json:"questions"`
}

// ListQuestions calls GET /api/v1/internal/workbook/{workbookId}/question.
func (c *QuestionAPIClient) ListQuestions(ctx context.Context, organizationID, workbookID string) ([]seed.QuestionListItem, error) {
	reqURL := c.baseURL + "/api/v1/internal/workbook/" + url.PathEscape(workbookID) + "/question"

	var resp listQuestionsResponse
	if err := c.do(ctx, http.MethodGet, reqURL, organizationID, nil, &resp, maxListQuestionsBodyBytes); err != nil {
		return nil, fmt.Errorf("list questions (workbook %s): %w", workbookID, err)
	}
	items := make([]seed.QuestionListItem, 0, len(resp.Questions))
	for _, q := range resp.Questions {
		items = append(items, seed.QuestionListItem{QuestionID: q.QuestionID, Tags: q.Tags})
	}
	return items, nil
}

type addQuestionRequest struct {
	QuestionType string   `json:"questionType"`
	Content      string   `json:"content"`
	Tags         []string `json:"tags,omitempty"`
	OrderIndex   int32    `json:"orderIndex"`
}

// AddQuestion calls POST /api/v1/internal/workbook/{workbookId}/question.
// A 400 response is returned as seed.ErrQuestionRejected.
func (c *QuestionAPIClient) AddQuestion(ctx context.Context, organizationID, workbookID string, input seed.AddQuestionInput) error {
	reqURL := c.baseURL + "/api/v1/internal/workbook/" + url.PathEscape(workbookID) + "/question"
	body := addQuestionRequest{
		QuestionType: input.QuestionType,
		Content:      input.Content,
		Tags:         input.Tags,
		OrderIndex:   input.OrderIndex,
	}
	err := c.do(ctx, http.MethodPost, reqURL, organizationID, body, nil, maxBodyBytes)
	if statusErr, ok := errors.AsType[*HTTPStatusError](err); ok && statusErr.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("add question (workbook %s): %w: %w", workbookID, seed.ErrQuestionRejected, err)
	}
	if err != nil {
		return fmt.Errorf("add question (workbook %s): %w", workbookID, err)
	}
	return nil
}

// do issues an authenticated HTTP request and decodes the JSON response into out
// (when out is non-nil).
func (c *QuestionAPIClient) do(ctx context.Context, method, reqURL, organizationID string, body, out any, maxDecodeBytes int64) error {
	req, err := c.newRequest(ctx, method, reqURL, organizationID, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
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

	if err := json.NewDecoder(http.MaxBytesReader(nil, resp.Body, maxDecodeBytes)).Decode(out); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return fmt.Errorf("decode response: response body exceeds %d bytes: %w", maxDecodeBytes, err)
		}
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

// HTTPStatusError is returned when cocotola-question answers with a non-2xx
// status. Body is a truncated snippet of the response body.
type HTTPStatusError struct {
	StatusCode int
	Body       string
}

func (e *HTTPStatusError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("status %d", e.StatusCode)
	}
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Body)
}

func statusError(resp *http.Response) error {
	errBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
	if readErr != nil {
		return fmt.Errorf("%w: read error body: %w", &HTTPStatusError{StatusCode: resp.StatusCode, Body: ""}, readErr)
	}
	return &HTTPStatusError{StatusCode: resp.StatusCode, Body: string(errBody)}
}
