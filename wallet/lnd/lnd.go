package lnd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"github.com/sulusolutions/gol402/wallet"
)


type DecodedLndInvoiceResponse struct {
	Destination     string `json:"destination"`
	PaymentHash     string `json:"payment_hash"`
	NumSatoshis     int    `json:"num_satoshis"`
	Timestamp       int    `json:"timestamp"`
	Expiry          int    `json:"expiry"`
	Description     string `json:"description"`
	DescriptionHash string `json:"description_hash"`
	FallbackAddr    string `json:"fallback_addr"`
	CltvExpiry      int    `json:"cltv_expiry"`
	RouteHints      []struct {
		HopHints []struct {
			NodeID                     string `json:"node_id"`
			ChanID                     string `json:"chan_id"`
			FeeBaseMSat                int    `json:"fee_base_msat"`
			FeeProportionalMillionths  int    `json:"fee_proportional_millionths"`
			CltvExpiryDelta            int    `json:"cltv_expiry_delta"`
		} `json:"hop_hints"`
	} `json:"route_hints"`
}


type LndWallet struct {
	BaseURL string
	macaroonBytes []byte
}

func readMacaroonFromFile() ([]byte, error) {
	// Get macaroon file path from environment variable
	macaroonPath := os.Getenv("MACAROONPATH")

	// Read macaroon from file
	macaroonBytes, err := ioutil.ReadFile(macaroonPath)
	if err != nil {
		fmt.Println("Error reading macaroon file:", err)
		return nil, fmt.Errorf("Error reading macaroon file: %w", err)
	}

	return macaroonBytes, nil
}

// NewLndWallet creates a new instance of LndWallet.
func NewAlbyWallet() *LndWallet {

	// macaroonBytesStr, err := readMacaroonFromFile()
	
	// if err != nil {
		// Handle error accordingly, e.g., log it or return an error
  // }
	macaroonBytesStr, _ := readMacaroonFromFile()	
	return &LndWallet{
		BaseURL:     os.Getenv("BASEURL"),
		macaroonBytes: macaroonBytesStr,
	}

}


// Attempt to decode an invoice
func (lnd * LndWallet) DecodeLndInvoice(ctx context.Context, invoice wallet.Invoice) (*wallet.DecodeLndInvoice, error) {
		path := fmt.Sprintf("/v1/payreq/%s", invoice)

		responseBody, err := lnd.makeGetRequest(ctx, path, )
		if err != nil {
			return nil, err
		}

		var decodeLndInvoiceResponse DecodedLndInvoiceResponse
		if err := json.Unmarshal(responseBody, &decodeLndInvoiceResponse); err != nil {
			return nil, fmt.Errorf("error unmarshaling LND response: %w", err)
		}
	
		var result wallet.DecodeLndInvoice
		result.Amount = decodeLndInvoiceResponse.NumSatoshis
	
		return &result, nil

}


func (lnd *LndWallet) makeGetRequest(ctx context.Context, path string ) ([]byte, error) {
	url := fmt.Sprintf("%s%s", lnd.BaseURL, path)

	var requestBody []byte
	var err error

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Grpc-Metadata-macaroon", string(lnd.macaroonBytes))

	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Pull the respoonse from the body 
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// check if status is noot 200 
	if resp.StatusCode != http.StatusOK {
		
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, responseBody)
	}

	return responseBody, nil
}
