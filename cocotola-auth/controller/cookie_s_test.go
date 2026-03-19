package controller_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
)

func Test_CookieConfig_SetTokenCookie_shouldSetCorrectAttributes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		cookieConfig *controller.CookieConfig
		token        string
		tokenTTLMin  int
		wantName     string
		wantValue    string
		wantPath     string
		wantMaxAge   int
		wantSecure   bool
		wantSameSite http.SameSite
	}{
		{
			name: "Lax SameSite with Secure",
			cookieConfig: &controller.CookieConfig{
				Name:     "access_token",
				Path:     "/",
				Secure:   true,
				SameSite: "Lax",
			},
			token:        "test-token",
			tokenTTLMin:  60,
			wantName:     "access_token",
			wantValue:    "test-token",
			wantPath:     "/",
			wantMaxAge:   3600,
			wantSecure:   true,
			wantSameSite: http.SameSiteLaxMode,
		},
		{
			name: "Strict SameSite without Secure",
			cookieConfig: &controller.CookieConfig{
				Name:     "token",
				Path:     "/api",
				Secure:   false,
				SameSite: "Strict",
			},
			token:        "another-token",
			tokenTTLMin:  30,
			wantName:     "token",
			wantValue:    "another-token",
			wantPath:     "/api",
			wantMaxAge:   1800,
			wantSecure:   false,
			wantSameSite: http.SameSiteStrictMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			w := httptest.NewRecorder()

			// when
			tt.cookieConfig.SetTokenCookie(w, tt.token, tt.tokenTTLMin)

			// then
			cookies := w.Result().Cookies()
			require.Len(t, cookies, 1)
			cookie := cookies[0]
			assert.Equal(t, tt.wantName, cookie.Name)
			assert.Equal(t, tt.wantValue, cookie.Value)
			assert.Equal(t, tt.wantPath, cookie.Path)
			assert.Equal(t, tt.wantMaxAge, cookie.MaxAge)
			assert.True(t, cookie.HttpOnly)
			assert.Equal(t, tt.wantSecure, cookie.Secure)
			assert.Equal(t, tt.wantSameSite, cookie.SameSite)
		})
	}
}
