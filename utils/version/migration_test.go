package version

import (
	"os"
	"testing"
)

func setup() func() {
	GlobalMigrationStatusDirPath, _ = os.MkdirTemp("", "migration-test")

	return func() {
		os.RemoveAll(GlobalMigrationStatusDirPath)
	}
}

func TestMigrationVersioning(t *testing.T) {
	defer setup()()

	m, err := GetGlobalMigrationStatus("test")
	if err != nil {
		t.Error(err)
	}

	if err := m.Done("v1"); err != nil {
		t.Error(err)
	}

	if m.LastMigratedVersion != "v1" {
		t.Errorf("expected v1, got %s", m.LastMigratedVersion)
	}

	if err := m.Done("v2"); err != nil {
		t.Error(err)
	}

	if m.LastMigratedVersion != "v2" {
		t.Errorf("expected v2, got %s", m.LastMigratedVersion)
	}

	m, err = GetGlobalMigrationStatus("test")
	if err != nil {
		t.Error(err)
	}

	if m.LastMigratedVersion != "v2" {
		t.Errorf("expected v2, got %s", m.LastMigratedVersion)
	}
}
