package version

import (
	"database/sql"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/ini.v1"

	_ "github.com/mattn/go-sqlite3" // nolint
)

const (
	LegacyCasaOSBinFilePath    = "/usr/bin/casaOS"
	LegacyCasaOSConfigFilePath = "/etc/casaos.conf"
	LegacyCasaOSServiceName    = "casaos.service"
	configKeyUniqueToZero3x    = "USBAutoMount"
	configKeyDBPath            = "DBPath"
)

var _configFile *ini.File

var ErrLegacyVersionNotFound = errors.New("legacy version not found")

func init() {
	if _, err := os.Stat(LegacyCasaOSConfigFilePath); os.IsNotExist(err) {
		return
	}

	_file, err := ini.Load(LegacyCasaOSConfigFilePath)
	if err != nil {
		return
	}

	_configFile = _file
}

func DetectLegacyVersion() (int, int, int, error) {
	if _configFile == nil {
		return -1, -1, -1, ErrLegacyVersionNotFound
	}

	binPath := LegacyCasaOSBinFilePath

	if _, err := os.Stat(LegacyCasaOSBinFilePath); os.IsNotExist(err) {
		path, err := exec.LookPath("casaos")
		if err != nil {
			return -1, -1, -1, ErrLegacyVersionNotFound
		}
		binPath = path
	}

	cmd := exec.Command(binPath, "-v")
	versionBytes, err := cmd.Output()
	if err != nil {
		minorVersion, err := DetectMinorVersion()
		if err != nil {
			return -1, -1, -1, err
		}

		if minorVersion == 2 {
			return 0, 2, 99, nil // 99 means we don't know the patch version.
		}

		isUserDataInDatabase, err := IsUserDataInDatabase()
		if err != nil {
			return -1, -1, -1, err
		}

		if !isUserDataInDatabase {
			return 0, 3, 0, nil // it could be 0.3.0, 0.3.1 or 0.3.2 but only one version can be returned.
		}

		return 0, 3, 3, nil // it could be 0.3.3 or 0.3.4 but only one version can be returned.
	}

	versionString := string(versionBytes[:5])

	if versionString == "0.3.5" {
		return 0, 3, 5, nil
	}

	return -1, -1, -1, ErrLegacyVersionNotFound
}

// Detect minor version of CasaOS. It returns 2 for "0.2.x" or 3 for "0.3.x"
//
// (This is often useful when failing to get version from API because CasaOS is not running.)
func DetectMinorVersion() (int, error) {
	if _configFile == nil {
		return -1, ErrLegacyVersionNotFound
	}

	if _configFile.Section("server").HasKey(configKeyUniqueToZero3x) {
		return 3, nil
	}

	return 2, nil
}

// Check if user data is stored in database (0.3.3+)
func IsUserDataInDatabase() (bool, error) {
	if _configFile == nil {
		return false, ErrLegacyVersionNotFound
	}

	if !_configFile.Section("app").HasKey(configKeyDBPath) {
		return false, nil
	}

	dbPath := _configFile.Section("app").Key(configKeyDBPath).String()

	dbFile := filepath.Join(dbPath, "db", "casaOS.db")

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false, nil
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return false, err
	}

	defer db.Close()

	sqlStatement := "SELECT name FROM sqlite_master WHERE type='table' AND name='o_users'"

	rows, err := db.Query(sqlStatement)
	if err != nil {
		return false, err
	}

	defer rows.Close()

	return true, nil
}
