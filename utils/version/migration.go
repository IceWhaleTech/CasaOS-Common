package version

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type GlobalMigrationStatus struct {
	ServiceName         string
	LastMigratedVersion string
}

var (
	GlobalMigrationStatusDirPath = "/var/lib/casaos/migration"

	ErrInvalidVersion     = errors.New("version should start with 'v'")
	ErrInvalidServiceName = errors.New("service name should not contain space or upper case letter")
)

func (m *GlobalMigrationStatus) Done(version string) error {
	// error if version does not start with 'v'
	if !strings.HasPrefix(version, "v") {
		return ErrInvalidVersion
	}

	m.LastMigratedVersion = version

	// create runtimePath if not exists
	if _, err := os.Stat(GlobalMigrationStatusDirPath); os.IsNotExist(err) {
		os.MkdirAll(GlobalMigrationStatusDirPath, 0o755)
	}

	// save m.LastMigratedVersion to filepath
	filepath := m.GetGlobalMigrationStatusFilePath()
	return os.WriteFile(filepath, []byte(m.LastMigratedVersion), 0o644)
}

func GetGlobalMigrationStatus(serviceName string) (*GlobalMigrationStatus, error) {
	if err := validateServiceName(serviceName); err != nil {
		return nil, err
	}

	m := &GlobalMigrationStatus{
		ServiceName:         serviceName,
		LastMigratedVersion: "",
	}

	filepath := m.GetGlobalMigrationStatusFilePath()

	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		// read string from filepath
		buf, err := os.ReadFile(filepath)
		if err != nil {
			return nil, err
		}

		m.LastMigratedVersion = strings.TrimSpace(string(buf))
	}

	return m, nil
}

func (m *GlobalMigrationStatus) GetGlobalMigrationStatusFilePath() string {
	return filepath.Join(GlobalMigrationStatusDirPath, m.ServiceName+".status")
}

func validateServiceName(serviceName string) error {
	// should not contain space
	if strings.Contains(serviceName, " ") {
		return ErrInvalidServiceName
	}

	// should be all lower case
	if strings.ToLower(serviceName) != serviceName {
		return ErrInvalidServiceName
	}

	return nil
}
