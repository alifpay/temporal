package payment

import (
	"time"

	"github.com/alifpay/temporal/models"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func PaymentWorkflow(ctx workflow.Context, payment models.Payment) error {

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 3.0,
			MaximumInterval:    20 * time.Second,
			MaximumAttempts:    5,
		},
	})

	var usdRate float64
	err := workflow.ExecuteActivity(ctx, GetUSDRate, payment.PayDate).Get(ctx, &usdRate)
	if err != nil {
		return err
	}
	payment.Amount = payment.Amount * usdRate

	err = workflow.ExecuteActivity(ctx, SendPayment, payment).Get(ctx, nil)
	if err != nil {
		return err
	}
	//get status of payment
	var sts models.PaymentStatus
	workflow.GetSignalChannel(ctx, models.PaymentStatusSignal).Receive(ctx, &sts)
	if sts.Status == "success" {
		err = workflow.ExecuteActivity(ctx, UpdatePaymentStatus, sts).Get(ctx, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
