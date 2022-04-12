package app

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
)

// SendMessage attempts to send a message.
// If a payment request is provided, a discussion with the recipient
// is created with default options if it does not exist.
// Note: Anonymous messages to group discussions are disallowed.
func (app *App) SendMessage(ctx context.Context, discID uint64, amtMsat int64, payReq string,
	payload string, opts model.MessageOptions) (*model.RawMessage, []*model.Payment, error) {

	// Validate arguments
	if (payReq != "") && (discID != 0) {
		return nil, nil, fmt.Errorf("exactly one of payment request " +
			"and discussion must be specified")
	}

	// Retrieve discussion
	var err error
	var disc *model.Discussion
	switch payReq {
	default:
		// Create discussion if it does not exist
		payRequest, err := app.LNManager.DecodePayReq(ctx, payReq)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not decode payment request")
		}
		disc, err = app.retrieveOrCreateDiscussion(&model.Discussion{
			Participants: []string{payRequest.Destination.String()},
			Options:      DefaultOptions,
		})
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not retrieve discussion")
		}
	case "":
		disc, err = app.retrieveDiscussion(ctx, discID)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not retrieve discussion")
		}
	}

	// Disallow anonymous messages in group discussions
	options := overrideOptions(disc.Options, true, opts)
	if len(disc.Participants) > 1 && options.Anonymous {
		return nil, nil, ErrDiscAnonymousMessage
	}
	payOpts := options.GetPaymentOptions()

	// Create raw message
	rawMsg, err := app.createRawMessage(ctx, disc, payload, !options.Anonymous)
	if err != nil {
		return nil, nil, err
	}
	tlvs := marshalPayload(rawMsg)

	// Perform payment attempts in parallel
	var wg sync.WaitGroup
	var results sync.Map
	for _, recipient := range disc.Participants {
		wg.Add(1)
		go func(dest string) {
			defer wg.Done()

			result, err := app.send(ctx, dest, amtMsat, payReq, payOpts, tlvs)
			if err != nil {
				results.Store(dest, err)
				return
			}
			results.Store(dest, result)
		}(recipient)
	}

	wg.Wait()

	// Aggregate payments and errors
	var errs []error
	payments := make([]*model.Payment, 0, len(disc.Participants))
	results.Range(func(key, val interface{}) bool {
		recipient, ok := key.(string)
		if !ok {
			return false
		}

		switch v := val.(type) {
		case lnchat.PaymentUpdate:
			if v.Err != nil {
				errs = append(errs, fmt.Errorf("payment error "+
					"for recipient %s: %w", recipient, v.Err))
				break
			}
			payments = append(payments, &model.Payment{
				PayerAddress: app.Self.Node.Address,
				PayeeAddress: recipient,
				Payment:      *v.Payment,
			})
		case error:
			errs = append(errs, fmt.Errorf("could not initiate "+
				"payment to recipient %s: %w", recipient, err))
		}
		return true
	})

	// If there are no payments associated with the message, fail early
	if len(payments) == 0 {
		return nil, nil, newCompositeError(errs)
	}

	// Associate payments with the message
	for _, payment := range payments {
		rawMsg.WithPaymentIndexes(payment.PaymentIndex)
	}

	// Store payments and raw message
	if err := app.Database.AddPayments(payments...); err != nil {
		return rawMsg, payments, errors.Wrap(err, "could not store payments")
	}
	if err := app.Database.AddRawMessage(rawMsg); err != nil {
		return rawMsg, payments, errors.Wrap(err, "could not store message")
	}

	return rawMsg, payments, newCompositeError(errs)
}

// defaultPaymentFilter is a payment update filter,
// accepting only successful payment updates.
func defaultPaymentFilter(p *lnchat.Payment) bool {
	return p.Status == lnchat.PaymentSUCCEEDED ||
		p.Status == lnchat.PaymentFAILED
}

// SendPayment attempts to send a payment.
func (app *App) SendPayment(ctx context.Context,
	dest string, amtMsat int64, payReq string,
	opts lnchat.PaymentOptions, tlvs map[uint64][]byte) (*model.Payment, error) {

	// Validate arguments
	if (payReq != "") == (dest != "") {
		return nil, fmt.Errorf("exactly one of payment request " +
			"and destination address must be specified")
	}

	recipient := dest

	if payReq != "" {
		decodedPayReq, err := app.LNManager.DecodePayReq(ctx, payReq)
		if err != nil {
			return nil, err
		}
		recipient = decodedPayReq.Destination.String()
	}

	// Perform payment attempt
	result, err := app.send(ctx, dest, amtMsat, payReq, opts, tlvs)
	if err != nil {
		return nil, err
	}

	if result.Err != nil {
		return nil, result.Err
	}

	payment := &model.Payment{
		PayerAddress: app.Self.Node.Address,
		PayeeAddress: recipient,
		Payment:      *result.Payment,
	}

	// Store payment
	if err := app.Database.AddPayments(payment); err != nil {
		return payment, err
	}

	return payment, nil
}

// Attempts payment and returns the final payment update.
func (app *App) send(ctx context.Context, dest string, amtMsat int64, payReq string,
	opts lnchat.PaymentOptions, tlvs map[uint64][]byte) (lnchat.PaymentUpdate, error) {

	var lastUpdate lnchat.PaymentUpdate

	amt := lnchat.NewAmount(amtMsat)
	updates, err := app.LNManager.SendPayment(ctx, dest, amt, payReq,
		opts, tlvs, defaultPaymentFilter)
	if err != nil {
		return lastUpdate, err
	}

	for update := range updates {
		lastUpdate = update
	}

	return lastUpdate, nil
}

// SendPay attempts to send a payment.
// A payment fulfils the payment request, if one is provided,
// otherwise it is addressed to the discussion participants.
// If payload is present, it is sent along with the payment
// (respecting the provided and discussion options, if applicable).
func (app *App) SendPay(ctx context.Context,
	payload string, amtMsat int64, discID uint64, payReq string,
	opts model.MessageOptions) (*model.Message, error) {

	rawMsg, payments, err := app.SendMessage(ctx,
		discID, amtMsat, payReq, payload, opts)
	if err != nil {
		return nil, err
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
