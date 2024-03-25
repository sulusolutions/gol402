package lnd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// MockRouterClient is a mock implementation of lnrpc.RouterClient for testing.
type MockRouterClient struct{}

func (m MockRouterClient) SendPayment(ctx context.Context, req lndclient.SendPaymentRequest) (<-chan *lnrpc.PaymentStatus, <-chan error, error) {
	// Mock implementation
	return nil, nil, nil
}

func TestHandleSettleInvoice_Success(t *testing.T) {
	// Mock successful payment status
	statusChan := make(chan *lnrpc.PaymentStatus)
	close(statusChan)

	mockInvoice := "mock_invoice"

	lndw := NewLNDWallet("mock_macaroon_path", "mock_address")
	result, err := lndw.HandleSettleInvoice(context.Background(), mockInvoice)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestHandleSettleInvoice_Failure(t *testing.T) {
	// Mock failed payment status
	statusChan := make(chan *lnrpc.PaymentStatus)
	close(statusChan)

	mockInvoice := "mock_invoice"

	lndw := NewLNDWallet("mock_macaroon_path", "mock_address")
	result, err := lndw.HandleSettleInvoice(context.Background(), mockInvoice)

	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestSendPaymentWithCustomRequest(t *testing.T) {
	mockRouter := MockRouterClient{}

	ctx := context.Background()
	req := CustomSendPaymentRequest{
		Invoice:          "mock_invoice",
		MaxFee:           100,
		Timeout:          time.Duration(2e16),
		AllowSelfPayment: true,
	}

	statusChan, errChan, err := sendPaymentWithCustomRequest(ctx, mockRouter, req)

	assert.Nil(t, err)
	assert.NotNil(t, statusChan)
	assert.NotNil(t, errChan)
}

func TestNewLNDWallet(t *testing.T) {
	macaroonPath := "/Users/macbookpro/.polar/networks/2/volumes/c-lightning/bob/rest-api/access.macaroon"
	address := "127.0.0.1:11002"

	lndw := NewLNDWallet(macaroonPath, address)

	assert.NotNil(t, lndw)
	assert.Equal(t, macaroonPath, lndw.macaroonPath)
	assert.Equal(t, address, lndw.BaseURL)
}
