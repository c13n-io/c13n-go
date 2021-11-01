package app

import (
	"context"
	"fmt"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/pkg/errors"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
)

// EstimatePayment attempts to calculate the details for
// sending a payment (or message) to a discussion.
// If the discussion contains multiple participants,
// one route for each participant is calculated and the fees are cumulative.
func (app *App) EstimatePayment(ctx context.Context,
	payload string, amtMsat int64, discID uint64,
	opts model.MessageOptions) (*model.Message, error) {

	// Retrieve the requested discussion.
	discussion, err := app.retrieveDiscussion(ctx, discID)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve discussion")
	}

	options := overrideOptions(DefaultOptions, true, discussion.Options, opts)
	payOpts := options.GetPaymentOptions()

	// Disallow anonymous messages in group discussions.
	if len(discussion.Participants) > 1 && options.Anonymous {
		return nil, ErrDiscAnonymousMessage
	}

	// Create a raw message.
	rawMsg, err := app.createRawMessage(ctx, discussion, payload, !options.Anonymous)
	if err != nil {
		return nil, err
	}

	paymentPayload := marshalPayload(rawMsg)

	var totalProb = 1.
	routes := make(map[string]lnchat.Route)
	var errs []error
	for _, recipient := range discussion.Participants {
		route, prob, err := app.LNManager.GetRoute(ctx, recipient,
			lnchat.NewAmount(amtMsat), payOpts, paymentPayload)
		switch err {
		case nil:
			routes[recipient] = *route
		default:
			errs = append(errs, fmt.Errorf("could not "+
				"find route to %s: %w", recipient, err))
		}
		totalProb *= prob
	}

	compositeErr := newCompositeError(errs)

	// This is a slight abuse of the model
	preimage, hash := lntypes.Preimage{}.String(), lntypes.ZeroHash.String()
	var payments []*model.Payment
	for recipient, route := range routes {
		payment := &model.Payment{
			PayerAddress: app.Self.Node.Address,
			PayeeAddress: recipient,
			Payment: lnchat.Payment{
				Preimage: preimage,
				Hash:     hash,
				Value:    lnchat.NewAmount(amtMsat),
				Htlcs: []lnchat.HTLCAttempt{
					{
						Status: lnrpc.HTLCAttempt_SUCCEEDED,
						Route:  route,
					},
				},
			},
		}
		payments = append(payments, payment)
		rawMsg.WithPaymentIndexes(payment.PaymentIndex)
	}

	// When there are no routes return only the errors,
	// for the same reason as in SendPayment.
	if len(payments) == 0 {
		return nil, compositeErr
	}

	msg, err := model.NewOutgoingMessage(rawMsg, false, payments...)
	if err != nil {
		return nil, errors.Wrap(err, "message marshalling failed")
	}

	msg.SuccessProb = totalProb

	return msg, nil
}
