package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Test-only re-exports of unexported identifiers used by the external
// gateway_test package. Confined to *_test.go so it does not affect
// the public API.

type (
	StudyRecordIter        = studyRecordIter
	StudyRecordBulkDeleter = studyRecordBulkDeleter
)

func DeleteStudyRecordDocs(iter StudyRecordIter, bw StudyRecordBulkDeleter) error {
	return deleteStudyRecordDocs(iter, bw)
}

const OrganizationIDCacheMaxEntries = organizationIDCacheMaxEntries

func NewCachedOrganizationResolverWithClock(resolver func(ctx context.Context, name string) (string, error), ttl time.Duration, now func() time.Time) func(ctx context.Context, name string) (string, error) {
	return newCachedOrganizationResolver(resolver, ttl, now)
}

func NewAuthMiddlewareWithClock(authBaseURL string, httpClient *http.Client, meCacheTTL time.Duration, now func() time.Time) gin.HandlerFunc {
	return newAuthMiddleware(authBaseURL, httpClient, meCacheTTL, now)
}
