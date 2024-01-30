package l402 // import "github.com/sulusolutions/l402"

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/sulusolutions/l402/tokenstore"
	"github.com/sulusolutions/l402/wallet"
)

// Challenge holds the parsed invoice and macaroon from the WWW-Authenticate header.
type Challenge struct {
	Invoice  string
	Macaroon string
}

// Client represents a client capable of handling L402 payments and making authenticated requests.
type Client struct {
	wallet wallet.Wallet
	store  tokenstore.Store
}

// NewClient creates a new L402 client with the provided wallet for handling payments
// and token store for storing L402 tokens.
func NewClient(w wallet.Wallet, s tokenstore.Store) *Client {
	return &Client{
		wallet: w,
		store:  s,
	}
}

// MakeRequest makes an HTTP request to the specified URL and handles L402 payment challenges.
// It automatically pays the invoice and retries the request with the L402 token if a 402 Payment Required response is received.
func (c *Client) MakeRequest(ctx context.Context, rawUrl string, method string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawUrl, nil)
	if err != nil {
		return nil, err
	}

	// Retrieve L402 token from store if available and try request.
	u, err := url.Parse(rawUrl)
	if err != nil {
		// For now we will just warn and continue, but this should be handled more gracefully.
		fmt.Printf("error parsing url: %v\n", err)
	}
	l402Token, ok := c.store.Get(u)
	if ok {
		req.Header.Set("Authorization", "L402 "+string(l402Token))
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Handle 402 Payment Required by delegating to the handlePaymentChallenge function
	if response.StatusCode == http.StatusPaymentRequired {
		authHeader := response.Header.Get("WWW-Authenticate")
		return c.handlePaymentChallenge(ctx, authHeader, rawUrl, method)
	}

	return response, nil
}

// handlePaymentChallenge handles the 402 Payment Required response by extracting the invoice and macaroon,
// paying the invoice, and constructing the L402 token for retrying the request.
func (c *Client) handlePaymentChallenge(ctx context.Context, authHeader, rawUrl, method string) (*http.Response, error) {
	challenge, err := parseHeader(authHeader)
	if err != nil {
		return nil, err
	}

	// Pay the invoice using the wallet
	paymentResult, err := c.wallet.PayInvoice(ctx, wallet.Invoice(challenge.Invoice))
	if err != nil {
		return nil, err
	}

	// Construct L402 token using the challenge details and the preimage from the payment result
	l402Token := constructL402Token(*challenge, paymentResult.Preimage)

	// Prepare a new request for retrying with the L402 token
	retryReq, err := http.NewRequestWithContext(ctx, method, rawUrl, nil)
	if err != nil {
		return nil, err
	}
	retryReq.Header.Set("Authorization", "L402 "+l402Token)
	u, err := url.Parse(rawUrl)
	if err != nil {
		// For now we will just warn and continue, but this should be handled more gracefully.
		fmt.Printf("error parsing url: %v\n", err)
	}

	c.store.Put(u, tokenstore.Token(l402Token))

	// Retry the request with Authorization header
	return http.DefaultClient.Do(retryReq)
}

// parseHeader uses regular expressions to extract invoice and macaroon from the WWW-Authenticate header.
func parseHeader(header string) (*Challenge, error) {
	// Define regular expressions for matching invoice and macaroon
	invoiceRegex := regexp.MustCompile(`invoice="([^"]+)"`)
	macaroonRegex := regexp.MustCompile(`macaroon="([^"]+)"`)

	// Find matches
	invoiceMatch := invoiceRegex.FindStringSubmatch(header)
	macaroonMatch := macaroonRegex.FindStringSubmatch(header)

	if invoiceMatch == nil || macaroonMatch == nil {
		return nil, fmt.Errorf("failed to parse WWW-Authenticate header")
	}

	// Extract invoice and macaroon from matches
	return &Challenge{
		Invoice:  invoiceMatch[1],
		Macaroon: macaroonMatch[1],
	}, nil
}

// constructL402Token constructs the L402 token from the given Challenge and preimage.
func constructL402Token(challenge Challenge, preimage string) string {
	// Encode the macaroon in base64
	macaroonBase64 := challenge.Macaroon
	// Encode the preimage in hex
	preimageHex := preimage
	// Construct the token in the format "macaroons:preimage"
	return macaroonBase64 + ":" + preimageHex
}
