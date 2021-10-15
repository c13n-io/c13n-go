package lnchat

import "github.com/lightningnetwork/lnd/lntypes"

// LightningNode represents a Lightning network node.
type LightningNode struct {
	// Alias is the Lightning network alias of the node.
	Alias string
	// Address is the Lightning address of the node.
	Address string
}

// Chain represents a blockchain and network of a Lightning node.
type Chain struct {
	// The blockchain of a node.
	Chain string
	// The network a node is operating on.
	Network string
}

// SelfInfo contains information about the underlying Lightning node.
type SelfInfo struct {
	// Node holds the general node information.
	Node LightningNode
	// Chains are the chains of the current node.
	Chains []Chain
}

// BalanceAllocation represents the distribution of
// balance in local and remote endpoints.
type BalanceAllocation struct {
	// The part of the balance available on the local end (in millisatoshi).
	LocalMsat uint64
	// The part of the balance available on the remote end (in millisatoshi).
	RemoteMsat uint64
}

// SelfBalance contains information about the underlying Lightning node balance.
type SelfBalance struct {
	// The confirmed balance of the node's wallet (in satoshi).
	WalletConfirmedBalanceSat int64
	// The unconfirmed balance of the node's wallet (in satoshi).
	WalletUnconfirmedBalanceSat int64

	// The balance available across all open channels.
	ChannelBalance BalanceAllocation
	// The balance in pending open channels.
	PendingOpenBalance BalanceAllocation
	// The unsettled balance across all open channels.
	UnsettledBalance BalanceAllocation
}

// TxFeeOptions represents ways to control the fee
// of on-chain transactions.
type TxFeeOptions struct {
	// A manual fee rate of sats per virtual byte of the funding transaction.
	SatPerVByte uint64
	// A number of blocks from the current that the transaction
	// should confirm by, which is used for fee estimation.
	TargetConfBlock uint32
}

// ChannelPoint represents a channel, as identified by its funding transaction.
type ChannelPoint struct {
	// The funding transaction ID of the channel opening
	// transaction (hex-encoded and byte-reversed).
	FundingTxid string
	// The output index of the funding transaction.
	OutputIndex uint32
}

// LightningChannel represents a Lightning network channel between two nodes.
type LightningChannel struct {
	// The ID of the Lightning channel.
	ChannelID uint64
	// The Lightning address of the first endpoint (node) of the channel.
	Node1Address string
	// The Lightning address of the second endpoint of the channel.
	Node2Address string
	// Capacity is the channel capacity (in millisatoshi).
	CapacityMsat int64
}

// PaymentOptions contains the payment details for sending a message.
type PaymentOptions struct {
	// FeeLimitMsat is the maximum amount of fees (in millisatoshi)
	// the sender is willing to give in order to send a message.
	FeeLimitMsat int64
	// FinalCltvDelta is the difference in blocks from the current height
	// that should be used for the timelock of the final hop.
	FinalCltvDelta int32
	// TimeoutSecs is the upper limit (in seconds) afforded for
	// attempting to send a message.
	TimeoutSecs int32
}

// PreImageHash is the preimage hash of a payment.
type PreImageHash = lntypes.Hash

// PreImage is the type of a payment preimage.
type PreImage = lntypes.Preimage
