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
	// messageTopic is the topic where message events are published.
	messageTopic = "message"
)

func (app *App) publishMessage(msg *model.Message) error {
	// Marshal as json for publishing in bus.
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return app.publish(messageTopic, msgBytes)
}

type MaybeMessage struct {
	Message *model.Message
	Error   error
}

// SubscribeMessages returns a channel over which received messages are sent.
// The subscriber is responsible for draining the channel
// once the subscription terminates.
func (app *App) SubscribeMessages(ctx context.Context) (<-chan MaybeMessage, error) {
	subCh, err := app.subscribe(ctx, messageTopic)
	if err != nil {
		return nil, err
	}

	msgCh := make(chan MaybeMessage)
	go func() {
		defer close(msgCh)

		// Forward messages until subscriber exits.
		for subMsg := range subCh {
			subMsg.Ack()

			// Unmarshal message data in a fresh variable.
			msg := new(model.Message)
			err := json.Unmarshal(subMsg.Payload, msg)

			msgCh <- MaybeMessage{
				Message: msg,
				Error:   err,
			}
		}
	}()

	return msgCh, nil
}
