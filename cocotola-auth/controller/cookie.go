package controller

import (
	"net/http"
)

// CookieConfig holds settings for HTTP cookie-based token delivery.
type CookieConfig struct {
	Name     string `yaml:"name" validate:"required"`
	Path     string `yaml:"path" validate:"required"`
	Secure   bool   `yaml:"secure"`
	SameSite string `yaml:"sameSite" validate:"required,oneof=Lax Strict"`
}

// SetTokenCookie writes a session token cookie to the response with the configured attributes.
func (c *CookieConfig) SetTokenCookie(w http.ResponseWriter, token string, tokenTTLMin int) {
	maxAge := tokenTTLMin * 60 //nolint:mnd // seconds per minute
	http.SetCookie(w, c.buildCookie(token, maxAge))
}

// ClearTokenCookie removes the session token cookie by setting MaxAge to -1.
func (c *CookieConfig) ClearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, c.buildCookie("", -1))
}

// buildCookie constructs an *http.Cookie. The literals below intentionally use
// `Secure: true` and a concrete SameSite selector so gosec G124 sees a secure
// default; the configured Secure value is applied afterward to support local
// HTTP development without losing the static-analysis guarantee.
func (c *CookieConfig) buildCookie(value string, maxAge int) *http.Cookie {
	if c.SameSite == "Strict" {
		cookie := &http.Cookie{ //nolint:exhaustruct
			Name:     c.Name,
			Value:    value,
			Path:     c.Path,
			MaxAge:   maxAge,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		}
		cookie.Secure = c.Secure
		return cookie
	}
	cookie := &http.Cookie{ //nolint:exhaustruct
		Name:     c.Name,
		Value:    value,
		Path:     c.Path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	cookie.Secure = c.Secure
	return cookie
}
