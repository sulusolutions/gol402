package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sulusolutions/gol402/tokenstore"
	"github.com/sulusolutions/gol402/wallet"
)

// TestNewClient verifies that the NewClient function returns a Client instance with the expected wallet.
func TestNewClient(t *testing.T) {
	m := wallet.NewMockWallet(nil)
	c := New(m, tokenstore.NewNoopStore())

	if c == nil {
		t.Errorf("NewClient returned nil")
	}
}

func TestMakeRequest(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler func(w http.ResponseWriter, r *http.Request)
		walletErr     error
		wantStatus    int
		wantError     bool
	}{
		{
			name: "Successful request without 402 challenge",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Handle 402 Payment Required with successful payment",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") == "" {
					w.Header().Set("WWW-Authenticate", `L402 macaroon="testMacaroon", invoice="testInvoice"`)
					w.WriteHeader(http.StatusPaymentRequired)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Handle 402 Payment Required with payment failure",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("WWW-Authenticate", `L402 macaroon="testMacaroon", invoice="testInvoice"`)
				w.WriteHeader(http.StatusPaymentRequired)
			},
			walletErr: fmt.Errorf("payment error"),
			wantError: true,
		},
		{
			name: "Server returns an error status code other than 402",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "Server returns 402 without WWW-Authenticate header",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusPaymentRequired)
			},
			wantError: true,
		},
		{
			name: "Malformed WWW-Authenticate header",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("WWW-Authenticate", `L402 malformed="header"`)
				w.WriteHeader(http.StatusPaymentRequired)
			},
			wantError: true,
		},
		// Additional test cases can be added here as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := NewMockServer(tt.serverHandler)
			server := mockServer.Start()
			defer server.Close()

			mockWallet := wallet.NewMockWallet(tt.walletErr)
			client := New(mockWallet, tokenstore.NewNoopStore())

			// Create the *http.Request object with the test server's URL
			req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := client.Do(req)

			if (err != nil) != tt.wantError {
				t.Errorf("MakeRequest() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if resp != nil && resp.StatusCode != tt.wantStatus {
				t.Errorf("MakeRequest() got status = %v, want %v", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

type mockServer struct {
	// HandlerFunc allows test cases to define custom behavior for the HTTP handler.
	HandlerFunc func(w http.ResponseWriter, r *http.Request)
}

// NewMockServer initializes a new instance of mockServer with the specified handler function.
func NewMockServer(handlerFunc func(w http.ResponseWriter, r *http.Request)) *mockServer {
	return &mockServer{
		HandlerFunc: handlerFunc,
	}
}

// Start runs the mock HTTP server and returns its URL.
func (ms *mockServer) Start() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(ms.HandlerFunc))
	return server
}
