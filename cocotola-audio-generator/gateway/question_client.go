// Package gateway provides outbound clients for cocotola-audio-generator.
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
	"strconv"
	"strings"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/domain"
)

// QuestionAPIClient talks to cocotola-question's internal /api/v1/internal/audio/*
// endpoints. The X-Service-Api-Key header authenticates batch traffic.
type QuestionAPIClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewQuestionAPIClient returns a configured client.
func NewQuestionAPIClient(baseURL, apiKey string, timeout time.Duration) *QuestionAPIClient {
	return &QuestionAPIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http: &http.Client{
			Timeout:       timeout,
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
		},
	}
}

type pendingItemDTO struct {
	WorkbookID  string `json:"workbookId"`
	QuestionID  string `json:"questionId"`
	SourceText  string `json:"sourceText"`
	SourceLang  string `json:"sourceLang"`
	TargetText  string `json:"targetText"`
	TargetLang  string `json:"targetLang"`
	InputHash   string `json:"inputHash"`
	FailedTries int    `json:"failedTries"`
}

type pendingResponseDTO struct {
	Items []pendingItemDTO `json:"items"`
}

// ListPending fetches up to limit pending audio items.
func (c *QuestionAPIClient) ListPending(ctx context.Context, limit int) ([]domain.PendingItem, error) {
	u, err := url.Parse(c.baseURL + "/api/v1/internal/audio/questions/pending")
	if err != nil {
		return nil, fmt.Errorf("parse pending url: %w", err)
	}
	q := u.Query()
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("new pending request: %w", err)
	}
	c.setAuthHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do pending request: %w", err)
	}
	defer closeBodyAndLog(ctx, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list pending: %s", readErrorBody(resp))
	}
	var out pendingResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode pending response: %w", err)
	}
	items := make([]domain.PendingItem, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, domain.PendingItem{
			WorkbookID:  it.WorkbookID,
			QuestionID:  it.QuestionID,
			SourceText:  it.SourceText,
			SourceLang:  it.SourceLang,
			TargetText:  it.TargetText,
			TargetLang:  it.TargetLang,
			InputHash:   it.InputHash,
			FailedTries: it.FailedTries,
		})
	}
	return items, nil
}

type claimRequest struct {
	InputHash string `json:"inputHash"`
}

// Claim transitions a question's audio from pending to generating.
// Returns ErrClaimRace when another worker holds the entry.
func (c *QuestionAPIClient) Claim(ctx context.Context, item domain.PendingItem) error {
	body, err := json.Marshal(claimRequest{InputHash: item.InputHash})
	if err != nil {
		return fmt.Errorf("marshal claim: %w", err)
	}
	endpoint := fmt.Sprintf("%s/api/v1/internal/audio/questions/%s/%s/claim",
		c.baseURL, url.PathEscape(item.WorkbookID), url.PathEscape(item.QuestionID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new claim request: %w", err)
	}
	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("do claim request: %w", err)
	}
	defer closeBodyAndLog(ctx, resp.Body)
	if resp.StatusCode == http.StatusConflict {
		return domain.ErrClaimRace
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("claim audio: %s", readErrorBody(resp))
	}
	return nil
}

type completeRefDTO struct {
	Path        string  `json:"path"`
	DurationSec float64 `json:"durationSec"`
	SizeBytes   int64   `json:"sizeBytes"`
}

type completeRequest struct {
	InputHash string                    `json:"inputHash"`
	Refs      map[string]completeRefDTO `json:"refs"`
}

// Complete transitions a question's audio to ready with the supplied refs.
func (c *QuestionAPIClient) Complete(ctx context.Context, item domain.PendingItem, refs map[string]domain.AudioRef) error {
	dto := completeRequest{
		InputHash: item.InputHash,
		Refs:      make(map[string]completeRefDTO, len(refs)),
	}
	for k, v := range refs {
		dto.Refs[k] = completeRefDTO{Path: v.Path, DurationSec: v.DurationSec, SizeBytes: v.SizeBytes}
	}
	return c.postJSON(ctx, "complete", item, dto)
}

type failRequest struct {
	InputHash string `json:"inputHash"`
	Reason    string `json:"reason"`
}

// Fail transitions a question's audio to failed and increments the attempt counter.
func (c *QuestionAPIClient) Fail(ctx context.Context, item domain.PendingItem, reason string) error {
	return c.postJSON(ctx, "fail", item, failRequest{InputHash: item.InputHash, Reason: reason})
}

func (c *QuestionAPIClient) postJSON(ctx context.Context, action string, item domain.PendingItem, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", action, err)
	}
	endpoint := fmt.Sprintf("%s/api/v1/internal/audio/questions/%s/%s/%s",
		c.baseURL, url.PathEscape(item.WorkbookID), url.PathEscape(item.QuestionID), action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new %s request: %w", action, err)
	}
	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("do %s request: %w", action, err)
	}
	defer closeBodyAndLog(ctx, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s audio: %s", action, readErrorBody(resp))
	}
	return nil
}

func (c *QuestionAPIClient) setAuthHeaders(req *http.Request) {
	req.Header.Set("X-Service-Api-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
}

func readErrorBody(resp *http.Response) string {
	if resp == nil {
		return "no response"
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("status=%d, read body: %v", resp.StatusCode, err)
	}
	return fmt.Sprintf("status=%d, body=%s", resp.StatusCode, string(body))
}

// ErrNotFound surfaces when an item endpoint returns 404.
var ErrNotFound = errors.New("question not found")

// closeBodyAndLog drains/closes an HTTP response body and logs any close
// error. Centralized so each call site does not have to re-implement the
// defer block (and so we satisfy errcheck without silent suppressions).
//
// ctx is used purely to propagate request-scoped attributes (trace IDs, etc.)
// to the structured logger; it is not consulted for cancellation since Close
// is non-blocking.
func closeBodyAndLog(ctx context.Context, body io.Closer) {
	if err := body.Close(); err != nil {
		slog.WarnContext(ctx, "close response body", slog.Any("error", err))
	}
}

type reclaimStaleRequest struct {
	StaleAfterSec int `json:"staleAfterSec"`
	Limit         int `json:"limit"`
}

type reclaimStaleResponse struct {
	Reclaimed int `json:"reclaimed"`
}

// ReclaimStale asks cocotola-question to transition any "generating" audio
// entries stuck longer than staleAfter back to "pending". Returns the number
// of entries reclaimed in this call.
func (c *QuestionAPIClient) ReclaimStale(ctx context.Context, staleAfter time.Duration, limit int) (int, error) {
	body, err := json.Marshal(reclaimStaleRequest{
		StaleAfterSec: int(staleAfter.Seconds()),
		Limit:         limit,
	})
	if err != nil {
		return 0, fmt.Errorf("marshal reclaim stale: %w", err)
	}
	endpoint := c.baseURL + "/api/v1/internal/audio/questions/reclaim-stale"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("new reclaim stale request: %w", err)
	}
	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do reclaim stale request: %w", err)
	}
	defer closeBodyAndLog(ctx, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("reclaim stale: %s", readErrorBody(resp))
	}
	var out reclaimStaleResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, fmt.Errorf("decode reclaim stale response: %w", err)
	}
	return out.Reclaimed, nil
}
