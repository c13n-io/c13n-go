package app

import (
	"context"
	"fmt"
	"time"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
)

// defaultInvoiceFilter is an invoice update filter,
// accepting only settled invoice updates.
func defaultInvoiceFilter(inv *lnchat.Invoice) bool {
	return inv.State == lnchat.InvoiceSETTLED
}

// verifySignature verifies the signature over the message and asserts that the
// recovered public key matches the provided address.
func (app *App) verifySignature(ctx context.Context,
	msg, sig []byte, senderAddr string) (bool, error) {

	ctxt, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	addr, err := app.LNManager.VerifySignatureExtractPubkey(ctxt, msg, sig)

	// A payload is considered verified if
	// the sender address and the signing address match.
	// Obviously, if the sender address does not exist,
	// there is no verification to be done.
	return senderAddr != "" && addr == senderAddr, err
}

func (app *App) subscribeInvoices(ctx context.Context, lastInvoiceIdx uint64) error {
	// Create subscription for invoice updates
	invSubscription, err := app.LNManager.SubscribeInvoiceUpdates(ctx,
		lastInvoiceIdx, defaultInvoiceFilter)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			// If the context is finished, terminate
			return nil
		case invUpdate, ok := <-invSubscription:
			// If the subscription channel is closed,
			// terminate with error.
			if !ok {
				return fmt.Errorf("subscription channel closed")
			}
			inv, err := invUpdate.Inv, invUpdate.Err
			if err != nil {
				// TODO: EOF in case of disconnect?
				return fmt.Errorf("invoice update failed: %w", invUpdate.Err)
			}

			// Store the invoice, regardless of payload being present.
			invoice := &model.Invoice{
				CreatorAddress: app.Self.Node.Address,
				Invoice:        *inv,
			}
			if err = app.Database.AddInvoice(invoice); err != nil {
				app.Log.WithError(err).Error("invoice storage failed")
			}

			// Extract a raw message, if one exists.
			rawMsg, err := payloadExtractor(inv,
				func(msg, sig []byte, sender string) (bool, error) {
					return app.verifySignature(ctx, msg, sig, sender)
				},
			)
			if err != nil {
				app.Log.WithError(err).Info("message extraction failed")
			}

			// If no payload exists, there is nothing more
			// to be done on this iteration.
			if rawMsg == nil {
				continue
			}

			// Retrieve (or create) the appropriate discussion,
			// and store the message.
			var disc *model.Discussion
			disc, err = app.retrieveOrCreateRawMsgDiscussion(rawMsg)
			if err != nil {
				app.Log.WithError(err).Error("discussion retrieval failed")
				continue
			}

			// Store invoice (and raw message, if present).
			rawMsg.DiscussionID = disc.ID
			if err := app.Database.AddRawMessage(rawMsg); err != nil {
				app.Log.WithError(err).Error("message storage failed")
				continue
			}

			// Publish the message to the appropriate topic.
			retrieveDisc := func(_ []string) (*model.Discussion, error) {
				return disc, nil
			}
			msg, err := model.NewIncomingMessage(rawMsg, invoice, retrieveDisc)
			if err != nil {
				app.Log.WithError(err).Error("message unmarshalling failed")
				continue
			}

			if err := app.publishMessage(msg); err != nil {
				app.Log.WithError(err).Error("message publish failed")
				continue
			}
		}
	}

	return nil
}

func (app *App) retrieveOrCreateRawMsgDiscussion(raw *model.RawMessage) (
	*model.Discussion, error) {

	_, participants, err := raw.UnmarshalPayload()
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve message participant set: %w", err)
	}

	// NOTE: Since currently our node address is included
	// in the participant set of an incoming message,
	// but not in the participant set of a discussion, remove it.
	selfAddrIdx := -1
	for i, p := range participants {
		if p == app.Self.Node.Address {
			selfAddrIdx = i
			break
		}
	}
	var trimmedParticipants []string
	switch selfAddrIdx {
	case -1:
		trimmedParticipants = participants
	default:
		trimmedParticipants = append(trimmedParticipants,
			participants[:selfAddrIdx]...)
		trimmedParticipants = append(trimmedParticipants,
			participants[selfAddrIdx+1:]...)
	}

	// If the sender is identified, add them to the participant set.
	if raw.Sender != "" {
		trimmedParticipants = append(trimmedParticipants, raw.Sender)
	}

	return app.retrieveOrCreateDiscussion(&model.Discussion{
		Participants: trimmedParticipants,
		Options:      DefaultOptions,
	})
}
