package modmanagement

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/IceWhaleTech/CasaOS-Common/codegen/mod_management"
	"github.com/IceWhaleTech/CasaOS-Common/external"
)

var ErrNoDataInResponse = fmt.Errorf("no data in response")

type ModManagementClient struct {
	Client *mod_management.ClientWithResponses
}

type ModManagementClientOpts struct {
	Port *int
}

func NewClient(opts ModManagementClientOpts) (*ModManagementClient, error) {
	port := 80
	if opts.Port != nil {
		port = *opts.Port
	}
	client, err := mod_management.NewClientWithResponses(fmt.Sprintf("http://localhost:%d/v2/mod_management", port))
	if err != nil {
		return nil, err
	}
	return &ModManagementClient{Client: client}, nil
}

func (c *ModManagementClient) InstalledModules() ([]mod_management.Module, error) {
	resp, err := c.Client.ModuleListWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get installed modules: %s", resp.Status())
	}

	if resp.JSON200.Data == nil {
		return []mod_management.Module{}, ErrNoDataInResponse
	}
	return *resp.JSON200.Data, nil
}

func (c *ModManagementClient) InstallableModules() ([]mod_management.RemoteModule, error) {
	resp, err := c.Client.InstallableModuleListWithResponse(context.Background())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get installable modules: %s", resp.Status())
	}

	if resp.JSON200.Data == nil {
		return []mod_management.RemoteModule{}, ErrNoDataInResponse
	}

	return *resp.JSON200.Data, nil
}

func (c *ModManagementClient) InstallModule(name string) error {
	resp, err := c.Client.ModuleInstallWithResponse(context.Background(), mod_management.ModuleInstallJSONRequestBody{
		Name: name,
	})
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to get installable modules: %s", resp.Status())
	}
	return nil
}
func (c *ModManagementClient) InstallModuleAsync(name string) error {
	resp, err := c.Client.ModuleInstallAsyncWithResponse(context.Background(), mod_management.ModuleInstallJSONRequestBody{
		Name: name,
	})
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to get installable modules: %s", resp.Status())
	}
	return nil
}

func (c *ModManagementClient) UninstallModule(name string) error {
	resp, err := c.Client.ModuleUninstallWithResponse(context.Background(), mod_management.ModuleId{
		Name: name,
	})
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to uninstall module: %s", resp.Status())
	}
	return nil
}

func RequireModule(name string, runtimePath string) error {
	gatway, err := external.NewManagementService(runtimePath)
	if err != nil {
		return err
	}

	err, port := gatway.GetPort()
	if err != nil {
		return err
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	client, err := NewClient(ModManagementClientOpts{
		Port: &portInt,
	})
	if err != nil {
		return err
	}

	modules, err := client.InstalledModules()
	if err != nil {
		return err
	}

	// 判断是否已经安装
	for _, module := range modules {
		if module.Name == nil {
			continue
		}
		if *module.Name == name {
			return nil
		}
	}

	// 安装
	err = client.InstallModule(name)
	if err != nil {
		return err
	}

	return nil
}
