package v1

import "testing"

func TestDeleteShare(t *testing.T) {
	share, err := NewShareService("/var/run/casaos")
	if err != nil {
		t.Fatal(err)
	}
	err = share.DeleteShare("1")
	if err != nil {
		t.Fatal(err)
	}
}
