package version

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseVersion1(t *testing.T) {
	v1, v2, v3, v4, a, err := ParseVersion("v1.2.3.4-alpha1")

	assert.NilError(t, err)
	assert.Equal(t, v1, 1)
	assert.Equal(t, v2, 2)
	assert.Equal(t, v3, 3)
	assert.Equal(t, v4, 4)
	assert.Equal(t, a, "alpha1")
}

func TestParseVersion2(t *testing.T) {
	v1, v2, v3, v4, a, err := ParseVersion("1.2.3.4-alpha1")

	assert.NilError(t, err)
	assert.Equal(t, v1, 1)
	assert.Equal(t, v2, 2)
	assert.Equal(t, v3, 3)
	assert.Equal(t, v4, 4)
	assert.Equal(t, a, "alpha1")
}

func TestParseVersion3(t *testing.T) {
	v1, v2, v3, v4, a, err := ParseVersion("1.2")

	assert.NilError(t, err)
	assert.Equal(t, v1, 1)
	assert.Equal(t, v2, 2)
	assert.Equal(t, v3, 0)
	assert.Equal(t, v4, 0)
	assert.Equal(t, a, "")
}

func TestParseVersion4(t *testing.T) {
	v1, v2, v3, v4, a, err := ParseVersion("1.2.3.4.5-alpha1-whatever")

	assert.NilError(t, err)
	assert.Equal(t, v1, 1)
	assert.Equal(t, v2, 2)
	assert.Equal(t, v3, 3)
	assert.Equal(t, v4, 4)
	assert.Equal(t, a, "alpha1-whatever")
}

func TestParseVersion5(t *testing.T) {
	v1, v2, v3, v4, a, err := ParseVersion("a.b")

	if err == nil {
		t.Error("expected error")
	}

	assert.Equal(t, v1, -1)
	assert.Equal(t, v2, -1)
	assert.Equal(t, v3, -1)
	assert.Equal(t, v4, -1)
	assert.Equal(t, a, "")
}
