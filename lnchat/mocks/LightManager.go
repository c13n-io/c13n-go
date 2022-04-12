// Code generated by mockery v1.0.0. DO NOT EDIT.

package lnmock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	lnchat "github.com/c13n-io/c13n-go/lnchat"
)

// LightManager is an autogenerated mock type for the LightManager type
type LightManager struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *LightManager) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ConnectNode provides a mock function with given fields: ctx, address, hostport
func (_m *LightManager) ConnectNode(ctx context.Context, address string, hostport string) error {
	ret := _m.Called(ctx, address, hostport)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, address, hostport)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateInvoice provides a mock function with given fields: ctx, memo, amt, expiry, privateHints
func (_m *LightManager) CreateInvoice(ctx context.Context, memo string, amt lnchat.Amount, expiry int64, privateHints bool) (*lnchat.Invoice, error) {
	ret := _m.Called(ctx, memo, amt, expiry, privateHints)

	var r0 *lnchat.Invoice
	if rf, ok := ret.Get(0).(func(context.Context, string, lnchat.Amount, int64, bool) *lnchat.Invoice); ok {
		r0 = rf(ctx, memo, amt, expiry, privateHints)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*lnchat.Invoice)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, lnchat.Amount, int64, bool) error); ok {
		r1 = rf(ctx, memo, amt, expiry, privateHints)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DecodePayReq provides a mock function with given fields: ctx, payReq
func (_m *LightManager) DecodePayReq(ctx context.Context, payReq string) (*lnchat.PayReq, error) {
	ret := _m.Called(ctx, payReq)

	var r0 *lnchat.PayReq
	if rf, ok := ret.Get(0).(func(context.Context, string) *lnchat.PayReq); ok {
		r0 = rf(ctx, payReq)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*lnchat.PayReq)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, payReq)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRoute provides a mock function with given fields: ctx, recipient, amt, payOpts, payload
func (_m *LightManager) GetRoute(ctx context.Context, recipient string, amt lnchat.Amount, payOpts lnchat.PaymentOptions, payload map[uint64][]byte) (*lnchat.Route, float64, error) {
	ret := _m.Called(ctx, recipient, amt, payOpts, payload)

	var r0 *lnchat.Route
	if rf, ok := ret.Get(0).(func(context.Context, string, lnchat.Amount, lnchat.PaymentOptions, map[uint64][]byte) *lnchat.Route); ok {
		r0 = rf(ctx, recipient, amt, payOpts, payload)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*lnchat.Route)
		}
	}

	var r1 float64
	if rf, ok := ret.Get(1).(func(context.Context, string, lnchat.Amount, lnchat.PaymentOptions, map[uint64][]byte) float64); ok {
		r1 = rf(ctx, recipient, amt, payOpts, payload)
	} else {
		r1 = ret.Get(1).(float64)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string, lnchat.Amount, lnchat.PaymentOptions, map[uint64][]byte) error); ok {
		r2 = rf(ctx, recipient, amt, payOpts, payload)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetSelfBalance provides a mock function with given fields: ctx
func (_m *LightManager) GetSelfBalance(ctx context.Context) (*lnchat.SelfBalance, error) {
	ret := _m.Called(ctx)

	var r0 *lnchat.SelfBalance
	if rf, ok := ret.Get(0).(func(context.Context) *lnchat.SelfBalance); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*lnchat.SelfBalance)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSelfInfo provides a mock function with given fields: ctx
func (_m *LightManager) GetSelfInfo(ctx context.Context) (lnchat.SelfInfo, error) {
	ret := _m.Called(ctx)

	var r0 lnchat.SelfInfo
	if rf, ok := ret.Get(0).(func(context.Context) lnchat.SelfInfo); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(lnchat.SelfInfo)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListNodes provides a mock function with given fields: ctx
func (_m *LightManager) ListNodes(ctx context.Context) ([]lnchat.LightningNode, error) {
	ret := _m.Called(ctx)

	var r0 []lnchat.LightningNode
	if rf, ok := ret.Get(0).(func(context.Context) []lnchat.LightningNode); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]lnchat.LightningNode)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LookupInvoice provides a mock function with given fields: ctx, payHash
func (_m *LightManager) LookupInvoice(ctx context.Context, payHash string) (*lnchat.Invoice, error) {
	ret := _m.Called(ctx, payHash)

	var r0 *lnchat.Invoice
	if rf, ok := ret.Get(0).(func(context.Context, string) *lnchat.Invoice); ok {
		r0 = rf(ctx, payHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*lnchat.Invoice)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, payHash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OpenChannel provides a mock function with given fields: ctx, address, private, amtMsat, pushAmtMsat, minOpenConfirmations, txOpts
func (_m *LightManager) OpenChannel(ctx context.Context, address string, private bool, amtMsat uint64, pushAmtMsat uint64, minOpenConfirmations int32, txOpts lnchat.TxFeeOptions) (*lnchat.ChannelPoint, error) {
	ret := _m.Called(ctx, address, private, amtMsat, pushAmtMsat, minOpenConfirmations, txOpts)

	var r0 *lnchat.ChannelPoint
	if rf, ok := ret.Get(0).(func(context.Context, string, bool, uint64, uint64, int32, lnchat.TxFeeOptions) *lnchat.ChannelPoint); ok {
		r0 = rf(ctx, address, private, amtMsat, pushAmtMsat, minOpenConfirmations, txOpts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*lnchat.ChannelPoint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, bool, uint64, uint64, int32, lnchat.TxFeeOptions) error); ok {
		r1 = rf(ctx, address, private, amtMsat, pushAmtMsat, minOpenConfirmations, txOpts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendPayment provides a mock function with given fields: ctx, recipient, amt, payReq, payOpts, payload, filter
func (_m *LightManager) SendPayment(ctx context.Context, recipient string, amt lnchat.Amount, payReq string, payOpts lnchat.PaymentOptions, payload map[uint64][]byte, filter func(*lnchat.Payment) bool) (<-chan lnchat.PaymentUpdate, error) {
	ret := _m.Called(ctx, recipient, amt, payReq, payOpts, payload, filter)

	var r0 <-chan lnchat.PaymentUpdate
	if rf, ok := ret.Get(0).(func(context.Context, string, lnchat.Amount, string, lnchat.PaymentOptions, map[uint64][]byte, func(*lnchat.Payment) bool) <-chan lnchat.PaymentUpdate); ok {
		r0 = rf(ctx, recipient, amt, payReq, payOpts, payload, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan lnchat.PaymentUpdate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, lnchat.Amount, string, lnchat.PaymentOptions, map[uint64][]byte, func(*lnchat.Payment) bool) error); ok {
		r1 = rf(ctx, recipient, amt, payReq, payOpts, payload, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignMessage provides a mock function with given fields: ctx, message
func (_m *LightManager) SignMessage(ctx context.Context, message []byte) ([]byte, error) {
	ret := _m.Called(ctx, message)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, []byte) []byte); ok {
		r0 = rf(ctx, message)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []byte) error); ok {
		r1 = rf(ctx, message)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SubscribeInvoiceUpdates provides a mock function with given fields: ctx, startIdx, filter
func (_m *LightManager) SubscribeInvoiceUpdates(ctx context.Context, startIdx uint64, filter func(*lnchat.Invoice) bool) (<-chan lnchat.InvoiceUpdate, error) {
	ret := _m.Called(ctx, startIdx, filter)

	var r0 <-chan lnchat.InvoiceUpdate
	if rf, ok := ret.Get(0).(func(context.Context, uint64, func(*lnchat.Invoice) bool) <-chan lnchat.InvoiceUpdate); ok {
		r0 = rf(ctx, startIdx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan lnchat.InvoiceUpdate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint64, func(*lnchat.Invoice) bool) error); ok {
		r1 = rf(ctx, startIdx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SubscribePaymentUpdates provides a mock function with given fields: ctx, startIdx, filter
func (_m *LightManager) SubscribePaymentUpdates(ctx context.Context, startIdx uint64, filter func(*lnchat.Payment) bool) (<-chan lnchat.PaymentUpdate, error) {
	ret := _m.Called(ctx, startIdx, filter)

	var r0 <-chan lnchat.PaymentUpdate
	if rf, ok := ret.Get(0).(func(context.Context, uint64, func(*lnchat.Payment) bool) <-chan lnchat.PaymentUpdate); ok {
		r0 = rf(ctx, startIdx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan lnchat.PaymentUpdate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint64, func(*lnchat.Payment) bool) error); ok {
		r1 = rf(ctx, startIdx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// VerifySignatureExtractPubkey provides a mock function with given fields: ctx, message, signature
func (_m *LightManager) VerifySignatureExtractPubkey(ctx context.Context, message []byte, signature []byte) (string, error) {
	ret := _m.Called(ctx, message, signature)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, []byte, []byte) string); ok {
		r0 = rf(ctx, message, signature)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []byte, []byte) error); ok {
		r1 = rf(ctx, message, signature)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
