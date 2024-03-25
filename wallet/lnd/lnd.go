
package lnd

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"
	"io/ioutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
)

	// Define payment request
	type CustomSendPaymentRequest struct {
		Invoice          string
		MaxFee           int64
		Timeout          time.Duration
		AllowSelfPayment bool
	}

type LNDWallet struct {
	BaseURL string

	macaroonString []byte 


}

func NewLNDWallet(macaroonPath string,  address string) *LNDWallet {
			// Read macaroon file
	macaroon, err := ioutil.ReadFile(macaroonPath)
	if err != nil {
		fmt.Println("Error reading macaroon file:", err)
		
	}

	// Decode macaroon from hex
	macaroonBytes, err := hex.DecodeString(string(macaroon))
	if err != nil {
		fmt.Println("Error decoding macaroon:", err)
		
	}

	return &LNDWallet{
		BaseURL: address,
		macaroonString: macaroonBytes,
	}
}


func (lndw *NewLNDWallet) HandleSettleInvoice(ctx context.Context, invoice wallet.Invoice) (*wallet.PaymentLndResult, error) {
	client, err := lndclient.NewLndServices(
		lndclient.LndServicesConfig{
			LndAddress: lndw.BaseURL,
		},
	)

	if err != nil {
		fmt.Println("Error connecting to LND:", err)
		return
	}

	payReq := CustomSendPaymentRequest{
		Invoice:          invoice,
		MaxFee:           100,
		Timeout:          time.Duration(2e16),
		AllowSelfPayment: true,
	}

	
	md := metadata.New(map[string]string{
		"macaroon": hex.EncodeToString(lndw.macaroonString),
	})

	
	ctx := metadata.NewOutgoingContext(context.Background(), md)

// Send payment with metadata
statusChan, errChan, err := sendPaymentWithCustomRequest(ctx, client.Router, payReq)
if err != nil {
	fmt.Println("Error sending payment:", err)
	return
}

// Wait for payment status
for {
	select {
	case paymentStatus, ok := <-statusChan:
		// If the channel is closed, exit the loop.
		if !ok {
			fmt.Println("Payment status channel closed")
			return nil, "Payment status channel closed"
		}

		if paymentStatus.State == lnrpc.Payment_SUCCEEDED {
			preimage := fmt.Sprintf("%s", paymentStatus.Preimage.String())
			fmt.Println("Payment successful. Preimage:", preimage)
			var result wallet.PaymentLndResult
			result.PaymentHash = result.PaymentHash
			result.Success = true
			return &result, nil

		} else if paymentStatus.State == lnrpc.Payment_FAILED {
			fmt.Println("Payment failed:", paymentStatus.FailureReason)
			return nil, paymentStatus.FailureReason
		}
	

	case err := <-errChan:
		fmt.Println("Error while waiting for payment status:", err)
		return nil, err
	}
}
	


}


func sendPaymentWithCustomRequest(ctx context.Context, router lnrpc.RouterClient, req CustomSendPaymentRequest) (<-chan *lnrpc.PaymentStatus, <-chan error, error) {
	payReq := lndclient.SendPaymentRequest{
		Invoice:          req.Invoice,
		MaxFee:           req.MaxFee,
		Timeout:          req.Timeout,
		AllowSelfPayment: req.AllowSelfPayment,
	};

	return router.SendPayment(ctx, payReq);
}
