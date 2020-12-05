package ssh

import (
	"encoding/pem"
	"testing"
)

func TestNewKeyPair(t *testing.T) {
	pair, err := NewKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	if privPem := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Headers: nil, Bytes: pair.PrivateKey}); len(privPem) == 0 {
		t.Fatal("No PEM returned")
	}
}
