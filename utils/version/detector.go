package version

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"

	_ "github.com/mattn/go-sqlite3" // nolint
)

const (
	LegacyCasaOSConfigFilePath = "/etc/casaos.conf"
	configKeyUniqueToZero3x    = "USBAutoMount"
	configKeyDBPath            = "DBPath"
)

var _configFile *ini.File

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

// Detect minor version of CasaOS. It returns 2 for "0.2.x" or 3 for "0.3.x"
//
// (This is often useful when failing to get version from API because CasaOS is not running.)
func DetectMinorVersion() (int, error) {
	if _configFile == nil {
		return -1, errors.New("config file not found")
	}

	if _configFile.Section("server").HasKey(configKeyUniqueToZero3x) {
		return 3, nil
	}

	return 2, nil
}

// Check if user data is stored in database (true) or in config file (false)
//
// (user data is stored in config file for 0.3.0-0.3.2)
func IsUserDataInDatabase() (bool, error) {
	if _configFile == nil {
		return false, errors.New("config file not found")
	}

	if !_configFile.Section("app").HasKey(configKeyDBPath) {
		return false, nil
	}

	dbPath := _configFile.Section("app").Key(configKeyDBPath).String()

	dbFile := filepath.Join(dbPath, "db", "casaOS.db")

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false, err
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

	for rows.Next() {
		return true, nil
	}

	return false, nil
}
