package lnd

import (
	"context"
	"fmt"

	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/sulusolutions/gol402/wallet"
)

// LndWallet implements the Wallet interface using an LND node.
type LndWallet struct {
	client lndclient.RouterClient
}

// NewLndWallet creates a new instance of LndWallet.
func NewLndWallet(client lndclient.RouterClient) *LndWallet {
	return &LndWallet{
		client: client,
	}
}

// LndWalletConfig holds configuration parameters for LndWallet.
type LndWalletConfig struct {
	MacaroonPath string
	TLSPath      string
	Network      string
	GrpcAddress  string
}

// NewLndWalletFromConfig creates a new LndWallet instance using the provided configuration.
func NewLndWalletFromConfig(cfg *LndWalletConfig) (*LndWallet, error) {
	lndCfg := &lndclient.LndServicesConfig{
		LndAddress:  cfg.GrpcAddress,
		Network:     lndclient.Network(cfg.Network),
		MacaroonDir: cfg.MacaroonPath,
		TLSPath:     cfg.TLSPath,
	}

	client, err := lndclient.NewLndServices(lndCfg)
	if err != nil {
		return nil, err
	}

	return &LndWallet{
		client: client.Router,
	}, nil
}

// PayInvoice attempts to pay the given invoice using the LND node.
func (lw *LndWallet) PayInvoice(ctx context.Context, invoice wallet.Invoice) (*wallet.PaymentResult, error) {
	// Construct the SendPaymentRequest
	payReq := lndclient.SendPaymentRequest{
		Invoice: string(invoice),
	}

	// Send the payment request to LND
	statusChan, errChan, err := lw.client.SendPayment(ctx, payReq)
	if err != nil {
		return nil, err
	}

	// Wait for payment status
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err() // Return the context's error
		case paymentStatus, ok := <-statusChan:
			if !ok {
				return nil, fmt.Errorf("payment status channel closed")
			}

			if paymentStatus.State == lnrpc.Payment_SUCCEEDED {
				preimage := paymentStatus.Preimage.String()
				return &wallet.PaymentResult{
					Preimage: preimage,
					Success:  true,
				}, nil
			} else if paymentStatus.State == lnrpc.Payment_FAILED {
				return nil, fmt.Errorf("payment failed: %v", paymentStatus.State)
			}

		case err := <-errChan:
			return nil, err
		}
	}
}
