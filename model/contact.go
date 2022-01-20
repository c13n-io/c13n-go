package model

// Contact represents a contact for the application.
type Contact struct {
	ID          uint64 `badgerhold:"key"`
	DisplayName string
	Node
}
