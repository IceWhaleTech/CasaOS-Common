package zimaos_common

import (
	"io"
	"os"
)

func GetModel() string {
	src := "/sys/class/dmi/id/board_version"
	_, err := os.Stat(src)
	if os.IsNotExist(err) {
		return ""
	} else {
		file, err := os.Open(src)
		if err != nil {
			return ""
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return ""
		}
		return string(content)
	}
}
func GetSerialNumber() string {
	src := "/sys/class/dmi/id/board_version"
	_, err := os.Stat(src)
	if os.IsNotExist(err) {
		return ""
	} else {
		file, err := os.Open(src)
		if err != nil {
			return ""
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return ""
		}
		return string(content)
	}
}
