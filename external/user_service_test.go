package external

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/orca-zhang/ecache"
)

func resetParseTokenCacheForTest(t *testing.T) {
	t.Helper()
	invalidParseTokenCache = ecache.NewLRUCache(2, 16, time.Minute)
	readUserServiceAddress = getAddress
	userServiceAddressFile = filepath.Join("/var/run/casaos", UserServiceAddressFilename)
	gatewaySockFile = filepath.Join("/var/run/casaos", GatewaySockFilename)
}

func TestParseTokenDoesNotCacheSuccessfulResponse(t *testing.T) {
	resetParseTokenCacheForTest(t)

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"success": 1,
			"message": "ok",
			"data": map[string]any{
				"valid":      true,
				"expires_at": time.Now().Add(time.Hour).Unix(),
				"username":   "cached",
				"role":       "user",
				"user_id":    3,
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	readUserServiceAddress = func(addressFile string) (string, error) {
		return server.URL, nil
	}

	for i := 0; i < 2; i++ {
		parsed, err := ParseToken("successful-token")
		if err != nil {
			t.Fatalf("ParseToken returned error: %v", err)
		}
		if parsed == nil || parsed.Username != "cached" {
			t.Fatalf("expected cached response, got %+v", parsed)
		}
	}

	if got := atomic.LoadInt32(&requests); got != 2 {
		t.Fatalf("expected successful token to be checked twice, got %d requests", got)
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

func TestParseTokenCachesUnauthorizedResponse(t *testing.T) {
	resetParseTokenCacheForTest(t)

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	readUserServiceAddress = func(addressFile string) (string, error) {
		return server.URL, nil
	}

	for i := 0; i < 2; i++ {
		parsed, err := ParseToken("deleted-user-token")
		if i == 0 {
			if err == nil {
				t.Fatal("expected first ParseToken call to fail")
			}
		} else if !errors.Is(err, errTokenInvalid) {
			t.Fatalf("expected cached errTokenInvalid, got %v", err)
		}
		if parsed != nil {
			t.Fatalf("expected nil parsed token, got %+v", parsed)
		}
	}

	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Fatalf("expected unauthorized token to be checked once, got %d requests", got)
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

	sockPath := shortUnixSocketPathForTest(t)
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

func shortUnixSocketPathForTest(t *testing.T) string {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("unix sockets are not supported on windows")
	}

	tempDir, err := os.MkdirTemp("/tmp", "co-")
	if err != nil {
		t.Fatalf("create short temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	return filepath.Join(tempDir, GatewaySockFilename)
}
