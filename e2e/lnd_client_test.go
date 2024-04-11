//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/sulusolutions/gol402/client"
	"github.com/sulusolutions/gol402/tokenstore"
	"github.com/sulusolutions/gol402/wallet/lnd"
)

func TestLndClientE2E(t *testing.T) {
	// Data comes from e2e testing environment defined with docker-compose.
	lndCfg := &lnd.LndWalletConfig{
		MacaroonPath: "/data/data/chain/bitcoin/regtest",
		TLSPath:      "/data/tls.cert",
		Network:      "regtest",
		GrpcAddress:  "lnd2:10010",
	}
	lndWallet, err := lnd.NewLndWalletFromConfig(lndCfg)
	require.NoError(t, err)

	// Initialize an in-memory token store
	tokenStore := tokenstore.NewInMemoryStore()

	// Create a new L402 client.
	client := client.New(lndWallet, tokenStore)

	// Create a new HTTP request to the L402 API
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://aperture:8700/randomnumber", nil)
	require.NoError(t, err)

	// Use the modified MakeRequest function which takes *http.Request
	response, err := client.Do(ctx, req)
	require.NoError(t, err)
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}
}
