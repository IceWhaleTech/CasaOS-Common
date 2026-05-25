package external

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/constants"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/ini.v1"
)

const (
	// GatewayConfigFilename is the gateway config file name under /etc/casaos.
	GatewayConfigFilename = "gateway.ini"
	// DefaultGatewayPortWatcherPollInterval is used when fsnotify is unavailable.
	DefaultGatewayPortWatcherPollInterval = time.Second
)

var (
	// DefaultGatewayConfigPath is the default gateway config file path.
	DefaultGatewayConfigPath = filepath.Join(constants.DefaultConfigPath, GatewayConfigFilename)

	newGatewayPortFileWatcher = fsnotify.NewWatcher
)

var errGatewayPortUsePolling = errors.New("use polling for gateway port watcher")

// GatewayPortConfig contains gateway HTTP and HTTPS listen settings.
type GatewayPortConfig struct {
	HTTPEnabled  bool `json:"http_enabled"`
	HTTPPort     int  `json:"http_port"`
	HTTPSPort    int  `json:"https_port"`
	HTTPSEnabled bool `json:"https_enabled"`
}

// GatewayPortChangeCallback receives valid gateway port config changes.
type GatewayPortChangeCallback func(config GatewayPortConfig)

// GatewayPortWatcher watches gateway.ini with fsnotify first and polling as fallback.
type GatewayPortWatcher struct {
	configPath   string
	pollInterval time.Duration
	callback     GatewayPortChangeCallback
}

// GatewayPortWatcherOption configures GatewayPortWatcher.
type GatewayPortWatcherOption func(*GatewayPortWatcher)

// WithGatewayPortConfigPath overrides the gateway.ini path.
func WithGatewayPortConfigPath(configPath string) GatewayPortWatcherOption {
	return func(w *GatewayPortWatcher) {
		w.configPath = configPath
	}
}

// WithGatewayPortPollInterval overrides the fallback polling interval.
func WithGatewayPortPollInterval(interval time.Duration) GatewayPortWatcherOption {
	return func(w *GatewayPortWatcher) {
		w.pollInterval = interval
	}
}

// NewGatewayPortWatcher creates a gateway port watcher.
func NewGatewayPortWatcher(callback GatewayPortChangeCallback, opts ...GatewayPortWatcherOption) *GatewayPortWatcher {
	watcher := &GatewayPortWatcher{
		configPath:   DefaultGatewayConfigPath,
		pollInterval: DefaultGatewayPortWatcherPollInterval,
		callback:     callback,
	}

	for _, opt := range opts {
		opt(watcher)
	}

	return watcher
}

// ListenGatewayPortChanges watches gateway.ini until ctx is canceled.
func ListenGatewayPortChanges(ctx context.Context, callback GatewayPortChangeCallback, opts ...GatewayPortWatcherOption) error {
	return NewGatewayPortWatcher(callback, opts...).Watch(ctx)
}

// ReadGatewayPortConfig reads HTTP and HTTPS port settings from gateway.ini.
func ReadGatewayPortConfig(configPath string) (GatewayPortConfig, error) {
	if configPath == "" {
		configPath = DefaultGatewayConfigPath
	}

	cfg, err := ini.Load(configPath)
	if err != nil {
		return GatewayPortConfig{}, err
	}

	gatewaySection := cfg.Section("gateway")
	httpPort, err := gatewaySection.Key("port").Int()
	if err != nil {
		return GatewayPortConfig{}, fmt.Errorf("read gateway port: %w", err)
	}
	if err := validateGatewayPort("gateway port", httpPort); err != nil {
		return GatewayPortConfig{}, err
	}

	httpEnabled := true
	if gatewaySection.HasKey("enabled") {
		httpEnabled, err = gatewaySection.Key("enabled").Bool()
		if err != nil {
			return GatewayPortConfig{}, fmt.Errorf("read gateway enabled: %w", err)
		}
	}

	sslSection := cfg.Section("ssl")
	httpsPort, err := sslSection.Key("port").Int()
	if err != nil {
		return GatewayPortConfig{}, fmt.Errorf("read ssl port: %w", err)
	}
	if err := validateGatewayPort("ssl port", httpsPort); err != nil {
		return GatewayPortConfig{}, err
	}

	httpsEnabled, err := sslSection.Key("enabled").Bool()
	if err != nil {
		return GatewayPortConfig{}, fmt.Errorf("read ssl enabled: %w", err)
	}

	return GatewayPortConfig{
		HTTPEnabled:  httpEnabled,
		HTTPPort:     httpPort,
		HTTPSPort:    httpsPort,
		HTTPSEnabled: httpsEnabled,
	}, nil
}

