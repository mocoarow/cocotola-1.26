package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

type getMeResponse struct {
	UserID           string `json:"userId"`
	OrganizationName string `json:"organizationName"`
	Language         string `json:"language"`
}

// AuthServiceMeFetcher resolves the caller of an Authorization header via cocotola-auth's GET /api/v1/auth/me.
type AuthServiceMeFetcher struct {
	authServiceClient
	cache *authMeCache
}

// NewAuthServiceMeFetcher creates a new AuthServiceMeFetcher.
func NewAuthServiceMeFetcher(authBaseURL string, httpClient *http.Client, meCacheTTL time.Duration) *AuthServiceMeFetcher {
	return newAuthServiceMeFetcher(authBaseURL, httpClient, meCacheTTL, time.Now)
}

func newAuthServiceMeFetcher(authBaseURL string, httpClient *http.Client, meCacheTTL time.Duration, now func() time.Time) *AuthServiceMeFetcher {
	return &AuthServiceMeFetcher{
		authServiceClient: newAuthServiceClient(authBaseURL, "", httpClient, "MeFetcher"),
		cache:             newAuthMeCache(meCacheTTL, now),
	}
}

// FetchMe resolves the caller identified by authorization.
func (f *AuthServiceMeFetcher) FetchMe(ctx context.Context, authorization string) (domain.AuthenticatedUser, error) {
	if me, ok := f.cache.get(authorization); ok {
		return me, nil
	}

	header := http.Header{}
	header.Set("Authorization", authorization)

	var result getMeResponse
	err := f.send(ctx, http.MethodGet, f.authBaseURL+"/api/v1/auth/me", header, nil, func(resp *http.Response) error {
		switch resp.StatusCode {
		case http.StatusOK:
			return decodeJSON(resp, &result)
		case http.StatusUnauthorized:
			return domain.ErrUnauthenticated
		default:
			return fmt.Errorf("auth me returned status %d: %w", resp.StatusCode, domain.ErrAuthServiceUnavailable)
		}
	})
	if err != nil {
		return domain.AuthenticatedUser{}, fmt.Errorf("fetch auth me: %w", err)
	}

	me := domain.AuthenticatedUser{
		UserID:           result.UserID,
		OrganizationName: result.OrganizationName,
		Language:         result.Language,
	}
	f.cache.set(authorization, me)
	return me, nil
}
