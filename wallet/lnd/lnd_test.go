

package lnd_test

import (
	"context"
	"errors"
	"testing"
	"github.com/sulusolutions/gol402/lnd"
)

func TestPayLndInvoice(t *testing.T) {
	// Mock LNDWallet instance
	mockLNDWallet := &lnd.LNDWallet{
		BaseURL:        "http://127.0.0.1:28332",
		MacaroonString: []byte("mock-macaroon"),
	}

	// Mock context
	ctx := context.Background()

	// Mock invoice
	mockInvoice := lnd.Invoice("mock-invoice")

	// Mock successful response
	mockResponseBody := []byte(`{"payment_hash":"mock-payment-hash"}`)
	mockMakeRequest := func(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
		return mockResponseBody, nil
	}
	mockLNDWallet.MakeRequest = mockMakeRequest // Injecting mock makeRequest function

	// Call PayLndInvoice
	result, err := mockLNDWallet.PayLndInvoice(ctx, mockInvoice)

	// Check for errors
	if err != nil {
		t.Errorf("PayLndInvoice returned unexpected error: %v", err)
	}

	// Check result
	expectedPaymentHash := "mock-payment-hash"
	if result.PaymentHash != expectedPaymentHash {
		t.Errorf("PayLndInvoice returned unexpected payment hash, got: %s, want: %s", result.PaymentHash, expectedPaymentHash)
	}
	if !result.Success {
		t.Error("PayLndInvoice returned unexpected success value, got: false, want: true")
	}

	// Mock failure response
	mockMakeRequestError := errors.New("mock makeRequest error")
	mockLNDWallet.MakeRequest = func(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
		return nil, mockMakeRequestError
	}

	// Call PayLndInvoice
	_, err = mockLNDWallet.PayLndInvoice(ctx, mockInvoice)

	// Check for errors
	if err == nil {
		t.Error("PayLndInvoice expected error but got nil")
	}
	if err != mockMakeRequestError {
		t.Errorf("PayLndInvoice returned unexpected error, got: %v, want: %v", err, mockMakeRequestError)
	}
}
