// Package wallet includes the functionality for wallet implementations capable of handling L402 payments.
package wallet

import (
	"context"
)

// Invoice represents the structure of an invoice for payment.
type Invoice string

// PaymentResult represents the result of a payment attempt.
type PaymentResult struct {
	// Include fields like Preimage, Success, Error, etc.
	Preimage string
	Success  bool
}

type PaymentLndResult struct {
	// Include fields like Preimage, Success, Error, etc.
	PaymentHash string
	Success  bool
}



// Wallet defines the interface for wallet implementations capable of handling L402 payments.
type Wallet interface {
	// PayInvoice attempts to pay the given invoice and returns the result.
	// It should handle necessary logic like decoding the invoice, making the payment through the wallet's API, and returning the preimage if successful.
	PayInvoice(ctx context.Context, invoice Invoice) (*PaymentResult, error)
}


type LndWallet interface {
	// PayLndInvoice attempts to pay the given invoice via LND Wallet and returns the result.
	// Return Payment Hash and Success boolean
	PayLndInvoice(ctx context.Context, invoice Invoice) (*PaymentLndResult, error)
}
