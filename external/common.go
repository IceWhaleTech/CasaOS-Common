package external

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"
)

func getAddress(addressFile string) (string, error) {
	buf, err := os.ReadFile(addressFile)
	if err != nil {
		return "", err
	}

	address := string(buf)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address+"/ping", nil)
	if err != nil {
		return "", err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", errors.New("failed to ping the service as address " + address)
	}

	return address, nil
}
