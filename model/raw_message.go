package model

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/pkg/errors"

	"github.com/c13n-io/c13n-backend/lnchat"
)

// RawMessage represents a raw message over the Lightning network,
// associated with a payment or an invoice - depending on whether it
// was incoming or outgoing.
// Exactly one of InvoiceSettleIndex or PaymentIndex must be populated.
type RawMessage struct {
	// The message id (store index).
	ID uint64 `badgerholdKey:"key"`
	// The id of the discussion the message belongs to.
	DiscussionID uint64 `badgerholdIndex:"DiscIdx"`
	// The raw message payload.
	RawPayload []byte
	// The Lightning address of the sender.
	// TODO: Replace Sender field type with lnchat.NodeID.
	Sender string
	// The message signature.
	Signature []byte
	// Whether the sender was the one that signed the payload.
	SignatureVerified bool
	// The SettleIndex of the invoice associated
	// with the message (incoming).
	InvoiceSettleIndex uint64
	// The PaymentIndexes of the payments
	// used to transport the message (outgoing).
	PaymentIndexes []uint64
	// The timestamp of the message.
	// It  is an internal field and does not correspond
	// to the sent or received time of the message.
	Timestamp time.Time
}

type compositePayload struct {
	Participants []string `json:"participants"`
	Message      string   `json:"message"`
}

// UnmarshalPayload returns the message participants.
func (raw *RawMessage) UnmarshalPayload() (string, []string, error) {
	msg := &compositePayload{}
	if raw.RawPayload != nil {
		if err := json.Unmarshal(raw.RawPayload, msg); err != nil {
			return "", nil, err
		}
	}

	return msg.Message, msg.Participants, nil
}

// NewRawMessage constructs a raw message from a discussion and payload.
func NewRawMessage(discussion *Discussion, payload string) (*RawMessage, error) {
	if discussion == nil {
		return nil, fmt.Errorf("cannot create raw message for empty discussion")
	}

	rawMsg := new(RawMessage)

	data, err := json.Marshal(compositePayload{
		Participants: discussion.Participants,
		Message:      payload,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal payload")
	}

	rawMsg.RawPayload = data
	rawMsg.DiscussionID = discussion.ID

	return rawMsg, nil
}

// WithTimestamp adds the provided timestamp to the raw message.
func (raw *RawMessage) WithTimestamp(ts time.Time) {
	raw.Timestamp = ts
}

// WithSignature adds the provided sender and signature to the raw message.
func (raw *RawMessage) WithSignature(sender string, signature []byte) error {
	if sender == "" || signature == nil {
		return fmt.Errorf("signature and sender address" +
			" are required for payload signing")
	}

	if _, err := lnchat.NewNodeFromString(sender); err != nil {
		return fmt.Errorf("invalid sender address: %w", err)
	}

	raw.Sender = sender
	raw.Signature = signature
	raw.SignatureVerified = (len(signature) != 0)

	return nil
}

// WithPaymentIndexes associates the provided payment indexes with the message.
func (raw *RawMessage) WithPaymentIndexes(paymentIdxs ...uint64) {
	raw.PaymentIndexes = append(raw.PaymentIndexes, paymentIdxs...)
}

func (raw *RawMessage) containsPaymentIndex(paymentIdx uint64) bool {
	for _, idx := range raw.PaymentIndexes {
		if idx == paymentIdx {
			return true
		}
	}

	return false
}

// NewIncomingMessage constructs a Message from a RawMessage and Invoice.
func NewIncomingMessage(rawMsg *RawMessage, inv *Invoice,
	discussionRetriever func([]string) (*Discussion, error)) (*Message, error) {

	if rawMsg == nil || inv == nil {
		return nil, fmt.Errorf("cannot construct message: " +
			"raw message or invoice missing")
	}
	if rawMsg.InvoiceSettleIndex != inv.SettleIndex {
		return nil, fmt.Errorf("cannot construct message: " +
			"provided invoice does not correspond to raw message")
	}

	preimageHash, err := preimageHashBytes(inv.Hash)
	if err != nil {
		return nil, err
	}
	preimage, err := preimageFromBytes(inv.Preimage)
	if err != nil {
		return nil, err
	}

	// Unmarshal the message payload.
	payload, participants, err := rawMsg.UnmarshalPayload()
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal payload: %w", err)
	}

	// NOTE: Since the sender address is not included
	// in the embedded participant set, in order to
	// retrieve the discussion it must be included in the participant set.
	fullParticipantSet := participants
	if rawMsg.Sender != "" {
		fullParticipantSet = append(fullParticipantSet, rawMsg.Sender)
	}

	// Retrieve the discussion id and verify the message signature.
	disc, err := discussionRetriever(fullParticipantSet)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve discussion ID: %w", err)
	}

	return &Message{
		ID:             rawMsg.ID,
		DiscussionID:   disc.ID,
		Payload:        payload,
		AmtMsat:        inv.AmtPaid.Msat(),
		Sender:         rawMsg.Sender,
		Receiver:       inv.CreatorAddress,
		SenderVerified: rawMsg.SignatureVerified,
		SentTimeNs:     time.Unix(inv.CreatedTimeSec, 0).UnixNano(),
		ReceivedTimeNs: time.Unix(inv.SettleTimeSec, 0).UnixNano(),
		Index:          inv.SettleIndex,
		PreimageHash:   preimageHash,
		Preimage:       preimage,
		PayReq:         inv.PaymentRequest,
		SuccessProb:    1.,
	}, nil
}

