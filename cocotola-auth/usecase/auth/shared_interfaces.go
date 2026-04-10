package auth

import (
	"context"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// Shared unexported interfaces used by multiple command/query files.
// Each interface follows ISP — only the methods needed by its consumers.

type accessTokenSaver interface {
	Save(ctx context.Context, token *domaintoken.AccessToken) error
}

type jwtCreator interface {
	CreateAccessToken(loginID string, userID domain.AppUserID, organizationName string, jti string) (string, error)
}

type jwtParser interface {
	ParseAccessToken(tokenString string) (*authservice.UserInfo, string, error)
}

type accessTokenCacheSetter interface {
	SetAccessToken(jti string, token *domaintoken.AccessToken)
}

type whitelistFinder interface {
	FindByUserID(ctx context.Context, userID domain.AppUserID) ([]domaintoken.WhitelistEntry, error)
}

type sessionTokenCacheReadWriter interface {
	GetSessionToken(hash string) (*domaintoken.SessionToken, bool)
	SetSessionToken(hash string, token *domaintoken.SessionToken)
}
