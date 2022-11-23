package time

import (
	"os"
	"path/filepath"
	"strings"
)

const LocalTimeFilePath = "/etc/localtime"

var systemTimeZoneName string

func GetSystemTimeZoneName() string {
	if systemTimeZoneName == "" {

		timeZoneFilePath, err := os.Readlink(LocalTimeFilePath)
		if err != nil {
			println("cannot read symbolic link from " + LocalTimeFilePath + ": " + err.Error())
			return ""
		}

		tokens := strings.Split(timeZoneFilePath, string(filepath.Separator))

		for {
			if len(tokens) == 0 {
				println("cannot get timezone name from timezone file path " + timeZoneFilePath)
				return ""
			}

			if tokens[0] == "zoneinfo" {
				break
			}

			tokens = tokens[1:]
		}

		if len(tokens) < 2 {
			println("cannot get timezone name from timezone file path " + timeZoneFilePath)
			return ""
		}

		systemTimeZoneName = strings.Join(tokens[1:], "/")
	}

	return systemTimeZoneName
}
