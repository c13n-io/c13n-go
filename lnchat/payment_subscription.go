package lnchat

import (
	"context"
	"io"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var maxPaymentsPerRequest uint64 = 100

func resolveContextError(err error) error {
	if err == nil {
		return nil
	}

	if status.Code(err) == codes.Canceled {
		return context.Canceled
	}
	return err
}

// SubscribePaymentUpdates creates and returns a channel
// over which payment updates are received.
//
// The returned updates are those for which filter is true.
// If startIdx is provided (non-zero), only updates for
// payments after that payment index are returned.
func (m *manager) SubscribePaymentUpdates(ctx context.Context, startIdx uint64,
	filter PaymentUpdateFilter) (<-chan PaymentUpdate, error) {

	// Since lnd does not offer a payment subscription method, emulate
	// by retrieving the payment list from the last known payment index
	// and tracking all unknown payments, and repeating
	// upon notification of an outgoing attempt
	// (HTLC ForwardEvent of type HtlcEvent_SEND).

	// Retrieves all payments after last known index.
	// Retrieval is performed in batches of maxPaymentsPerRequest,
	// and payments are returned in ascending order.
	retrievePayments := func(ctx context.Context,
		lastKnownIdx uint64) ([]*lnrpc.Payment, error) {

		var payments []*lnrpc.Payment

		req := &lnrpc.ListPaymentsRequest{
			IncludeIncomplete: true,
			MaxPayments:       maxPaymentsPerRequest,
			IndexOffset:       lastKnownIdx,
		}

		for {
			resp, err := m.lnClient.ListPayments(ctx, req)
			if err != nil {
				return nil, resolveContextError(err)
			}

			payments = append(payments, resp.GetPayments()...)

			req.IndexOffset = resp.GetLastIndexOffset()

			if uint64(len(resp.GetPayments())) < req.MaxPayments {
				break
			}
		}

		return payments, nil
	}

	// Tracks updates for payment corresponding to provided hash,
	// and sends updates to the provided channel.
	trackPayment := func(ctx context.Context,
		paymentHash string, updateCh chan<- *lnrpc.Payment) error {

		hash, err := lntypes.MakeHashFromStr(paymentHash)
		if err != nil {
			return err
		}

		req := &routerrpc.TrackPaymentRequest{
			PaymentHash: hash[:],
		}
		updates, err := m.routeClient.TrackPaymentV2(ctx, req)
		if err != nil {
			return resolveContextError(err)
		}

		for {
			update, err := updates.Recv()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return resolveContextError(err)
			}
			updateCh <- update
		}
	}

	g, ctxt := errgroup.WithContext(ctx)

	htlcReq := &routerrpc.SubscribeHtlcEventsRequest{}
	htlcEventStream, err := m.routeClient.SubscribeHtlcEvents(ctxt, htlcReq)
	if err != nil {
		return nil, errors.Wrap(err, "could not create htlc event subscription")
	}

	paymentUpdatesCh := make(chan PaymentUpdate)
	go func() {
		// Upon termination, wait for the goroutines
		// and close the update channel.
		defer func() {
			if err := g.Wait(); err != nil {
				paymentUpdatesCh <- PaymentUpdate{nil, err}
			}
			close(paymentUpdatesCh)
		}()

		// Publish htlc events on event channel.
		htlcEvents := func() <-chan bool {
			htlcEventCh := make(chan bool, 1)

			// Send an initial event to trigger
			// tracking of any updates since last known state.
			htlcEventCh <- true

			g.Go(func() error {
				defer close(htlcEventCh)
				for {
					event, err := htlcEventStream.Recv()
					if err != nil {
						return resolveContextError(err)
					}

					if event.GetEventType() == routerrpc.HtlcEvent_SEND &&
						event.GetForwardEvent() != nil {

						htlcEventCh <- true
					}
				}
			})

			return htlcEventCh
		}()

		// Create a channel to aggregate all papyment tracking updates.
		trackingUpdatesCh := make(chan *lnrpc.Payment)
		defer close(trackingUpdatesCh)

		lastKnownIdx := startIdx
		for {
			select {
			case <-ctxt.Done():
				// Terminate when the context is finished.
				return
			case _, ok := <-htlcEvents:
				// Terminate if the event channel is closed.
				if !ok {
					return
				}

				// Poll for new payments upon htlc event.

				// Retrieve all payments after the last known one.
				payments, err := retrievePayments(ctxt, lastKnownIdx)
				if err := resolveContextError(err); err != nil {
					if err != context.Canceled {
						paymentUpdatesCh <- PaymentUpdate{nil, err}
					}
					return
				}

				// Track new payments.
				for _, p := range payments {
					idx, hash := p.GetPaymentIndex(), p.GetPaymentHash()
					if idx <= lastKnownIdx {
						continue
					}
					lastKnownIdx = idx

					g.Go(func() error {
						return trackPayment(ctxt, hash, trackingUpdatesCh)
					})
				}
			case update, _ := <-trackingUpdatesCh:
				payUpdate, err := unmarshalPayment(update)
				if err != nil {
					paymentUpdatesCh <- PaymentUpdate{nil, err}
				}

				// Propagate the update if it matches the filter.
				if !filter(payUpdate) {
					continue
				}

				paymentUpdatesCh <- PaymentUpdate{payUpdate, nil}
			}
		}
	}()

	return paymentUpdatesCh, nil
}
