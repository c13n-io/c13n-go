package model

import "github.com/c13n-io/c13n-go/lnchat"

// Hop represents a hop in a payment route.
type Hop struct {
	// The channel id.
	ChanID uint64 `json:"chan_id"`
	// The address (public key) of the hop.
	HopAddress string `json:"hop_address"`
	// The amount to be forwarded through the hop (in millisatoshi).
	AmtToForwardMsat int64 `json:"amt_to_forward_msat"`
	// The fee to be paid to the hop (in millisatoshi).
	FeeMsat int64 `json:"fee_msat"`
	// The custom records for the hop.
	CustomRecords map[uint64][]byte `json:"custom_records"`
}

// Route represents a payment route.
type Route struct {
	// The total timelock across the entire route.
	TotalTimeLock uint32 `json:"total_timelock"`
	// The amount sent via this route, disregarding the route fees (in millisatoshi).
	RouteAmtMsat int64 `json:"route_amt_msat"`
	// The total route fees (in millisatoshi).
	RouteFeesMsat int64 `json:"route_fees_msat"`
	// The list of hops for the route.
	RouteHops []Hop `json:"route_hops"`
}

// Message represents a message.
type Message struct {
	// The id of the message.
	ID uint64 `json:"id" badgerhold:"key"`
	// The id of the discussion the message is part of.
	DiscussionID uint64 `json:"discussion_id"`
	// The message payload.
	Payload string `json:"payload"`
	// The total amount sent with this message across all routes (in millisatoshi).
	AmtMsat int64 `json:"amt_msat"`
	// The address of the sender, if present.
	Sender string `json:"sender"`
	// The address of the recipient.
	Receiver string `json:"receiver"`
	// Whether the sender was verified via signature.
	SenderVerified bool `json:"sender_verified"`
	// The time the message was sent (in nanoseconds since Unix Epoch).
	SentTimeNs int64 `json:"sent_time_ns"`
	// The time the message was received (in nanoseconds since Unix Epoch).
	ReceivedTimeNs int64 `json:"received_time_ns"`
	// The message index (settle_index for received messages,
	// payment_index for sent messages), as it relates to lnd.
	Index uint64
	// The total fees used to send this message across all routes (in millisatoshi).
	TotalFeesMsat int64 `json:"total_fees_msat"`
	// The routes used to send this message.
	Routes []Route `json:"routes"`
	// Preimage hash.
	PreimageHash []byte
	// The payment preimage.
	Preimage lnchat.PreImage `json:"preimage"`
	// The payment request (invoice) to be paid. If empty, corresponds to a spontaneous payment.
	PayReq string `json:"pay_req"`
	// Arrival success probability.
	SuccessProb float64
}
