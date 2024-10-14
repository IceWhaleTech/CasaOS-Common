package modmanagement_test

import (
	"testing"

	modmanagement "github.com/IceWhaleTech/CasaOS-Common/pkg/mod_management"
	"github.com/stretchr/testify/assert"
)

func TestInstallableModules(t *testing.T) {
	client, err := modmanagement.NewClient(modmanagement.ModManagementClientOpts{})
	assert.NoError(t, err)
	modules, err := client.InstallableModules()
	assert.NoError(t, err)

	t.Log(modules)
}
