package rmq

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/alifpay/temporal/models"
	"github.com/wagslane/go-rabbitmq"
	"go.temporal.io/sdk/client"
)

var (
	conn           *rabbitmq.Conn
	publisher      *rabbitmq.Publisher
	consumer       *rabbitmq.Consumer
	temporalClient client.Client
)

func Connect(tmp client.Client) error {
	temporalClient = tmp
	var err error
	conn, err = rabbitmq.NewConn(
		"amqp://guest:guest@192.168.215.2:5672",
		rabbitmq.WithConnectionOptionsLogging,
		rabbitmq.WithConnectionOptionsReconnectInterval(5*time.Second),
	)
	if err != nil {
		return err
	}
	publisher, err = rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName("events"),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsConfirm,
	)
	if err != nil {
		return err
	}

	consumer, err = rabbitmq.NewConsumer(
		conn,
		"status_payment_queue",
		rabbitmq.WithConsumerOptionsRoutingKey("status.key"),
		rabbitmq.WithConsumerOptionsExchangeName("events"),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsConsumerName("status_payment_queue"),
		rabbitmq.WithConsumerOptionsQueueDurable,
		rabbitmq.WithConsumerOptionsQOSPrefetch(1),
		rabbitmq.WithConsumerOptionsConsumerAutoAck(false),
	)
	if err != nil {
		return err
	}
	return nil
}

func Close() {
	if conn != nil {
		conn.Close()
	}
	if publisher != nil {
		publisher.Close()
	}
	if consumer != nil {
		consumer.Close()
	}
}

// publish with confirmation
func Publish(ctx context.Context, routingKey string, body []byte) error {
	confirms, err := publisher.PublishWithDeferredConfirmWithContext(
		ctx,
		body,
		[]string{routingKey},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsExchange("events"),
	)
	if err != nil {
		return err
	}
	if len(confirms) == 0 || confirms[0] == nil {
		return errors.New("no confirmation received")
	}
	ok, err := confirms[0].WaitContext(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("message not acknowledged by broker")
	}
	return nil
}

// ConsumePaymentStatus runs the consumer until ctx is canceled.
// On cancellation it closes the consumer so consumer.Run can return.
func ConsumePaymentStatus(ctx context.Context) error {
	if consumer == nil {
		return errors.New("rabbitmq consumer is not initialized")
	}

	// Stop consumer on context cancellation.
	go func() {
		<-ctx.Done()
		if consumer != nil {
			consumer.Close()
		}
	}()

	err := consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		var p models.PaymentStatus
		if err := json.Unmarshal(d.Body, &p); err != nil {
			log.Println("Error unmarshaling payment status:", err)
			return rabbitmq.NackDiscard
		}
		if err := sendSignal(p); err != nil {
			log.Println("Error sending signal to workflow:", err)
			return rabbitmq.NackRequeue
		}
		return rabbitmq.Ack
	})
	if err != nil {
		// When shutting down, Close() will typically make Run() return an error.
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	return nil
}

func sendSignal(p models.PaymentStatus) error {
	err := temporalClient.SignalWorkflow(context.Background(), p.ID, "", models.PaymentStatusSignal, p)
	if err != nil {
		return err
	}
	return nil
}
