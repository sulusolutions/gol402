package l402

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sulusolutions/l402/wallet"
)

// TestNewClient verifies that the NewClient function returns a Client instance with the expected wallet.
func TestNewClient(t *testing.T) {
	m := wallet.NewMockWallet(nil)
	c := NewClient(m)

	if c == nil {
		t.Errorf("NewClient returned nil")
	}
}

func TestMakeRequest(t *testing.T) {
	tests := []struct {
		name           string
		serverHandler  func(w http.ResponseWriter, r *http.Request)
		mockWalletErr  error
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Successful request without 402 challenge",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
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
			expectedStatus: http.StatusOK,
		},
		{
			name: "Handle 402 Payment Required with payment failure",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("WWW-Authenticate", `L402 macaroon="testMacaroon", invoice="testInvoice"`)
				w.WriteHeader(http.StatusPaymentRequired)
			},
			mockWalletErr: fmt.Errorf("payment error"),
			expectError:   true,
		},
		{
			name: "Server returns an error status code other than 402",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Server returns 402 without WWW-Authenticate header",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusPaymentRequired)
			},
			expectError: true,
		},
		{
			name: "Malformed WWW-Authenticate header",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("WWW-Authenticate", `L402 malformed="header"`)
				w.WriteHeader(http.StatusPaymentRequired)
			},
			expectError: true,
		},
		// Additional test cases can be added here as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := NewMockServer(tt.serverHandler)
			server := mockServer.Start()
			defer server.Close()

			mockWallet := wallet.NewMockWallet(tt.mockWalletErr)
			client := NewClient(mockWallet)

			resp, err := client.MakeRequest(context.Background(), server.URL, "GET")

			if (err != nil) != tt.expectError {
				t.Errorf("MakeRequest() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if resp != nil && resp.StatusCode != tt.expectedStatus {
				t.Errorf("MakeRequest() expected status = %v, got %v", tt.expectedStatus, resp.StatusCode)
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
