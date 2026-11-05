package secrets

import (
	"bytes"
	"testing"
)

func TestCipherRoundTripAndScopeBinding(t *testing.T) {
	cipher, err := New(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := cipher.Seal([]byte("provider-token"), []byte("connection-1"))
	if err != nil {
		t.Fatal(err)
	}
	opened, err := cipher.Open(sealed, []byte("connection-1"))
	if err != nil || string(opened) != "provider-token" {
		t.Fatalf("opened=%q err=%v", opened, err)
	}
	if _, err := cipher.Open(sealed, []byte("connection-2")); err == nil {
		t.Fatal("expected associated-data mismatch")
	}
}
