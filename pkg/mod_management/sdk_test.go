package modmanagement_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"

	codegen "github.com/IceWhaleTech/CasaOS-Common/codegen/mod_management"
	"github.com/IceWhaleTech/CasaOS-Common/external"
	modmanagement "github.com/IceWhaleTech/CasaOS-Common/pkg/mod_management"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallableModules(t *testing.T) {
	_, port := newLocalhostServerForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/mod_management/management/modules" {
			http.NotFound(w, r)
			return
		}

		writeJSONForTest(t, w, http.StatusOK, map[string]any{
			"data": []map[string]string{
				{
					"name":  "doconverter",
					"title": "Doconverter",
				},
			},
		})
	}))

	client := newModManagementClientForTest(t, port)
	modules, err := client.InstallableModules()
	require.NoError(t, err)

	require.Len(t, modules, 1)
	assert.Equal(t, "doconverter", modules[0].Name)
}

func TestInstallableModulesReturnsNoDataForUnexpectedOKResponse(t *testing.T) {
	_, port := newLocalhostServerForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/mod_management/management/modules" {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	client := newModManagementClientForTest(t, port)
	modules, err := client.InstallableModules()

	assert.ErrorIs(t, err, modmanagement.ErrNoDataInResponse)
	assert.Empty(t, modules)
}

func TestInstallModule(t *testing.T) {
	var installRequests atomic.Int32
	var port int
	server, serverPort := newLocalhostServerForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handleManagementRouteForTest(t, w, r, port) {
			return
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/mod_management/modules":
			writeJSONForTest(t, w, http.StatusOK, map[string]any{
				"data": []map[string]string{},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/mod_management/management/modules":
			installRequests.Add(1)
			writeJSONForTest(t, w, http.StatusOK, map[string]string{
				"message": "ok",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	port = serverPort

	runtimePath := t.TempDir()
	writeManagementAddressForTest(t, runtimePath, server.URL)

	err := modmanagement.RequireModule("doconverter", runtimePath)
	require.NoError(t, err)
	assert.Equal(t, int32(1), installRequests.Load())
}

func TestInstallNoExistModule(t *testing.T) {
	var port int
	server, serverPort := newLocalhostServerForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handleManagementRouteForTest(t, w, r, port) {
			return
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/mod_management/modules":
			writeJSONForTest(t, w, http.StatusOK, map[string]any{
				"data": []map[string]string{},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v2/mod_management/management/modules":
			http.Error(w, "module not exist", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	port = serverPort

	runtimePath := t.TempDir()
	writeManagementAddressForTest(t, runtimePath, server.URL)

	err := modmanagement.RequireModule("abc", runtimePath)
	assert.ErrorIs(t, err, modmanagement.ErrModuleNoInStore)
}

func newModManagementClientForTest(t *testing.T, port int) *modmanagement.ModManagementClient {
	t.Helper()

	client, err := codegen.NewClientWithResponses(fmt.Sprintf("http://localhost:%d/v2/mod_management", port))
	require.NoError(t, err)

	return &modmanagement.ModManagementClient{
		Client: client,
	}
}

func newLocalhostServerForTest(t *testing.T, handler http.Handler) (*httptest.Server, int) {
	t.Helper()

	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	_, portString, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	port, err := strconv.Atoi(portString)
	require.NoError(t, err)

	return server, port
}

func handleManagementRouteForTest(t *testing.T, w http.ResponseWriter, r *http.Request, port int) bool {
	t.Helper()

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/ping":
		w.WriteHeader(http.StatusOK)
		return true
	case r.Method == http.MethodGet && r.URL.Path == "/v1/gateway/port":
		writeJSONForTest(t, w, http.StatusOK, map[string]int{
			"data": port,
		})
		return true
	}

	return false
}

func writeManagementAddressForTest(t *testing.T, runtimePath string, address string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(runtimePath, external.ManagementURLFilename), []byte(address), 0o644)
	require.NoError(t, err)
}

func writeJSONForTest(t *testing.T, w http.ResponseWriter, statusCode int, value any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Errorf("encode response: %v", err)
	}
}
