package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/lnchat"
)

func TestGetPaymentOptions(t *testing.T) {
	cases := []struct {
		name            string
		opts            MessageOptions
		expectedPayOpts lnchat.PaymentOptions
	}{
		{
			name: "fee limit",
			opts: MessageOptions{
				FeeLimitMsat: 32109,
				Anonymous:    false,
			},
			expectedPayOpts: lnchat.PaymentOptions{
				FeeLimitMsat:   32109,
				FinalCltvDelta: defaultPaymentOpts.FinalCltvDelta,
				TimeoutSecs:    defaultPaymentOpts.TimeoutSecs,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			payOpts := c.opts.GetPaymentOptions()
			assert.EqualValues(t, c.expectedPayOpts, payOpts)
		})
	}
}

func TestOptionsWithFeeLimit(t *testing.T) {
	opts := MessageOptions{
		FeeLimitMsat: 3021,
		Anonymous:    false,
	}

	var newFeeLimit int64 = 5000
	res := opts.WithFeeLimit(newFeeLimit)

	expected := MessageOptions{
		FeeLimitMsat: newFeeLimit,
		Anonymous:    opts.Anonymous,
	}

	assert.Equal(t, expected, res)
}
