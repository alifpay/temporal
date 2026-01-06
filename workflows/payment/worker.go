package payment

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func RunCompleteSessionWorkflow(ctx context.Context, c client.Client) error {

	w := worker.New(c, "payment", worker.Options{})

	w.RegisterWorkflow(PaymentWorkflow)
	w.RegisterActivity(GetUSDRate)
	w.RegisterActivity(SendPayment)
	w.RegisterActivity(UpdatePaymentStatus)
	
	err := w.Run(interrupt(ctx))
	if err != nil {
		return err
	}
	return nil
}

func interrupt(ctx context.Context) <-chan any {
	intCh := make(chan any)

	go func() {
		defer close(intCh)
		<-ctx.Done()
	}()

	return intCh
}
