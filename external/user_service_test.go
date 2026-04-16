package external

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/orca-zhang/ecache"
)

func resetParseTokenCacheForTest(t *testing.T) {
	t.Helper()
	validParseTokenCache = ecache.NewLRUCache(2, 8, time.Minute)
	invalidParseTokenCache = ecache.NewLRUCache(2, 16, time.Minute)
}

func writeUserServiceAddressFile(t *testing.T, runtimePath, address string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(runtimePath, UserServiceAddressFilename), []byte(address), 0o644); err != nil {
		t.Fatalf("write address file: %v", err)
	}
}

func TestParseTokenRejectsExpiredTokenFromService(t *testing.T) {
	resetParseTokenCacheForTest(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/parse-token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"success": 1,
			"message": "ok",
			"data": map[string]any{
				"valid":      true,
				"expires_at": time.Now().Add(-time.Minute).Unix(),
				"username":   "alice",
				"role":       "admin",
				"id":         1,
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	runtimePath := t.TempDir()
	writeUserServiceAddressFile(t, runtimePath, server.URL)

	parsed, err := ParseToken(runtimePath, "expired-token")
	if err == nil {
		t.Fatalf("expected error for expired token, got nil")
	}
	if parsed != nil {
		t.Fatalf("expected nil parsed token, got %+v", parsed)
	}
}

func TestParseTokenDoesNotReturnExpiredCachedToken(t *testing.T) {
	resetParseTokenCacheForTest(t)

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"success": 1,
			"message": "ok",
			"data": map[string]any{
				"valid":      true,
				"expires_at": time.Now().Add(time.Hour).Unix(),
				"username":   "bob",
				"role":       "user",
				"id":         2,
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	runtimePath := t.TempDir()
	writeUserServiceAddressFile(t, runtimePath, server.URL)

	const token = "cached-token"
	validParseTokenCache.Put(token, &ParsedToken{
		Valid:     true,
		ExpiresAt: time.Now().Add(-time.Minute).Unix(),
		Username:  "stale",
		Role:      "user",
		ID:        3,
	})

	parsed, err := ParseToken(runtimePath, token)
	if !errors.Is(err, errTokenExpired) {
		t.Fatalf("expected errTokenExpired, got %v", err)
	}
	if parsed != nil {
		t.Fatalf("expected nil parsed token, got %+v", parsed)
	}
	if requestCount != 0 {
		t.Fatalf("expected 0 service requests, got %d", requestCount)
	}
}

func TestParseTokenDoesNotFallbackForInvalidCachedSentinel(t *testing.T) {
	resetParseTokenCacheForTest(t)

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	runtimePath := t.TempDir()
	writeUserServiceAddressFile(t, runtimePath, server.URL)

	const token = "invalid-token"
	invalidParseTokenCache.Put(token, tokenCacheSentinelInvalid)

	parsed, err := ParseToken(runtimePath, token)
	if !errors.Is(err, errTokenInvalid) {
		t.Fatalf("expected errTokenInvalid, got %v", err)
	}
	if parsed != nil {
		t.Fatalf("expected nil parsed token, got %+v", parsed)
	}
	if requestCount != 0 {
		t.Fatalf("expected 0 service requests, got %d", requestCount)
	}
}

func TestParseTokenCachesInvalidTokenSentinel(t *testing.T) {
	resetParseTokenCacheForTest(t)

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"success": 1,
			"message": "ok",
			"data": map[string]any{
				"valid":      false,
				"expires_at": time.Now().Add(time.Hour).Unix(),
				"username":   "",
				"role":       "",
				"id":         0,
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	runtimePath := t.TempDir()
	writeUserServiceAddressFile(t, runtimePath, server.URL)

	const token = "invalid-from-service"

	parsed, err := ParseToken(runtimePath, token)
	if !errors.Is(err, errTokenInvalid) {
		t.Fatalf("expected errTokenInvalid, got %v", err)
	}
	if parsed != nil {
		t.Fatalf("expected nil parsed token, got %+v", parsed)
	}

	parsed, err = ParseToken(runtimePath, token)
	if !errors.Is(err, errTokenInvalid) {
		t.Fatalf("expected cached errTokenInvalid, got %v", err)
	}
	if parsed != nil {
		t.Fatalf("expected nil parsed token on cache hit, got %+v", parsed)
	}
	if requestCount != 1 {
		t.Fatalf("expected 1 service request, got %d", requestCount)
	}
}
