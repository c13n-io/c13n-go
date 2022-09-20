package lndata

type fragment struct {
	start   uint32
	payload []byte

	totalSize uint32
	fragsetId uint64

	verified bool
}
