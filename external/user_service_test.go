package external

import (
	"encoding/json"
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
	parseTokenCache = ecache.NewLRUCache(8, 32, 30*time.Second)
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
	parseTokenCache.Put(token, &ParsedToken{
		Valid:     true,
		ExpiresAt: time.Now().Add(-time.Minute).Unix(),
		Username:  "stale",
		Role:      "user",
		ID:        3,
	})

	parsed, err := ParseToken(runtimePath, token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed token, got nil")
	}
	if parsed.Username != "bob" {
		t.Fatalf("expected refreshed token from service, got %+v", parsed)
	}
	if requestCount != 1 {
		t.Fatalf("expected 1 service request, got %d", requestCount)
	}
}