// Watch blocks until ctx is canceled. It uses fsnotify first, then falls back to polling by mtime and size.
func (w *GatewayPortWatcher) Watch(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if w.callback == nil {
		return errors.New("gateway port change callback is nil")
	}
	if w.configPath == "" {
		w.configPath = DefaultGatewayConfigPath
	}
	if w.pollInterval <= 0 {
		w.pollInterval = DefaultGatewayPortWatcherPollInterval
	}

	state := &gatewayPortWatchState{}
	_ = w.refresh(state, true)

	fileWatcher, err := newGatewayPortFileWatcher()
	if err == nil {
		if err = fileWatcher.Add(filepath.Dir(w.configPath)); err == nil {
			err = w.watchWithFSNotify(ctx, fileWatcher, state)
			_ = fileWatcher.Close()
			if !errors.Is(err, errGatewayPortUsePolling) {
				return err
			}
		} else {
			_ = fileWatcher.Close()
		}
	}

	return w.watchWithPolling(ctx, state)
}

type gatewayPortWatchState struct {
	metadata    gatewayPortConfigMetadata
	hasMetadata bool
	config      GatewayPortConfig
	hasConfig   bool
}

type gatewayPortConfigMetadata struct {
	modTime time.Time
	size    int64
}

func (m gatewayPortConfigMetadata) equal(other gatewayPortConfigMetadata) bool {
	return m.size == other.size && m.modTime.Equal(other.modTime)
}

func (w *GatewayPortWatcher) watchWithFSNotify(ctx context.Context, fileWatcher *fsnotify.Watcher, state *gatewayPortWatchState) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-fileWatcher.Events:
			if !ok {
				return errGatewayPortUsePolling
			}
			if !isGatewayPortConfigEvent(event, w.configPath) {
				continue
			}
			_ = w.refresh(state, true)
		case _, ok := <-fileWatcher.Errors:
			if !ok {
				return errGatewayPortUsePolling
			}
			return errGatewayPortUsePolling
		}
	}
}

func (w *GatewayPortWatcher) watchWithPolling(ctx context.Context, state *gatewayPortWatchState) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			_ = w.refresh(state, false)
		}
	}
}

func (w *GatewayPortWatcher) refresh(state *gatewayPortWatchState, forceRead bool) error {
	metadata, err := readGatewayPortConfigMetadata(w.configPath)
	if err != nil {
		state.hasMetadata = false
		return err
	}

	if !forceRead && state.hasMetadata && metadata.equal(state.metadata) {
		return nil
	}

	config, err := ReadGatewayPortConfig(w.configPath)
	if err != nil {
		state.metadata = metadata
		state.hasMetadata = true
		return err
	}

	state.metadata = metadata
	state.hasMetadata = true

	if !state.hasConfig {
		state.config = config
		state.hasConfig = true
		return nil
	}

	if state.config == config {
		return nil
	}

	state.config = config
	w.callback(config)

	return nil
}

func readGatewayPortConfigMetadata(configPath string) (gatewayPortConfigMetadata, error) {
	info, err := os.Stat(configPath)
	if err != nil {
		return gatewayPortConfigMetadata{}, err
	}
	if info.IsDir() {
		return gatewayPortConfigMetadata{}, fmt.Errorf("%s is a directory", configPath)
	}

	return gatewayPortConfigMetadata{
		modTime: info.ModTime(),
		size:    info.Size(),
	}, nil
}

func isGatewayPortConfigEvent(event fsnotify.Event, configPath string) bool {
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove|fsnotify.Chmod) == 0 {
		return false
	}

	return filepath.Clean(event.Name) == filepath.Clean(configPath)
}

func validateGatewayPort(name string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s %d is out of range", name, port)
	}

	return nil
}
