package lnchat

//go:generate mockery -dir=. -output=./mocks -outpkg=lnmock -name=LightManager

import (
	"context"
)

// LightManager is the API for the lnchat service.
type LightManager interface {
	GetSelfInfo(ctx context.Context) (SelfInfo, error)
	ListNodes(ctx context.Context) ([]LightningNode, error)
	GetSelfBalance(ctx context.Context) (*SelfBalance, error)

	ConnectNode(ctx context.Context, address string, hostport string) error
	OpenChannel(ctx context.Context, address string,
		private bool, amtMsat, pushAmtMsat uint64,
		minOpenConfirmations int32, txOpts TxFeeOptions) (*ChannelPoint, error)

	VerifySignatureExtractPubkey(ctx context.Context, message, signature []byte) (string, error)
	SignMessage(ctx context.Context, message []byte) ([]byte, error)

	SubscribeInvoiceUpdates(ctx context.Context, startIdx uint64,
		filter InvoiceUpdateFilter) (<-chan InvoiceUpdate, error)
	SubscribePaymentUpdates(ctx context.Context, startIdx uint64,
		filter PaymentUpdateFilter) (<-chan PaymentUpdate, error)
	SendPayment(ctx context.Context, recipient string, amt Amount, payReq string,
		payOpts PaymentOptions, payload map[uint64][]byte,
		filter PaymentUpdateFilter) (<-chan PaymentUpdate, error)

	DecodePayReq(ctx context.Context, payReq string) (*PayReq, error)
	CreateInvoice(ctx context.Context, memo string, amt Amount,
		expiry int64, privateHints bool) (*Invoice, error)
	LookupInvoice(ctx context.Context, payHash string) (*Invoice, error)

	GetRoute(ctx context.Context, recipient string, amt Amount,
		payOpts PaymentOptions, payload map[uint64][]byte) (
		route *Route, prob float64, err error)

	Close() error
}
