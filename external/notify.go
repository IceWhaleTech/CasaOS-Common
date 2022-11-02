package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const (
	CasaOSURLFilename = "casaos.url"
	APICasaOSNotify   = "/v1/notify"
)

type NotifyService interface {
	SendNotify(path string, message map[string]interface{}) error
	SendSystemStatusNotify(message map[string]interface{}) error
}
type notifyService struct {
	addressFile string
	httpClient  *http.Client
}

func (n *notifyService) SendNotify(path string, message map[string]interface{}) error {
	address, err := getAddress(n.addressFile)
	if err != nil {
		return err
	}

	url := strings.TrimSuffix(address, "/") + APICasaOSNotify + "/" + path
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	response, err := n.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.New("failed to send notify (status code: " + fmt.Sprint(response.StatusCode) + ")")
	}
	return nil
}

// disk: "sys_disk":{"size":56866869248,"avail":5855485952,"health":true,"used":48099700736}
// usb:   "sys_usb":[{"name": "sdc","size": 7747397632,"model": "DataTraveler_2.0","avail": 7714418688,"children": null}]
func (n *notifyService) SendSystemStatusNotify(message map[string]interface{}) error {
	address, err := getAddress(n.addressFile)
	if err != nil {
		return err
	}

	url := strings.TrimSuffix(address, "/") + APICasaOSNotify + "/system_status"

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	response, err := n.httpClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return errors.New("failed to send notify (status code: " + fmt.Sprint(response.StatusCode) + ")")
	}

	response.Body.Close()

	return nil
}

func NewNotifyService(runtimePath string) NotifyService {
	return &notifyService{
		addressFile: filepath.Join(runtimePath, CasaOSURLFilename),
		httpClient:  &http.Client{},
	}
}
