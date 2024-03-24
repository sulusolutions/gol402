package lnd

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"io/ioutil"
	"github.com/sulusolutions/gol402/wallet"
)

type lndPaymentResponse struct {
	Amount          int    `json:"amount"`
	Description     string `json:"description"`
	Destination     string `json:"destination"`
	Fee             int    `json:"fee"`
	PaymentHash     string `json:"payment_hash"`
	PaymentPreimage string `json:"payment_preimage"`
	PaymentRequest  string `json:"payment_request"`
}

type lndWalletResponse struct {
	PaymentError    string `json:"payment_error"`
	PaymentPreimage string `json:"payment_preimage"`
	PaymentRoute    struct{} `json:"payment_route"`
	PaymentHash     string `json:"payment_hash"`
}


// LNDWallet implements the Wallet interface using the LND WALLET REST API.
type LNDWallet struct {
	// BaseURL is the base URL for the  LND wallet API.
	BaseURL string

	macaroonString []byte 


}

// NewLNDWallet creates a new instance of LNDWallet.
func NewLNDWallet(macaroonPath string,  address string) *LNDWallet {
			// Read macaroon file
	macaroon, err := ioutil.ReadFile(macaroonPath)
	if err != nil {
		fmt.Println("Error reading macaroon file:", err)
		
	}

	// Decode macaroon from hex
	macaroonBytes, err := hex.DecodeString(string(macaroon))
	if err != nil {
		fmt.Println("Error decoding macaroon:", err)
		
	}

	return &LNDWallet{
		BaseURL: address,
		macaroonString: macaroonBytes,
	}
}

// PayInvoice attempts to pay the given invoice and returns the result.
func (lndw *NewLNDWallet) PayInvoice(ctx context.Context, invoice wallet.Invoice) (*wallet.PaymentLndResult, error) {
	path := "/v2/invoices/settle"
	body := map[string]interface{}{
		"preimage": invoice,
	}

	responseBody, err := lndw.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, err
	}

	var lndWalletResponse lndWalletResponse
	if err := json.Unmarshal(responseBody, &lndWalletResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling lnd wallet response: %w", err)
	}

	var result wallet.PaymentLndResult
	result.PaymentHash = result.PaymentHash
	result.Success = true

	return &result, nil
}

func (lndw *LNDWallet) makeRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", lndw.BaseURL, path)

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
		req.Header.Set("Grpc-Metadata-macaroon", hex.EncodeToString(lndw.macaroonString))
		req.Header.Set("Content-Type", "application/json")

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
