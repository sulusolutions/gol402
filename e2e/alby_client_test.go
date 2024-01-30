//go:build e2e
// +build e2e

// Package e2e contains end-to-end tests for the L402 client.
package e2e

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/sulusolutions/gol402/client"
	"github.com/sulusolutions/gol402/tokenstore"
	"github.com/sulusolutions/gol402/wallet/alby"
)

func TestClientE2E(t *testing.T) {
	// Retrieve the bearer token from an environment variable
	bearerToken := os.Getenv("ALBY_BEARER_TOKEN")
	if bearerToken == "" {
		t.Fatalf("ALBY_BEARER_TOKEN is not set, skipping E2E test")
	}

	// Initialize the Alby wallet with the bearer token
	albyWallet := alby.NewAlbyWallet(bearerToken)

	// Initialize an in-memory token store
	tokenStore := tokenstore.NewInMemoryStore()

	// Create a new L402 client with the Alby wallet and in-memory token store
	client := client.NewClient(albyWallet, tokenStore)

	// Make a request to the L402 API
	ctx := context.Background()
	url := "https://rnd.ln.sulu.sh/randomnumber"
	response, err := client.MakeRequest(ctx, url, "GET")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}
}
