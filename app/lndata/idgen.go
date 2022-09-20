package lndata

import (
	"math"
	"sync"
)

// Interface describing a pool that dispenses uint64 IDs, starting from 1.
// An ID of 0 should be considered invalid, and should
// only be returned in case the pool has ran out of IDs.
type IdPool interface {
	// Get retrieves an ID from the pool.
	Get() (id uint64)
	// Release returns an ID back to the pool.
	Release(id uint64)
}

// IdPool implementation that releases sequential IDs.
type sequencePool struct {
	m         sync.RWMutex
	lastId    uint64
	wrapOnMax bool
}

func (sp *sequencePool) Get() uint64 {
	sp.m.Lock()
	defer sp.m.Unlock()

	// Return invalid ID if case of overflow and no wrap around.
	if sp.lastId == math.MaxUint64 {
		if !sp.wrapOnMax {
			return 0
		}
		sp.lastId = 0
	}

	sp.lastId++
	return sp.lastId
}

func (sp *sequencePool) Release(_ uint64) {
	// No returns accepted on sequence pool.
}

// NewSequencePool returns an ID pool that dispenses sequential IDs.
func NewSequencePool(wrap bool) *sequencePool {
	return &sequencePool{
		wrapOnMax: wrap,
	}
}

// IDGenerator is the ID pool used for id generation.
var IDGenerator IdPool = NewSequencePool(false)
