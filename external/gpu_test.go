package external_test

import (
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/external"
	"gotest.tools/v3/assert"
)

func TestGPUInfo(t *testing.T) {
	t.Skip()
	result, err := external.GPUInfoList()
	assert.NilError(t, err)
	assert.Equal(t, len(result), 1)
}
