package external

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	http2 "github.com/IceWhaleTech/CasaOS-Common/utils/http"
)

const (
	ManageURLFilename = "app-management.url"
	APIComposeInfo    = "/v2/app_management/compose"
	APIComposeStatus  = "/v2/app_management/compose"
)

type AppManageService interface {
	GetAppInfo(storeId string) (string, error)
	PutAppStatus(storeId string, status string) (string, error)
}

type appManageService struct {
	address string
}

func (m *appManageService) GetAppInfo(storeId string) (string, error) {
	url := strings.TrimSuffix(m.address, "/") + "/" + strings.TrimPrefix(APIComposeInfo, "/"+storeId)

	response, err := http2.Get(url, 30*time.Second)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return "", errors.New("failed to create route (status code: " + fmt.Sprint(response.StatusCode) + ")")
	}
	str, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	return string(str), nil
}

func (m *appManageService) PutAppStatus(storeId string, status string) (string, error) {
	url := strings.TrimSuffix(m.address, "/") + "/" + strings.TrimPrefix(APIComposeStatus, "/"+storeId)

	body := []byte(status)
	response, err := http2.Put(url, body, 30*time.Second)
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return "", errors.New("failed to change status (status code: " + fmt.Sprint(response.StatusCode) + ")")
	}
	str, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	return string(str), nil
}

func NewAppManageService(RuntimePath string) (AppManageService, error) {
	managementAddressFile := filepath.Join(RuntimePath, ManagementURLFilename)

	retry := 10

	for retry > 0 {
		if _, err := os.Stat(managementAddressFile); err == nil {
			break
		}

		fmt.Printf("gateway management address file `%s` not found, retrying in 1 second...(%d)\n", managementAddressFile, retry)

		time.Sleep(1 * time.Second)

		retry--
	}

	address, err := getAddress(managementAddressFile)
	if err != nil {
		return nil, err
	}

	if err := ping(address, 30*time.Second); err != nil {
		return nil, err
	}

	return &appManageService{
		address: address,
	}, nil
}
