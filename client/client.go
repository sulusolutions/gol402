package client // import "github.com/sulusolutions/l402"

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/sulusolutions/gol402/tokenstore"
	"github.com/sulusolutions/gol402/wallet"
)

// Challenge holds the parsed invoice and macaroon from the WWW-Authenticate header.
type Challenge struct {
	HeaderKey string
	Invoice   string
	Macaroon  string
}

// Client represents a client capable of handling L402 payments and making authenticated requests.
type Client struct {
	wallet wallet.Wallet
	store  tokenstore.Store
}

// New creates a new L402 client with the provided wallet for handling payments
// and token store for storing L402 tokens.
func New(w wallet.Wallet, s tokenstore.Store) *Client {
	return &Client{
		wallet: w,
		store:  s,
	}
}

// Do makes an HTTP request and handles L402 payment challenges.
// It automatically pays the invoice and retries the request with the L402 token if a 402 Payment Required response is received.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Ensure the request context is set
	req = req.WithContext(ctx)

	// Try to retrieve and use L402 token if available
	l402Token, ok := c.store.Get(req.URL)
	if ok {
		req.Header.Set("Authorization", "L402 "+string(l402Token))
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusPaymentRequired {
		authHeader := response.Header.Get("WWW-Authenticate")
		return c.handlePaymentChallenge(ctx, authHeader, req.URL.String(), req.Method)
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
	retryReq.Header.Set("Authorization", l402Token)
	u, err := url.Parse(rawUrl)
	if err != nil {
		// For now we will just warn and continue, but this should be handled more gracefully.
		fmt.Printf("error parsing url: %v\n", err)
	}

	c.store.Put(u, tokenstore.Token(l402Token))

	// Retry the request with Authorization header
	return http.DefaultClient.Do(retryReq)
}

var (
	headerKeyRegex = regexp.MustCompile(`^(LSAT|L402)`)
	invoiceRegex   = regexp.MustCompile(`invoice="([^"]+)"`)
	macaroonRegex  = regexp.MustCompile(`macaroon="([^"]+)"`)
)

// parseHeader uses regular expressions to extract the header key, invoice, and macaroon from the WWW-Authenticate header.
func parseHeader(header string) (*Challenge, error) {
	// Find matches using the pre-compiled regex
	headerKeyMatch := headerKeyRegex.FindString(header)
	invoiceMatch := invoiceRegex.FindStringSubmatch(header)
	macaroonMatch := macaroonRegex.FindStringSubmatch(header)

	// Check for each match and return specific errors
	if headerKeyMatch == "" {
		return nil, fmt.Errorf("header key (LSAT or L402) not found in WWW-Authenticate header")
	}
	if invoiceMatch == nil {
		return nil, fmt.Errorf("invoice not found in WWW-Authenticate header")
	}
	if macaroonMatch == nil {
		return nil, fmt.Errorf("macaroon not found in WWW-Authenticate header")
	}

	// Extract header key, invoice, and macaroon from matches and return the Challenge struct
	return &Challenge{
		HeaderKey: headerKeyMatch,
		Invoice:   invoiceMatch[1],
		Macaroon:  macaroonMatch[1],
	}, nil
}

// constructL402Token constructs the L402 token from the given Challenge and preimage.
func constructL402Token(challenge Challenge, preimage string) string {
	// Construct and return the token using fmt.Sprintf for formatting
	return fmt.Sprintf("%s %s:%s", challenge.HeaderKey, challenge.Macaroon, preimage)
}
