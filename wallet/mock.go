package wallet

import (
	"context"
)

const (
	mockPreimage = "12345abcd"
)

// MockWallet is a mock implementation of the Wallet interface for testing purposes.
type MockWallet struct {
	// Error to be returned by PayInvoice. If nil, PayInvoice will simulate a successful payment.
	PaymentError error
}

// NewMockWallet creates a new instance of MockWallet with customizable behavior.
func NewMockWallet(err error) *MockWallet {
	return &MockWallet{
		PaymentError: err,
	}
}

// PayInvoice simulates the payment process of an invoice.
// It returns a specified error or simulates a successful payment if the error is nil.
func (mw *MockWallet) PayInvoice(ctx context.Context, invoice Invoice) (*PaymentResult, error) {
	// Check if the context is done before proceeding, simulating a cancellation or timeout.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Continue if the context is not done.
	}

	if mw.PaymentError != nil {
		return &PaymentResult{
			Preimage: "",
			Success:  false,
			Error:    mw.PaymentError,
		}, nil
	}

	return &PaymentResult{
		Preimage: mockPreimage,
		Success:  true,
		Error:    nil,
	}, nil
}
