package lnchat

import (
	"fmt"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/record"
)

func copyCustomRecords(customRecords map[uint64][]byte,
	preimage []byte) map[uint64][]byte {

	records := make(map[uint64][]byte)
	for k, v := range customRecords {
		records[k] = v
	}

	if preimage != nil {
		records[record.KeySendType] = preimage[:]
	}

	return records
}

func createSendPaymentRequest(dest []byte, amtMsat int64, payReq string,
	customRecords map[uint64][]byte, options PaymentOptions) (
	*routerrpc.SendPaymentRequest, error) {

	var finalCltvDelta int32
	var preimage, paymentHash []byte

	// If a payment request is not provided, the preimage must
	// be set in the custom records (spontaneous payment).
	if payReq == "" {
		preimg, err := generatePreimage()
		if err != nil {
			return nil, err
		}
		hash := preimg.Hash()

		preimage = preimg[:]
		paymentHash = hash[:]

		finalCltvDelta = options.FinalCltvDelta
	}

	request := &routerrpc.SendPaymentRequest{
		Dest:           dest,
		AmtMsat:        amtMsat,
		PaymentRequest: payReq,
		PaymentHash:    paymentHash,
		FinalCltvDelta: finalCltvDelta,
		FeeLimitMsat:   options.FeeLimitMsat,
		TimeoutSeconds: options.TimeoutSecs,
		DestFeatures: []lnrpc.FeatureBit{
			lnrpc.FeatureBit_TLV_ONION_OPT,
		},
	}

	records := copyCustomRecords(customRecords, preimage)
	if len(records) > 0 {
		request.DestCustomRecords = records
	}

	return request, nil
}

// unmarshalPayment creates an lnchat.Payment from an lnrpc.Payment.
func unmarshalPayment(p *lnrpc.Payment) (*Payment, error) {
	if p == nil {
		return nil, fmt.Errorf("cannot unmarshal nil payment")
	}

	payment := &Payment{
		Hash:           p.PaymentHash,
		Preimage:       p.PaymentPreimage,
		Value:          NewAmount(p.ValueMsat),
		PaymentRequest: p.PaymentRequest,
		CreationTimeNs: p.CreationTimeNs,
		PaymentIndex:   p.PaymentIndex,
	}

	switch p.Status {
	case lnrpc.Payment_UNKNOWN:
		payment.Status = PaymentUNKNOWN
	case lnrpc.Payment_IN_FLIGHT:
		payment.Status = PaymentINFLIGHT
	case lnrpc.Payment_SUCCEEDED:
		payment.Status = PaymentSUCCEEDED
	case lnrpc.Payment_FAILED:
		payment.Status = PaymentFAILED
	}

	paymentHtlcs := make([]HTLCAttempt, len(p.Htlcs))
	for i, htlc := range p.Htlcs {
		paymentHtlc, err := unmarshalHtlcAttempt(htlc)
		if err != nil {
			return nil, err
		}
		paymentHtlcs[i] = paymentHtlc
	}
	if len(p.Htlcs) != 0 {
		payment.Htlcs = paymentHtlcs
	}

	return payment, nil
}

// unmarshalHtlcAttempt creates an lnchat.HTLCAttempt
// from the corresponding lnrpc type.
func unmarshalHtlcAttempt(htlc *lnrpc.HTLCAttempt) (HTLCAttempt, error) {
	if htlc == nil {
		return HTLCAttempt{}, fmt.Errorf("cannot unmarshal nil HTLC attempt")
	}

	htlcRoute, err := unmarshalRoute(htlc.Route)
	if err != nil {
		return HTLCAttempt{}, err
	}

	attempt := HTLCAttempt{
		AttemptTimeNs: htlc.AttemptTimeNs,
		ResolveTimeNs: htlc.ResolveTimeNs,
		Status:        htlc.Status,
		Preimage:      htlc.Preimage,
		Route:         *htlcRoute,
	}

	if htlc.Failure != nil {
		attempt.Failure = &HTLCFailure{
			Code:      htlc.Failure.Code,
			NodeIndex: htlc.Failure.FailureSourceIndex,
		}
	}

	return attempt, nil
}

// unmarshalRoute creates an lnchat.Route from an lnrpc.Route.
func unmarshalRoute(route *lnrpc.Route) (*Route, error) {
	if route == nil {
		return nil, fmt.Errorf("cannot unmarshal empty route")
	}

	routeHops := make([]RouteHop, len(route.Hops))
	for i, hop := range route.Hops {
		hopNode, err := NewNodeFromString(hop.PubKey)
		if err != nil {
			return nil, fmt.Errorf("cannot decode hop address %s", hop.PubKey)
		}

		routeHops[i] = RouteHop{
			ChannelID:     hop.ChanId,
			NodeID:        hopNode,
			AmtToForward:  NewAmount(hop.AmtToForwardMsat),
			Fees:          NewAmount(hop.FeeMsat),
			Expiry:        hop.Expiry,
			CustomRecords: hop.CustomRecords,
		}
	}

	return &Route{
		TimeLock: route.TotalTimeLock,
		Amt:      NewAmount(route.TotalAmtMsat - route.TotalFeesMsat),
		Fees:     NewAmount(route.TotalFeesMsat),
		Hops:     routeHops,
	}, nil
}
