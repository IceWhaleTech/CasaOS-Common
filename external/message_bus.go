package external

import (
	"path/filepath"
	"strings"
)

const (
	MessageBusAddresFilename = "message_bus.url"
	APIMessageBus            = "/v2/message_bus"
)

func GetMessageBusAddress(runtimePath string) string {
	address, err := getAddress(filepath.Join(runtimePath, MessageBusAddresFilename))
	if err != nil {
		return ""
	}

	return strings.TrimRight(address, "/") + APIMessageBus
}
