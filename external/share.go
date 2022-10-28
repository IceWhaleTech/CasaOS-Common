package external

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	APICasaOSShare = "/v1/samba/shares"
)

type ShareService interface {
	DeleteShare(id string) error
}
type shareService struct {
	addressFile string
}

func (n *shareService) DeleteShare(id string) error {
	address, err := getAddress(n.addressFile)
	if err != nil {
		return err
	}

	url := strings.TrimSuffix(address, "/") + APICasaOSShare + "/" + id
	fmt.Println(url)
	message := "{}"
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// Fetch Request
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.New("failed to send share (status code: " + fmt.Sprint(response.StatusCode) + ")")
	}
	return nil

}

func NewShareService(runtimePath string) ShareService {
	return &shareService{
		addressFile: filepath.Join(runtimePath, CasaOSURLFilename),
	}
}
