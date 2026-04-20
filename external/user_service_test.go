package external

import (
	"encoding/json"
	"errors"
	"net"
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
	readUserServiceAddress = getAddress
	userServiceAddressFile = filepath.Join("/var/run/casaos", UserServiceAddressFilename)
	gatewaySockFile = filepath.Join("/var/run/casaos", GatewaySockFilename)
}

func TestParseTokenDoesNotReturnExpiredCachedToken(t *testing.T) {
	resetParseTokenCacheForTest(t)

	const token = "cached-token"
	validParseTokenCache.Put(token, &ParsedToken{
		Valid:     true,
		ExpiresAt: time.Now().Add(-time.Minute).Unix(),
		Username:  "stale",
		Role:      "user",
		UserID:    3,
	})

	parsed, err := ParseToken(token)
	if !errors.Is(err, errTokenExpired) {
		t.Fatalf("expected errTokenExpired, got %v", err)
	}
	if parsed != nil {
		t.Fatalf("expected nil parsed token, got %+v", parsed)
	}
}

func TestParseTokenDoesNotFallbackForInvalidCachedSentinel(t *testing.T) {
	resetParseTokenCacheForTest(t)

	const token = "invalid-token"
	invalidParseTokenCache.Put(token, tokenCacheSentinelInvalid)

	parsed, err := ParseToken(token)
	if !errors.Is(err, errTokenInvalid) {
		t.Fatalf("expected errTokenInvalid, got %v", err)
	}
	if parsed != nil {
		t.Fatalf("expected nil parsed token, got %+v", parsed)
	}
}

func TestParseTokenPrefersUserServiceAddressFile(t *testing.T) {
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
				"expires_at": time.Now().Add(time.Hour).Unix(),
				"username":   "primary",
				"role":       "admin",
				"id":         1,
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	readUserServiceAddress = func(addressFile string) (string, error) {
		if addressFile != userServiceAddressFile {
			t.Fatalf("expected user service address file, got %s", addressFile)
		}
		return server.URL, nil
	}

	parsed, err := ParseToken("primary-token")
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if parsed == nil || parsed.Username != "primary" {
		t.Fatalf("expected primary response, got %+v", parsed)
	}
}

func TestParseTokenFallsBackToGatewaySockWhenUserServiceAddressFileMissing(t *testing.T) {
	resetParseTokenCacheForTest(t)

	tempDir := t.TempDir()
	sockPath := filepath.Join(tempDir, GatewaySockFilename)
	gatewaySockFile = sockPath

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("listen unix socket: %v", err)
	}
	defer listener.Close()

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/users/parse-token" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]any{
				"success": 1,
				"message": "ok",
				"data": map[string]any{
					"valid":      true,
					"expires_at": time.Now().Add(time.Hour).Unix(),
					"username":   "gateway",
					"role":       "admin",
					"id":         2,
				},
			}); err != nil {
				t.Fatalf("encode response: %v", err)
			}
		}),
	}
	defer server.Close()
	go func() {
		_ = server.Serve(listener)
	}()

	readUserServiceAddress = func(addressFile string) (string, error) {
		if addressFile != userServiceAddressFile {
			t.Fatalf("expected user service address file lookup, got %s", addressFile)
		}
		return "", os.ErrNotExist
	}

	parsed, err := ParseToken("gateway-token")
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if parsed == nil || parsed.Username != "gateway" {
		t.Fatalf("expected gateway response, got %+v", parsed)
	}
}
