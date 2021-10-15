package model

// TxFeeOptions specifies options used
// for fee calculation of an on-chain transaction.
type TxFeeOptions struct {
	SatPerVByte     uint64
	TargetConfBlock uint32
}
