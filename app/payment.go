package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
)

// SendPay attempts to send a payment.
// A payment fulfils the payment request, if one is provided,
// otherwise it is addressed to the discussion participants.
// If payload is present, it is sent along with the payment
// (respecting the provided and discussion options, if applicable).
func (app *App) SendPay(ctx context.Context,
	payload string, amtMsat int64, discID uint64, payReq string,
	opts model.MessageOptions) (*model.Message, error) {

	return app.sendPayment(ctx, payload, amtMsat, discID, payReq, opts)
}

// defaultPaymentFilter is a payment update filter,
// accepting only successful payment updates.
func defaultPaymentFilter(p *lnchat.Payment) bool {
	return p.Status == lnchat.PaymentSUCCEEDED ||
		p.Status == lnchat.PaymentFAILED
}

func (app *App) sendPayment(ctx context.Context,
	payload string, amtMsat int64, discID uint64, payReq string,
	opts model.MessageOptions) (*model.Message, error) {

	// Exactly one of discussion and payment request must be defined.
	if payReq != "" && discID != 0 {
		return nil, fmt.Errorf("exactly one of payment request" +
			" and discussion must be specified")
	}

	// In case a payment request is not provided, retrieve discussion.
	var discussion *model.Discussion
	var err error
	if payReq == "" {
		discussion, err = app.retrieveDiscussion(ctx, discID)
		if err != nil {
			return nil, errors.Wrap(err, "could not retrieve discussion")
		}
	}

	// In case of payment to payment request, decode it and
	// create a discussion with the recipient if one does not exist.
	var payRequest *lnchat.PayReq
	if payReq != "" {
		payRequest, err = app.LNManager.DecodePayReq(ctx, payReq)
		if err != nil {
			return nil, errors.Wrap(err, "could not"+
				" decode payment request")
		}

		discussion, err = app.retrieveOrCreateDiscussion(&model.Discussion{
			Participants: []string{payRequest.Destination.String()},
			Options:      DefaultOptions,
		})
		if err != nil {
			return nil, errors.Wrap(err, "could not retrieve discussion")
		}
	}

	options := overrideOptions(DefaultOptions, true, discussion.Options, opts)
	payOpts := options.GetPaymentOptions()

	// Disallow anonymous messages in group discussions
	if len(discussion.Participants) > 1 && options.Anonymous {
		return nil, ErrDiscAnonymousMessage
	}

	// NOTE: Currently, sending a payment means creating a message
	// even if the payload is empty ("").

	rawMsg, err := app.createRawMessage(ctx, discussion, payload, !options.Anonymous)
	if err != nil {
		return nil, err
	}

	paymentPayload := marshalPayload(rawMsg)

	// Unified logic for payReq and spontaneous payment
	// (possibly to multiple recipients), due to handling
	// recipient and payReq compatibility in lnchat.
	var recipients []string
	switch payReq {
	case "":
		recipients = discussion.Participants
	default:
		recipients = []string{payRequest.Destination.String()}
	}

	// Send payments and retrieve final updates.
	var errs []error
	var payments []*model.Payment
	for _, recipient := range recipients {
		paymentUpdates, err := app.LNManager.SendPayment(ctx,
			recipient, lnchat.NewAmount(amtMsat), payReq, payOpts,
			paymentPayload, defaultPaymentFilter)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not "+
				"initiate payment to %s: %w", recipient, err))
			continue
		}

		// NOTE: Waiting for updates can be handled concurrently.
		update := <-paymentUpdates
		switch {
		case update.Err != nil:
			errs = append(errs, fmt.Errorf("payment error "+
				"for recipient %s: %w", recipient, update.Err))
		case update.Payment != nil:
			payments = append(payments, &model.Payment{
				PayerAddress: app.Self.Node.Address,
				PayeeAddress: recipient,
				Payment:      *update.Payment,
			})
		}
	}

	// Associate only successful payments with the message.
	for _, payment := range payments {
		if payment.Status == lnchat.PaymentSUCCEEDED {
			rawMsg.WithPaymentIndexes(payment.PaymentIndex)
		}
	}

	// Save all payments (irrespective of status).
	if err := app.Database.AddPayments(payments...); err != nil {
		return nil, errors.Wrap(err, "payment storage failed")
	}

	// Store raw message, if it has associated payments
	if len(rawMsg.PaymentIndexes) <= 0 {
		if err := newCompositeError(errs); err != nil {
			return nil, fmt.Errorf("failed to send message: %w", err)
		}
		return nil, fmt.Errorf("failed to send message")
	}

	if err := app.Database.AddRawMessage(rawMsg); err != nil {
		return nil, errors.Wrap(err, "message storage failed")
	}

	msg, err := model.NewOutgoingMessage(rawMsg, true, payments...)
	if err != nil {
		return nil, errors.Wrap(err, "message marshalling failed")
	}

	return msg, nil
}

func (app *App) createRawMessage(ctx context.Context, discussion *model.Discussion,
	payload string, withSig bool) (*model.RawMessage, error) {

	rawMsg, err := model.NewRawMessage(discussion, payload)
	if err != nil {
		return nil, errors.Wrap(err, "could not create raw message")
	}
	if withSig {
		sig, err := app.LNManager.SignMessage(ctx, rawMsg.RawPayload)
		if err != nil {
			return nil, errors.Wrap(err, "could not sign message payload")
		}
		err = rawMsg.WithSignature(app.Self.Node.Address, sig)
		if err != nil {
			return nil, errors.Wrap(err, "could not add signature to raw message")
		}
	}

	return rawMsg, nil
}

func newCompositeError(es []error) error {
	if len(es) == 0 {
		return nil
	}

	err := &multierror.Error{
		ErrorFormat: func(errs []error) string {
			es := make([]string, len(errs))
			for i, e := range errs {
				es[i] = e.Error()
			}
			return strings.Join(es, "; ")
		},
		Errors: es,
	}

	return err
}
