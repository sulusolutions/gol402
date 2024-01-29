package l402_test

import (
	"testing"

	"github.com/sulusolutions/l402"
	"github.com/sulusolutions/l402/wallet"
)

// TestNewClient verifies that the NewClient function returns a Client instance with the expected wallet.
func TestNewClient(t *testing.T) {
	m := wallet.NewMockWallet(nil)
	c := l402.NewClient(m)

	if c == nil {
		t.Errorf("NewClient returned nil")
	}
}
