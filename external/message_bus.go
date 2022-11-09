package external

import (
	"path/filepath"
	"strings"
)

const (
	MessageBusAddressFilename = "message-bus.url"
	APIMessageBus             = "/v2/message_bus"
)

func GetMessageBusAddress(runtimePath string) (string, error) {
	address, err := getAddress(filepath.Join(runtimePath, MessageBusAddressFilename))
	if err != nil {
		return "", err
	}

	return strings.TrimRight(address, "/") + APIMessageBus, nil
}
