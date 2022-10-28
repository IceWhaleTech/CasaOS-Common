package external

import (
	"errors"
	"net/http"
	"os"
)

func getAddress(addressFile string) (string, error) {
	buf, err := os.ReadFile(addressFile)
	if err != nil {
		return "", err
	}

	address := string(buf)

	response, err := http.Get(address + "/ping")
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", errors.New("failed to ping the service as address " + address)
	}

	return address, nil
}
