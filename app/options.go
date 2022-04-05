package app

import (
	"github.com/c13n-io/c13n-go/model"
)

var (
	// DefaultOptions defines the default message options to be used when
	// no overrides have been set.
	DefaultOptions = model.MessageOptions{
		FeeLimitMsat: 3000,
		Anonymous:    false,
	}
	// DefaultPaymentOptions defines the default payment options to be used.
	DefaultPaymentOptions = DefaultOptions.GetPaymentOptions()
)

// overrideOptions consolidates the provided options
// with the first argument and returns the result.
// If relaxation of the fee limit is not allowed,
// the fee limit is capped by the initial value of opts.
// A fee limit of 0 is ignored and does not override a previous value.
func overrideOptions(opts model.MessageOptions, allowRelax bool,
	overrides ...model.MessageOptions) model.MessageOptions {

	res := opts
	for _, o := range overrides {
		res.Anonymous = o.Anonymous

		relaxFee := o.FeeLimitMsat > opts.FeeLimitMsat
		switch {
		// Ignore a fee limit of 0 in overrides
		case o.FeeLimitMsat == 0:
			continue
		// Explicitly allow relaxation wrt the first argument
		case allowRelax && relaxFee:
			res.FeeLimitMsat = o.FeeLimitMsat
		// Allow restriction
		case !relaxFee:
			res.FeeLimitMsat = o.FeeLimitMsat
		}
	}

	return res
}
