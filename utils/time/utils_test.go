package time

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestGetSystemTimeZoneName(t *testing.T) {
	timezone := GetSystemTimeZoneName()

	assert.Assert(t, timezone != "")

	location, err := time.LoadLocation(timezone)

	assert.NilError(t, err)
	assert.Assert(t, location != nil)
}
