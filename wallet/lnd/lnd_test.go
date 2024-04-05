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

	// Add fields to control the mock's behavior, e.g.,
	paymentSuccess bool
}

func (m *mockRouterClient) SendPayment(ctx context.Context, req lndclient.SendPaymentRequest) (chan lndclient.PaymentStatus, chan error, error) {
	statusChan := make(chan lndclient.PaymentStatus)
	errChan := make(chan error)

	go func() {
		if m.paymentSuccess {
			var preimage lntypes.Preimage
			if _, err := rand.Read(preimage[:]); err != nil {
				errChan <- err // Handle error appropriately
			}
			statusChan <- lndclient.PaymentStatus{
				State:    lnrpc.Payment_SUCCEEDED,
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
	t.Run("Successful payment", func(t *testing.T) {
		mockClient := &mockRouterClient{paymentSuccess: true}
		lndWallet := NewLndWallet(mockClient)

		invoice := wallet.Invoice("mock_invoice")
		result, err := lndWallet.PayInvoice(context.Background(), invoice)

		require.NoError(t, err)
		require.True(t, result.Success)
		// Add more assertions as needed
	})

	t.Run("Failed payment", func(t *testing.T) {
		mockClient := &mockRouterClient{paymentSuccess: false}
		lndWallet := NewLndWallet(mockClient)

		invoice := wallet.Invoice("mock_invoice")
		_, err := lndWallet.PayInvoice(context.Background(), invoice)

		require.Error(t, err)
		// Add more assertions as needed
	})

	// Add more test cases for different scenarios, e.g., context cancellation, invalid invoices, etc.
}
