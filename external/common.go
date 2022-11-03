package external

import (
	"errors"
	"net/http"
	"os"
	"time"

	http2 "github.com/IceWhaleTech/CasaOS-Common/utils/http"
)

func getAddress(addressFile string) (string, error) {
	buf, err := os.ReadFile(addressFile)
	if err != nil {
		return "", err
	}
	address := string(buf)

	response, err := http2.Get(address+"/ping", 30*time.Second)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", errors.New("failed to ping the service as address " + address)
	}

	return address, nil
}
