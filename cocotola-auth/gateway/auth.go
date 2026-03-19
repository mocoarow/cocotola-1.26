package gateway

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

type userClaims struct {
	LoginID          string `json:"loginId"`
	UserID           int    `json:"userId"`
	OrganizationName string `json:"organizationName"`
	jwt.RegisteredClaims
}

// JWTManager implements JWT access token creation and parsing using HMAC signing.
type JWTManager struct {
	signingKey    []byte
	signingMethod jwt.SigningMethod
	tokenTimeout  time.Duration
}

// NewJWTManager returns a new JWTManager with the given signing parameters.
func NewJWTManager(signingKey []byte, signingMethod jwt.SigningMethod, tokenTimeout time.Duration) *JWTManager {
	return &JWTManager{
		signingKey:    signingKey,
		signingMethod: signingMethod,
		tokenTimeout:  tokenTimeout,
	}
}

// CreateAccessToken generates a signed JWT for the given user with the specified JTI.
func (m *JWTManager) CreateAccessToken(loginID string, userID int, organizationName string, jti string) (string, error) {
	now := time.Now()
	claims := userClaims{
		LoginID:          loginID,
		UserID:           userID,
		OrganizationName: organizationName,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    "cocotola-auth",
			Subject:   "AccessToken",
			Audience:  []string{"cocotola"},
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.tokenTimeout)),
		},
	}
	token := jwt.NewWithClaims(m.signingMethod, claims)
	signed, err := token.SignedString(m.signingKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

// ParseAccessToken validates a JWT string and returns the embedded user info and JTI.
func (m *JWTManager) ParseAccessToken(tokenString string) (*authservice.UserInfo, string, error) {
	claims, err := m.parseToken(tokenString)
	if err != nil {
		return nil, "", fmt.Errorf("parse token: %w", err)
	}

	userInfo, err := authservice.NewUserInfo(claims.UserID, claims.LoginID, claims.OrganizationName, claims.ExpiresAt.Time)
	if err != nil {
		return nil, "", fmt.Errorf("create user info: %w", err)
	}

	return userInfo, claims.ID, nil
}

func (m *JWTManager) parseToken(tokenString string) (*userClaims, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.signingKey, nil
	}

	currentToken, err := jwt.ParseWithClaims(tokenString, &userClaims{
		LoginID:          "",
		UserID:           0,
		OrganizationName: "",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "",
			Subject:   "",
			Audience:  nil,
			ExpiresAt: nil,
			NotBefore: nil,
			IssuedAt:  nil,
			ID:        "",
		},
	}, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !currentToken.Valid {
		return nil, errors.New("invalid token")
	}

	currentClaims, ok := currentToken.Claims.(*userClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return currentClaims, nil
}
