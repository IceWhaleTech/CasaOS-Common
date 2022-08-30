package version

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/ini.v1"

	"github.com/IceWhaleTech/CasaOS-Common/utils/file"
	_ "github.com/mattn/go-sqlite3" // nolint
)

const (
	LegacyCasaOSServiceName = "casaos.service"
	configKeyUniqueToZero3x = "USBAutoMount"
	configKeyDBPath         = "DBPath"
)

var (
	_configFile        *ini.File
	_casaOSBinFilePath string
)

var ErrLegacyVersionNotFound = errors.New("legacy version not found")

func init() {
	serviceFilePath := file.FindFirstFile("/etc/systemd", LegacyCasaOSServiceName)
	if serviceFilePath == "" {
		return
	}

	serviceFile, err := ini.Load(serviceFilePath)
	if err != nil {
		return
	}

	section, err := serviceFile.GetSection("Service")
	if err != nil {
		return
	}

	key, err := section.GetKey("ExecStart")
	if err != nil {
		return
	}

	execStart := key.Value()
	texts := strings.Split(execStart, " ")

	_casaOSBinFilePath = texts[0]

	if _, err := os.Stat(_casaOSBinFilePath); os.IsNotExist(err) {
		_casaOSBinFilePath, err = exec.LookPath("casaos")

		if err != nil {
			return
		}
	}

	var configFilePath string
	for i, text := range texts {
		if text == "-c" {
			configFilePath = texts[i+1]
			break
		}
	}

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return
	}

	_configFile, _ = ini.Load(configFilePath)
}

func DetectLegacyVersion() (int, int, int, error) {
	if _configFile == nil {
		return -1, -1, -1, ErrLegacyVersionNotFound
	}

	cmd := exec.Command(_casaOSBinFilePath, "-v")
	versionBytes, err := cmd.Output()
	if err != nil {
		minorVersion, err := DetectMinorVersion()
		if err != nil {
			return -1, -1, -1, err
		}

		if minorVersion == 2 {
			return 0, 2, 99, nil // 99 means we don't know the patch version.
		}

		configKeyDBPathExist, err := IsConfigKeyDBPathExist()
		if err != nil {
			return -1, -1, -1, err
		}

		if !configKeyDBPathExist {
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
func IsConfigKeyDBPathExist() (bool, error) {
	if _configFile == nil {
		return false, ErrLegacyVersionNotFound
	}

	if !_configFile.Section("app").HasKey(configKeyDBPath) {
		return false, nil
	}

	return true, nil
}
