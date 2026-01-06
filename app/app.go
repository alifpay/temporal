package app

import (
	"context"
	"log"

	"github.com/alifpay/temporal/models"
	"github.com/alifpay/temporal/workflows/payment"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

var TemporalClient client.Client

func startPaymentWorkflow(ctx context.Context, pay models.Payment) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                    pay.ID,
		TaskQueue:             "payment",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	_, err := TemporalClient.ExecuteWorkflow(ctx, workflowOptions, payment.PaymentWorkflow, pay)
	if err != nil {
		log.Println("Unable to start PaymentWorkflow", err, pay.ID)
	}
}
