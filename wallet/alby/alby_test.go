package alby

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockAlbyServer struct {
	server        *httptest.Server
	ExpectedToken string
	Error         error
}

func NewMockAlbyServer(creds string, err error) *MockAlbyServer {
	mock := &MockAlbyServer{
		ExpectedToken: creds,
		Error:         err,
	}

	mock.server = httptest.NewServer(http.HandlerFunc(mock.handler))
	return mock
}

func (m *MockAlbyServer) handler(w http.ResponseWriter, r *http.Request) {
	// Fail if error set.
	if m.Error != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}

	// Check for the correct bearer token.
	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+m.ExpectedToken {
		http.Error(w, `{"error": "Invalid or missing bearer token"}`, http.StatusUnauthorized)
		return
	}

	// Validate the request URL and body format.
	if r.URL.Path != "/payments/bolt11" || r.Method != "POST" {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	// If everything is good, return a fake successful response.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"payment_preimage": "preimage123", "success": true}`)) //nolint:errcheck
}

func (m *MockAlbyServer) Close() {
	m.server.Close()
}

func (m *MockAlbyServer) URL() string {
	return m.server.URL
}

func TestPayInvoice(t *testing.T) {
	creds := "correctToken"
	s := NewMockAlbyServer(creds, nil)
	defer s.Close()

	w := NewAlbyWallet(creds)
	w.BaseURL = s.URL() // Point the wallet to the mock server

	_, err := w.PayInvoice(context.Background(), "validInvoice")
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Test with incorrect token
	w.credentials = "wrongToken"
	_, err = w.PayInvoice(context.Background(), "validInvoice")
	if err == nil {
		t.Fatal("Expected an error due to wrong bearer token, but got none")
	}

	// Additional tests can be added here to cover more scenarios
}

func TestPayInvoiceErrors(t *testing.T) {
	creds := "correctToken"
	s := NewMockAlbyServer(creds, fmt.Errorf("Internal server error"))
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
			w := NewAlbyWallet(tc.creds)
			w.BaseURL = s.URL() // Point the wallet to the mock server

			s := NewMockAlbyServer(creds, tc.serverError)
			defer s.Close()

			_, err := w.PayInvoice(context.Background(), "validInvoice")

			// Check for the expected error
			if err == nil {
				t.Errorf("Expected error but got nil.")
			}
		})
	}
}
