package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/SternKater/carsharing/pkg/mymoney"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MyMoneyHandler struct {
	mymoney.UnimplementedMyMoneyServiceServer
}

func NewMyMoneyHandler() *MyMoneyHandler {
	return &MyMoneyHandler{}
}

func (h *MyMoneyHandler) GetPaymentLink(ctx context.Context, req *mymoney.PaymentLinkRequest) (*mymoney.PaymentLinkResponse, error) {
	log.Printf("[PAYMENT-PROVIDER]: Creating invoice %d for amount %d", req.InvoiceId, req.AmountInCents)

	// some fake uuid
	paymentID := fmt.Sprintf("pay_uuid_%d", time.Now().UnixNano())
	// Фейковая ссылка, по которой якобы перейдет пользователь
	fakeURL := fmt.Sprintf("https://mymoney.ru/%v", paymentID)

	// goroutine for simulate payment
	go func(invoiceID int64, amount int64, payID string, callbackURL string) {
		// Zzzz....
		time.Sleep(5 * time.Second)

		log.Printf("[PAYMENT-PROVIDER]: User paid invoice %d! Sending Webhook to %s", invoiceID, callbackURL)

		// connect to our carsharing callback
		conn, err := grpc.NewClient(callbackURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("[PAYMENT-PROVIDER-ERROR]: Failed to connect to gateway callback: %v", err)
			return
		}
		defer conn.Close()

		client := mymoney.NewMyMoneyCallbackServiceClient(conn)
		
		// Alles ist gut!
		_, err = client.OnPaymentComplete(context.Background(), &mymoney.WebhookNotify{
			InvoiceId:     invoiceID,
			AmountInCents: amount,
			PaymentId:     payID,
			Status:        "success",
		})
		if err != nil {
			log.Printf("[PAYMENT-PROVIDER-ERROR]: Failed to deliver webhook: %v", err)
		} else {
			log.Printf("[PAYMENT-PROVIDER]: Webhook for invoice %d successfully delivered!", invoiceID)
		}
	}(req.InvoiceId, req.AmountInCents, paymentID, req.CallbackUrl)

	return &mymoney.PaymentLinkResponse{
		PaymentUrl: fakeURL,
		PaymentId:  paymentID,
	}, nil
}
