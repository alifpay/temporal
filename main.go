package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alifpay/temporal/app"
	"github.com/alifpay/temporal/infra/rmq"
	"github.com/alifpay/temporal/infra/server"
	"github.com/alifpay/temporal/workflows/payment"
	"go.temporal.io/sdk/client"
	"golang.org/x/sync/errgroup"
)

func main() {
	g, ctx := errgroup.WithContext(context.Background())
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	var err error

	// temporal.io client
	app.TemporalClient, err = client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer app.TemporalClient.Close()
	// RabbitMQ connection
	err = rmq.Connect(app.TemporalClient)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rmq.Close()
	// Application entry point
	// RabbitMQ consumer for payment status updates
	g.Go(func() error {
		return rmq.ConsumePaymentStatus(ctx)
	})
	// Payment workflow worker
	g.Go(func() error {
		return payment.RunCompleteSessionWorkflow(ctx, app.TemporalClient)
	})
	// HTTP server
	g.Go(func() error {
		return server.Start(ctx, ":8989")
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Println("server exited with error: ", err)
	}
}
