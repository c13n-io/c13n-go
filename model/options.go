package model

import "github.com/c13n-io/c13n-backend/lnchat"

var defaultPaymentOpts = lnchat.PaymentOptions{
	FeeLimitMsat:   3000,
	FinalCltvDelta: 20,
	TimeoutSecs:    30,
}

// MessageOptions represents options for a message.
type MessageOptions struct {
	// The maximum fee allowed for sending a message (in millisatoshi).
	FeeLimitMsat int64 `json:"fee_limit_msat"`
	// Whether to include the sender address in the message.
	Anonymous bool `json:"anonymous"`
}

// WithFeeLimit sets the fee limit option.
func (o MessageOptions) WithFeeLimit(feeLimitMsat int64) MessageOptions {
	o.FeeLimitMsat = feeLimitMsat

	return o
}

// GetPaymentOptions returns the corresponding lnchat.PaymentOptions.
func (o MessageOptions) GetPaymentOptions() lnchat.PaymentOptions {
	payOpts := defaultPaymentOpts
	payOpts.FeeLimitMsat = o.FeeLimitMsat

	return payOpts
}
