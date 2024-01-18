package external_test

import (
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/external"
	"gotest.tools/v3/assert"
)

func TestGPUInfos(t *testing.T) {
	t.Skip()
	GPUsInfo, err := external.GPUsInfo()
	assert.NilError(t, err)
	assert.Equal(t, len(GPUsInfo), 1)
}
