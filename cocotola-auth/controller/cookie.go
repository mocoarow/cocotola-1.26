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
	maxAge := tokenTTLMin * 60      //nolint:mnd // seconds per minute
	http.SetCookie(w, &http.Cookie{ //nolint:exhaustruct
		Name:     c.Name,
		Value:    token,
		Path:     c.Path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: parseSameSite(c.SameSite),
	})
}

// ClearTokenCookie removes the session token cookie by setting MaxAge to -1.
func (c *CookieConfig) ClearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{ //nolint:exhaustruct
		Name:     c.Name,
		Value:    "",
		Path:     c.Path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: parseSameSite(c.SameSite),
	})
}

func parseSameSite(s string) http.SameSite {
	switch s {
	case "Strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}
