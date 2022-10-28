package external

import "testing"

func TestDeleteShare(t *testing.T) {
	share := NewShareService("/var/run/casaos")

	err := share.DeleteShare("1")
	if err != nil {
		t.Fatal(err)
	}
}
