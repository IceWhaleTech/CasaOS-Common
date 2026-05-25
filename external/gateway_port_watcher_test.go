package external

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestReadGatewayPortConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), GatewayConfigFilename)
	writeGatewayPortConfigForTest(t, configPath, 80, 443, true)

	config, err := ReadGatewayPortConfig(configPath)
	if err != nil {
		t.Fatalf("ReadGatewayPortConfig returned error: %v", err)
	}

	if config.HTTPPort != 80 {
		t.Fatalf("expected HTTP port 80, got %d", config.HTTPPort)
	}
	if !config.HTTPEnabled {
		t.Fatal("expected HTTP to be enabled by default")
	}
	if config.HTTPSPort != 443 {
		t.Fatalf("expected HTTPS port 443, got %d", config.HTTPSPort)
	}
	if !config.HTTPSEnabled {
		t.Fatal("expected HTTPS to be enabled")
	}
}

func TestReadGatewayPortConfigWithHTTPDisabled(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), GatewayConfigFilename)
	writeGatewayPortConfigWithHTTPEnabledForTest(t, configPath, 80, false, 443, true)

	config, err := ReadGatewayPortConfig(configPath)
	if err != nil {
		t.Fatalf("ReadGatewayPortConfig returned error: %v", err)
	}

	if config.HTTPEnabled {
		t.Fatal("expected HTTP to be disabled")
	}
	if !config.HTTPSEnabled {
		t.Fatal("expected HTTPS to be enabled")
	}
}

func TestGatewayPortWatcherFallsBackToPolling(t *testing.T) {
	var watcherStarted sync.Once
	started := make(chan struct{})

	originalNewWatcher := newGatewayPortFileWatcher
	newGatewayPortFileWatcher = func() (*fsnotify.Watcher, error) {
		watcherStarted.Do(func() {
			close(started)
		})
		return nil, errors.New("fsnotify unavailable")
	}
	t.Cleanup(func() {
		newGatewayPortFileWatcher = originalNewWatcher
	})

	configPath := filepath.Join(t.TempDir(), GatewayConfigFilename)
	writeGatewayPortConfigForTest(t, configPath, 80, 443, true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	changes := make(chan GatewayPortConfig, 4)
	watcher := NewGatewayPortWatcher(
		func(config GatewayPortConfig) {
			changes <- config
		},
		WithGatewayPortConfigPath(configPath),
		WithGatewayPortPollInterval(10*time.Millisecond),
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- watcher.Watch(ctx)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("watcher did not start")
	}

	assertNoGatewayPortConfigForTest(t, changes)

	writeGatewayPortConfigForTest(t, configPath, 80, 443, true)
	assertNoGatewayPortConfigForTest(t, changes)

	writeGatewayPortConfigWithHTTPEnabledForTest(t, configPath, 8080, false, 8443, false)
	waitGatewayPortConfigForTest(t, changes, GatewayPortConfig{
		HTTPEnabled:  false,
		HTTPPort:     8080,
		HTTPSPort:    8443,
		HTTPSEnabled: false,
	})

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Watch returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop after context cancellation")
	}
}

func assertNoGatewayPortConfigForTest(t *testing.T, changes <-chan GatewayPortConfig) {
	t.Helper()

	select {
	case got := <-changes:
		t.Fatalf("expected no gateway port config change, got %+v", got)
	case <-time.After(50 * time.Millisecond):
	}
}

func writeGatewayPortConfigForTest(t *testing.T, configPath string, httpPort int, httpsPort int, httpsEnabled bool) {
	t.Helper()

	writeGatewayPortConfigContentForTest(t, configPath, fmt.Sprintf(`port        = %d`, httpPort), httpsPort, httpsEnabled)
}

func writeGatewayPortConfigWithHTTPEnabledForTest(t *testing.T, configPath string, httpPort int, httpEnabled bool, httpsPort int, httpsEnabled bool) {
	t.Helper()

	writeGatewayPortConfigContentForTest(t, configPath, fmt.Sprintf(`enabled     = %t
port        = %d`, httpEnabled, httpPort), httpsPort, httpsEnabled)
}

func writeGatewayPortConfigContentForTest(t *testing.T, configPath string, gatewaySection string, httpsPort int, httpsEnabled bool) {
	t.Helper()

	content := fmt.Sprintf(`[common]
runtimepath = /var/run/casaos

[gateway]
logfileext  = log
logpath     = /var/log/casaos
logsavename = gateway
%s

[ssl]
enabled   = %t
port      = %d
cert_path = /var/lib/casaos/ssl/zimaos.local.fullchain.pem
key_path  = /var/lib/casaos/ssl/zimaos.local.key.pem
domain    = zimaos.local
type      = local
`, gatewaySection, httpsEnabled, httpsPort)

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write gateway config: %v", err)
	}
}

func waitGatewayPortConfigForTest(t *testing.T, changes <-chan GatewayPortConfig, want GatewayPortConfig) {
	t.Helper()

	timeout := time.After(2 * time.Second)
	for {
		select {
		case got := <-changes:
			if got == want {
				return
			}
		case <-timeout:
			t.Fatalf("timed out waiting for gateway port config %+v", want)
		}
	}
}
