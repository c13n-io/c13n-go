package lndata

const (
	defaultDataStructKey uint64 = 0x117C17A7 + 2*iota
	defaultDataSigKey

	// DataStructVersion holds the DataStruct
	// version implemented by this package.
	DataStructVersion uint32 = 1
	// DataSigVersion holds the DataSig
	// version implemented by this package.
	DataSigVersion
)

var (
	// DataStructKey is the default DataStruct key used by this package.
	DataStructKey = defaultDataStructKey
	// DataSigKey is the default DataSig key used by this package.
	DataSigKey = defaultDataSigKey
)
