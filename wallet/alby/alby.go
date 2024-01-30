package alby

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sulusolutions/l402/wallet"
)

// AlbyWallet implements the Wallet interface using the Alby REST API.
type AlbyWallet struct {
	// BaseURL is the base URL for the Alby API.
	BaseURL string
	// credentials is the Bearer token for authorization.
	credentials string
}

// NewAlbyWallet creates a new instance of AlbyWallet.
func NewAlbyWallet(token string) *AlbyWallet {
	return &AlbyWallet{
		BaseURL:     "https://api.getalby.com",
		credentials: token,
	}
}

// PayInvoice attempts to pay the given invoice and returns the result.
func (aw *AlbyWallet) PayInvoice(ctx context.Context, invoice wallet.Invoice) (*wallet.PaymentResult, error) {
	path := "/payments/bolt11"
	body := map[string]interface{}{
		"invoice": invoice,
		// Include "amount" if necessary
	}

	responseBody, err := aw.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, err
	}

	var result wallet.PaymentResult
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	result.Success = true
	return &result, nil
}

func (aw *AlbyWallet) makeRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", aw.BaseURL, path)

	var requestBody []byte
	var err error
	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", aw.credentials))
	req.Header.Set("User-Agent", "alby-go")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		// Here, you might want to unmarshal the response body to a structured error type, similar to the PHP example
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, responseBody)
	}

	return responseBody, nil
}
