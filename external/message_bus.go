package external

import (
	"path/filepath"
	"strings"
)

const (
	MessageBusAddresFilename = "message_bus.url"
	APIMessageBus            = "/v2/message_bus"
)

func GetMessageBusAddress(runtimePath string) (string, error) {
	address, err := getAddress(filepath.Join(runtimePath, MessageBusAddresFilename))
	if err != nil {
		return "", err
	}

	return strings.TrimRight(address, "/") + APIMessageBus, nil
}
