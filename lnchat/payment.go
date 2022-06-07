package lnchat

import (
	"fmt"

	"github.com/lightningnetwork/lnd/lnrpc"
)

// Amount represents an amount on the Lightning network.
type Amount int64

// NewAmount creates an amount from a value in millisatoshi.
func NewAmount(msat int64) Amount {
	return Amount(msat)
}

// Msat returns the amount in millisatoshi.
func (a Amount) Msat() int64 {
	return int64(a)
}

// PayReq represents a request for payment on the Lightning network.
type PayReq struct {
	// The pubkey of the payee.
	Destination NodeID
	// The payment hash of the invoice.
	Hash string
	// The amount to be paid.
	Amt Amount
	// Creation timestamp (in seconds since Unix epoch).
	CreatedTimeSec int64
	// Expiry offset since creation time (in seconds).
	Expiry int64
	// Delta to use for the timelock of the final hop.
	CltvExpiry uint64
	// Route hints that can be used for payment.
	RouteHints []RouteHint
}

func unmarshalPaymentRequest(payReq *lnrpc.PayReq) (*PayReq, error) {
	node, err := NewNodeFromString(payReq.Destination)
	if err != nil {
		return nil, err
	}

	var hints []RouteHint
	if len(payReq.RouteHints) != 0 {
		hints, err = unmarshalRouteHints(payReq.RouteHints)
		if err != nil {
			return nil, err
		}
	}

	return &PayReq{
		Destination:    node,
		Hash:           payReq.PaymentHash,
		Amt:            NewAmount(payReq.NumMsat),
		CreatedTimeSec: payReq.Timestamp,
		Expiry:         payReq.Expiry,
		CltvExpiry:     uint64(payReq.CltvExpiry),
		RouteHints:     hints,
	}, nil
}

// RouteHint is a hint that can assist a payment
// in reaching a destination node.
type RouteHint struct {
	// A list of hop hints that define the route hint.
	HopHints []HopHint
}

// HopHint contains the necessary information (hop id, fee variables)
// for a payment to traverse a route leg.
type HopHint struct {
	// Public key of the node at the channel ingress direction.
	NodeID NodeID
	// The channel id of the channel to be used as hop (short channel id).
	ChanID uint64
	// The base fee of the channel (in millisatoshi).
	FeeBaseMsat uint32
	// The fee rate of the channel (in microsatoshi/sat sent).
	FeeRate uint32
	// The timelock delta of the channel.
	CltvExpiryDelta uint32
}

// Payment represents a Lightning payment.
type Payment struct {
	// The payment hash of the payment.
	Hash string
	// The preimage of the payment hash.
	Preimage string
	// The payment value.
	Value Amount
	// The timestamp of payment creation (in nanoseconds since Unix epoch).
	CreationTimeNs int64
	// The payment request the payment fulfils. May be empty.
	PaymentRequest string
	// The status of the payment.
	Status PaymentStatus
	// The payment index of the payment.
	PaymentIndex uint64
	// The HTLC attempts made to settle the payment.
	Htlcs []HTLCAttempt
}

func (p *Payment) GetDestination() (NodeID, error) {
	var dest NodeID
	switch true {
	case p == nil:
		return dest, fmt.Errorf("nil payment")
	case len(p.Htlcs) == 0:
		return dest, fmt.Errorf("payment contains no HTLCs")
	}

	for _, htlc := range p.Htlcs {
		hops := htlc.Route.Hops
		if len(hops) == 0 {
			return dest, fmt.Errorf("payment contains HTLC without route")
		}
		dest = hops[len(hops)-1].NodeID
	}

	return dest, nil
}

// GetCustomRecords retrieves the custom records
// contained in the successful payment HTLCs.
func (p *Payment) GetCustomRecords() []map[uint64][]byte {
	if p == nil || len(p.Htlcs) == 0 {
		return nil
	}

	var htlcRecords []map[uint64][]byte
	for _, htlc := range p.Htlcs {
		if htlc.Status != lnrpc.HTLCAttempt_SUCCEEDED {
			continue
		}
		lastHop := htlc.Route.Hops[len(htlc.Route.Hops)-1]
		if len(lastHop.CustomRecords) > 0 {
			htlcRecords = append(htlcRecords, lastHop.CustomRecords)
		}
	}

	return htlcRecords
}

// PaymentStatus represents the status of a payment.
type PaymentStatus int32

const (
	// PaymentUNKNOWN denotes the status of a payment as unknown.
	PaymentUNKNOWN PaymentStatus = iota
	// PaymentINFLIGHT signifies that a payment has not been resolved yet.
	PaymentINFLIGHT
	// PaymentSUCCEEDED signifies the success of a payment
	PaymentSUCCEEDED
	// PaymentFAILED signifies the failure of a payment.
	PaymentFAILED
)

// HTLCAttempt represents an attempt to pay (a portion of) a payment.
type HTLCAttempt struct {
	// The route of the HTLC.
	Route Route
	// The time the HTLC was sent (in nanoseconds since Unix epoch).
	AttemptTimeNs int64
	// The time the HTLC was resolved (in nanoseconds since Unix epoch).
	ResolveTimeNs int64
	// The status of the HTLC.
	Status lnrpc.HTLCAttempt_HTLCStatus
	// The details of HTLC failure to settle (if any).
	Failure *HTLCFailure
	// The preimage used to settle the HTLC.
	Preimage []byte
}

// Route represents a route (a list of hops) on the Lightning network
// through which a payment is (partially) fulfilled.
type Route struct {
	// The cumulative (final) time lock across the entire route.
	TimeLock uint32
	// The amount sent via this route, disregarding the fees.
	Amt Amount
	// The fees of this route.
	Fees Amount
	// The list of hops of the route.
	Hops []RouteHop
}

// RouteHop represents a hop of a route.
type RouteHop struct {
	// The (unique) ID of the channel used for the hop.
	ChannelID uint64
	// The pubkey of the hop egress (may be empty).
	NodeID NodeID
	// The amount to be forwarded to the next hop.
	AmtToForward Amount
	// The fees awarded to the egress node to forward the HTLC to the next hop.
	Fees Amount
	// HTLC expiry.
	Expiry uint32
	// TODO: We probably need MPPRecord and AMPRecord as well.
	// The custom records for the hop.
	CustomRecords map[uint64][]byte
}

// HTLCFailure represents a detailed reason for an HTLC failure.
type HTLCFailure struct {
	// The failure code (as defined in BOLT #4, Section "Failure Messages").
	Code lnrpc.Failure_FailureCode
	// The route position of the node that generated the failure.
	// Index 0 represents the sending node.
	NodeIndex uint32
}
