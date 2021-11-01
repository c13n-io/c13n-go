package app

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/c13n-io/c13n-go/model"
)

const (
	// ReceiveTopic is the pubsub topic for received messages.
	ReceiveTopic = "message.receive"
)

// publishMessage publishes a message.
func (app *App) publishMessage(msg *model.Message) error {
	// Marshal message as json for publishing
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return Error{Kind: MarshalError, details: "publishMessage", Err: err}
	}

	// Publish message under message topic
	pubMsg := message.NewMessage(watermill.NewUUID(), msgBytes)
	if err := app.PubSubBus.Publish(ReceiveTopic, pubMsg); err != nil {
		return Error{Kind: InternalError, details: "Publish error", Err: err}
	}

	return nil
}

// SubscribeMessages returns a channel over which received messages are sent.
func (app *App) SubscribeMessages(ctx context.Context) (<-chan *message.Message, error) {
	msgCh, err := app.PubSubBus.Subscribe(ctx, ReceiveTopic)
	if err != nil {
		return msgCh, Error{Kind: InternalError, details: "Subscribe error", Err: err}
	}

	return msgCh, nil
}