func preimageHashBytes(hash string) ([]byte, error) {
	hb, err := lntypes.MakeHashFromStr(hash)
	return hb[:], err
}

func preimageFromBytes(preimage []byte) (lnchat.PreImage, error) {
	return lntypes.MakePreimage(preimage)
}

func preimageFromString(preimage string) (lnchat.PreImage, error) {
	return lntypes.MakePreimageFromStr(preimage)
}

// NewOutgoingMessage constructs a Message from a RawMessage and a list of Payments.
func NewOutgoingMessage(rawMsg *RawMessage, onlySuccessfulPayments bool,
	payments ...*Payment) (*Message, error) {

	// Filter unsuccessful payments if requested.
	usedPayments := payments
	if onlySuccessfulPayments {
		usedPayments = func(ps []*Payment) []*Payment {
			var res []*Payment
			for _, p := range ps {
				if p.Status != lnchat.PaymentSUCCEEDED {
					continue
				}
				res = append(res, p)
			}
			return res
		}(payments)
	}

	if rawMsg == nil || len(usedPayments) <= 0 {
		return nil, fmt.Errorf("raw message or payments missing")
	}
	sender := payments[0].PayerAddress
	for _, payment := range usedPayments {
		if !rawMsg.containsPaymentIndex(payment.PaymentIndex) {
			return nil, fmt.Errorf("provided payment does not" +
				" correspond to raw message")
		}
		if payment.PayerAddress != sender {
			return nil, fmt.Errorf("mismatched" +
				" payer addresses in payments")
		}
	}

	return newOutgoingMessage(rawMsg, usedPayments...)
}

func newOutgoingMessage(rawMsg *RawMessage, payments ...*Payment) (*Message, error) {
	payload, _, err := rawMsg.UnmarshalPayload()
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal raw message")
	}

	var (
		sender                 = payments[0].PayerAddress
		amt, fees              int64
		startTimeNs, endTimeNs int64
		routes                 []Route
	)
	startTimeNs, endTimeNs = math.MaxInt64, math.MinInt64
	for _, payment := range payments {
		if payment.CreationTimeNs < startTimeNs {
			startTimeNs = payment.CreationTimeNs
		}
		for _, htlc := range payment.Htlcs {
			if htlc.Status != lnrpc.HTLCAttempt_SUCCEEDED {
				continue
			}
			if htlc.ResolveTimeNs > endTimeNs {
				endTimeNs = htlc.ResolveTimeNs
			}

			route := htlc.Route

			amt += route.Amt.Msat()
			fees += route.Fees.Msat()

			routes = append(routes, newRoute(route))
		}

	}

	// These only make sense being populated in single-payment messages.
	var (
		recipient, payReq string
		paymentIdx        uint64
		preimageHash      []byte
		preimage          lnchat.PreImage
	)
	if len(payments) == 1 {
		payment := payments[0]

		recipient = payment.PayeeAddress
		payReq = payment.PaymentRequest
		paymentIdx = payment.PaymentIndex
		if preimageHash, err = preimageHashBytes(payment.Hash); err != nil {
			return nil, errors.Wrap(err, "could not construct message hash")
		}
		if preimage, err = preimageFromString(payment.Preimage); err != nil {
			return nil, errors.Wrap(err, "could not construct message preimage")
		}
	}

	return &Message{
		ID:             rawMsg.ID,
		DiscussionID:   rawMsg.DiscussionID,
		Payload:        payload,
		AmtMsat:        amt,
		Sender:         sender,
		Receiver:       recipient,
		SenderVerified: rawMsg.SignatureVerified,
		SentTimeNs:     startTimeNs,
		ReceivedTimeNs: endTimeNs,
		Index:          paymentIdx,
		TotalFeesMsat:  fees,
		Routes:         routes,
		PreimageHash:   preimageHash,
		Preimage:       preimage,
		PayReq:         payReq,
	}, nil
}

func newRoute(route lnchat.Route) Route {
	hops := make([]Hop, len(route.Hops))
	for i, hop := range route.Hops {
		hops[i] = Hop{
			ChanID:           hop.ChannelID,
			HopAddress:       hop.NodeID.String(),
			AmtToForwardMsat: hop.AmtToForward.Msat(),
			FeeMsat:          hop.Fees.Msat(),
			CustomRecords:    hop.CustomRecords,
		}
	}

	return Route{
		TotalTimeLock: route.TimeLock,
		RouteAmtMsat:  route.Amt.Msat(),
		RouteFeesMsat: route.Fees.Msat(),
		RouteHops:     hops,
	}
}
