package app

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/c13n-io/c13n-go/model"
)

type BusError struct {
	op    string
	topic string
	e     error
}

func (be BusError) Error() string {
	return "bus " + be.op + " error on topic " + be.topic + ": " + be.e.Error()
}

func (app *App) publish(topic string, data []byte) error {
	busMsg := message.NewMessage(watermill.NewUUID(), data)
	if err := app.bus.Publish(topic, busMsg); err != nil {
		return BusError{op: "publish", topic: topic, e: err}
	}
	return nil
}

func (app *App) subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	subCh, err := app.bus.Subscribe(ctx, topic)
	if err != nil {
		return subCh, BusError{op: "subscribe", topic: topic, e: err}
	}

	return subCh, nil
}

const (
	// The below constants define the bus event topics.
	messageTopic = "message"
	invoiceTopic = "invoice"
	paymentTopic = "payment"
)

func (app *App) publishMessage(msg model.MessageAggregate) error {
	// Marshal as json for publishing in bus.
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return app.publish(messageTopic, msgBytes)
}

// SubscribeMessages returns a subscription for message notifications.
// The subscriber is responsible for draining the channel
// once the subscription terminates.
func (app *App) SubscribeMessages(ctx context.Context) (<-chan model.MessageAggregate, error) {
	subCh, err := app.subscribe(ctx, messageTopic)
	if err != nil {
		return nil, err
	}

	clientCh := make(chan model.MessageAggregate)
	go func() {
		defer close(clientCh)

		// Forward messages until subscriber exits.
		for subMsg := range subCh {
			subMsg.Ack()

			// Unmarshal message data in a fresh variable.
			msg := new(model.MessageAggregate)
			if err := json.Unmarshal(subMsg.Payload, msg); err != nil {
				e := BusError{op: "subscribe", topic: messageTopic, e: err}
				app.Log.Error(e)
				continue
			}

			clientCh <- *msg
		}
	}()

	return clientCh, nil
}

func (app *App) publishInvoice(inv *model.Invoice) error {
	invBytes, err := json.Marshal(inv)
	if err != nil {
		return BusError{op: "publish", topic: invoiceTopic, e: err}
	}

	return app.publish(invoiceTopic, invBytes)
}

func (app *App) SubscribeInvoices(ctx context.Context) (<-chan *model.Invoice, error) {
	subCh, err := app.subscribe(ctx, invoiceTopic)
	if err != nil {
		return nil, err
	}

	clientCh := make(chan *model.Invoice)
	go func() {
		defer close(clientCh)

		for subMsg := range subCh {
			subMsg.Ack()

			inv := new(model.Invoice)
			if err := json.Unmarshal(subMsg.Payload, inv); err != nil {
				e := BusError{op: "subscribe", topic: invoiceTopic, e: err}
				app.Log.Error(e)
				continue
			}

			clientCh <- inv
		}
	}()
	return clientCh, nil
}

func (app *App) publishPayment(pmnt *model.Payment) error {
	pmntBytes, err := json.Marshal(pmnt)
	if err != nil {
		return BusError{op: "publish", topic: paymentTopic, e: err}
	}

	return app.publish(paymentTopic, pmntBytes)
}

func (app *App) SubscribePayments(ctx context.Context) (<-chan *model.Payment, error) {
	subCh, err := app.subscribe(ctx, paymentTopic)
	if err != nil {
		return nil, err
	}

	clientCh := make(chan *model.Payment)
	go func() {
		defer close(clientCh)

		for subMsg := range subCh {
			subMsg.Ack()

			pmnt := new(model.Payment)
			if err := json.Unmarshal(subMsg.Payload, pmnt); err != nil {
				e := BusError{op: "subscribe", topic: paymentTopic, e: err}
				app.Log.Error(e)
				continue
			}

			clientCh <- pmnt
		}
	}()
	return clientCh, nil
}
