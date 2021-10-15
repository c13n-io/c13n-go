package model

// PageOptions represents pagination options.
type PageOptions struct {
	// LastID represents the id of the element to return.
	// 0 represents the first element (or last if reverse is specified).
	LastID uint64
	// PageSize represents the number
	// of requested elements (inclusive of the first element).
	// 0 represents no limit.
	PageSize uint64
	// Reverse represents that the requested range
	// retrieve elements from LastID going backwards.
	Reverse bool
}
