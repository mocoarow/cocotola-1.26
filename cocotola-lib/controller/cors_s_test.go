package controller_test

import (
	"errors"
	"testing"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

func Test_InitCORS_shouldReturnAllowAllOrigins_whenWildcardWithoutCredentials(t *testing.T) {
	t.Parallel()

	// given
	cfg := &controller.CORSConfig{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST",
		AllowHeaders:     "Content-Type",
		AllowCredentials: false,
	}

	// when
	result, err := controller.InitCORS(cfg)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.AllowAllOrigins {
		t.Fatal("expected AllowAllOrigins to be true")
	}
	if len(result.AllowOrigins) != 0 {
		t.Fatalf("expected empty AllowOrigins, got %v", result.AllowOrigins)
	}
}

func Test_InitCORS_shouldReturnError_whenWildcardWithCredentials(t *testing.T) {
	t.Parallel()

	// given
	cfg := &controller.CORSConfig{
		AllowOrigins:     "*",
		AllowMethods:     "GET",
		AllowHeaders:     "",
		AllowCredentials: true,
	}

	// when
	_, err := controller.InitCORS(cfg)

	// then
	if !errors.Is(err, controller.ErrCORSCredentialsWithWildcard) {
		t.Fatalf("expected ErrCORSCredentialsWithWildcard, got %v", err)
	}
}

func Test_InitCORS_shouldReturnSpecificOrigins_whenNotWildcard(t *testing.T) {
	t.Parallel()

	// given
	cfg := &controller.CORSConfig{
		AllowOrigins:     "http://localhost:3000,http://example.com",
		AllowMethods:     "GET,POST,PUT",
		AllowHeaders:     "Authorization,Content-Type",
		AllowCredentials: true,
	}

	// when
	result, err := controller.InitCORS(cfg)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AllowAllOrigins {
		t.Fatal("expected AllowAllOrigins to be false")
	}
	if len(result.AllowOrigins) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(result.AllowOrigins))
	}
	if result.AllowOrigins[0] != "http://localhost:3000" {
		t.Fatalf("expected first origin to be http://localhost:3000, got %s", result.AllowOrigins[0])
	}
	if result.AllowOrigins[1] != "http://example.com" {
		t.Fatalf("expected second origin to be http://example.com, got %s", result.AllowOrigins[1])
	}
	if !result.AllowCredentials {
		t.Fatal("expected AllowCredentials to be true")
	}
	if len(result.AllowMethods) != 3 {
		t.Fatalf("expected 3 methods, got %d", len(result.AllowMethods))
	}
	if len(result.AllowHeaders) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(result.AllowHeaders))
	}
}
