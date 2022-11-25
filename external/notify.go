package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/model/notify"
	http2 "github.com/IceWhaleTech/CasaOS-Common/utils/http"
)

const (
	CasaOSURLFilename = "casaos.url"
	APICasaOSNotify   = "/v1/notify"
)

type NotifyService interface {
	SendNotify(path string, message interface{}) error
	SendSystemStatusNotify(message map[string]interface{}) error
	SendInstallAppBySocket(app notify.Application) error
	SendUninstallAppBySocket(app notify.Application) error
}
type notifyService struct {
	addressFile string
}

func (n *notifyService) SendNotify(path string, message interface{}) error {
	address, err := getAddress(n.addressFile)
	if err != nil {
		return err
	}

	url := strings.TrimSuffix(address, "/") + APICasaOSNotify + "/" + path

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	response, err := http2.Post(url, body, 5*time.Second)
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
	return n.SendNotify("system_status", message)
}

func (n *notifyService) SendInstallAppBySocket(app notify.Application) error {
	return n.SendNotify("install_app", app)
}

func (n *notifyService) SendUninstallAppBySocket(app notify.Application) error {
	return n.SendNotify("uninstall_app", app)
}

func NewNotifyService(runtimePath string) NotifyService {
	return &notifyService{
		addressFile: filepath.Join(runtimePath, CasaOSURLFilename),
	}
}
