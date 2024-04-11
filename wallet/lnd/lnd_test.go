package lnd

import (
	"context"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/stretchr/testify/require"
	"github.com/sulusolutions/gol402/wallet"
)

type mockRouterClient struct {
	lndclient.RouterClient

	completePaymentOp bool                        // Result is relayed through status channel only if this is true
	paymentStatus     lnrpc.Payment_PaymentStatus // Status to be returned by SendPayment through status channel.
	mockError         error                       // Add field to simulate an error from SendPayment
}

func (m *mockRouterClient) SendPayment(ctx context.Context, req lndclient.SendPaymentRequest) (chan lndclient.PaymentStatus, chan error, error) {
	if m.mockError != nil {
		return nil, nil, m.mockError // Return the mock error immediately
	}

	statusChan := make(chan lndclient.PaymentStatus)
	errChan := make(chan error)

	go func() {
		if m.completePaymentOp {
			var preimage lntypes.Preimage
			if _, err := rand.Read(preimage[:]); err != nil {
				errChan <- err // Handle error appropriately
			}
			statusChan <- lndclient.PaymentStatus{
				State:    m.paymentStatus,
				Preimage: preimage, // Replace with a valid preimage if needed
			}
		} else {
			errChan <- errors.New("mock payment failure")
		}

		close(statusChan)
		close(errChan)
	}()

	return statusChan, errChan, nil
}

func TestLndWallet_PayInvoice(t *testing.T) {
	// Define the test cases
	tests := []struct {
		name              string                      // Name of the test case
		completePaymentOp bool                        // Mock the payment result
		paymentStatus     lnrpc.Payment_PaymentStatus // Mock the payment status
		mockError         error                       // Mock error returned by SendPayment
		expectSuccess     bool                        // Expected result of PayInvoice
		expectError       bool                        // Whether an error is expected
		setupContext      func() context.Context      // Function to setup the context (e.g., for cancellation)
	}{
		{
			name:              "Successful payment",
			completePaymentOp: true,
			paymentStatus:     lnrpc.Payment_SUCCEEDED,
			expectSuccess:     true,
			expectError:       false,
			setupContext:      context.Background,
		},
		{
			name:              "Failed payment",
			completePaymentOp: true,
			paymentStatus:     lnrpc.Payment_FAILED,
			expectSuccess:     false,
			expectError:       true,
			setupContext:      context.Background,
		},
		{
			name:              "Failed payment",
			completePaymentOp: false,
			paymentStatus:     lnrpc.Payment_FAILED,
			expectSuccess:     false,
			expectError:       true,
			setupContext:      context.Background,
		},
		{
			name:              "Context canceled before payment",
			completePaymentOp: true,
			paymentStatus:     lnrpc.Payment_FAILED,
			expectSuccess:     false,
			expectError:       true,
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
		},
		{
			name:              "Invalid invoice format",
			completePaymentOp: false,
			paymentStatus:     lnrpc.Payment_FAILED,
			expectSuccess:     false,
			expectError:       true,
			setupContext:      context.Background,
		},
		{
			name:              "Unexpected error from SendPayment",
			completePaymentOp: false,
			paymentStatus:     lnrpc.Payment_FAILED,
			mockError:         errors.New("unexpected error"),
			expectSuccess:     false,
			expectError:       true,
			setupContext:      context.Background,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockRouterClient{
				completePaymentOp: tc.completePaymentOp,
				paymentStatus:     tc.paymentStatus,
				mockError:         tc.mockError,
			}
			lndWallet := NewLndWallet(mockClient)

			ctx := tc.setupContext()
			invoice := wallet.Invoice("mock_invoice")
			result, err := lndWallet.PayInvoice(ctx, invoice)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectSuccess, result.Success)
			}
		})
	}
}
