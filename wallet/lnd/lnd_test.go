package lnd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockLndServer struct {
	server        *httptest.Server
	address string
	macaroonString string
	Error         error
}

func NewMockLndServer(macaroonString string address string err error) *MockLndServer {
	mock := &MockLndServer{
		address: address
		macaroonString: macaroonString,
		Error:         err,
	}

	mock.server = httptest.NewServer(http.HandlerFunc(mock.handler))
	return mock
}

func (m *MockLndServer) handler(w http.ResponseWriter, r *http.Request) {
	// Fail if error set.
	if m.Error != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}

	// Check for the correct macaroon auth.
	macaroon_auth := r.Header.Get("Grpc-Metadata-macaroon")
	if macaroon_auth != "Bearer "+m.ExpectedToken {
		http.Error(w, `{"error": "Invalid or missing bearer token"}`, http.StatusUnauthorized)
		return
	}

	// Validate the request URL and body format.
	if r.URL.Path != "/v2/invoices/settle" || r.Method != "POST" {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	// If everything is good, return a fake successful response.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"PaymentHash": "PaymentHash", "success": true}`)) //nolint:errcheck
}

func (m *MockLndServer) Close() {
	m.server.Close()
}

func (m *MockLndServer) URL() string {
	return m.server.URL
}

func TestPayInvoice(t *testing.T) {
	macaroonString := "/Users/macbookpro/Library/Application Support/Lnd/data/chain/bitcoin/regtest/admin.macaroon"
	address := "http://127.0.0.1:28332"
	s := NewMockAlbyServer(macaroonString, address nil)
	defer s.Close()

	w := NewMockLndServer(creds)
	w.BaseURL = s.address() // Point the wallet to the mock server

	_, err := w.PayLndInvoice(context.Background(), "validInvoice")
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Test with incorrect token
	w.credentials = "wrongToken"
	_, err = w.PayLndInvoice(context.Background(), "validInvoice")
	if err == nil {
		t.Fatal("Expected an error due to wrong bearer token, but got none")
	}

	// Additional tests can be added here to cover more scenarios
}

func TestPayInvoiceErrors(t *testing.T) {
	macaroonString := "/Users/macbookpro/Library/Application Support/Lnd/data/chain/bitcoin/regtest/admin.macaroon"
	address := "http://127.0.0.1:28332"
	s := NewMockAlbyServer(macaroonString, address nil)
	defer s.Close()


	tests := []struct {
		name        string
		creds       string
		serverError error
	}{
		{
			name:        "Server Error",
			creds:       creds,
			serverError: fmt.Errorf("Internal server error"),
		},
		{
			name:  "Invalid Token",
			creds: "wrongToken",
		},
		// Add more error scenarios as needed
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := NewMockAlbyServer(macaroonString, address nil)
			w.BaseURL = s.address() // Point the wallet to the mock server

			s := NewMockLndServer(creds, tc.serverError)
			defer s.Close()

			_, err := w.PayLndInvoice(context.Background(), "validInvoice")

			// Check for the expected error
			if err == nil {
				t.Errorf("Expected error but got nil.")
			}
		})
	}
}
