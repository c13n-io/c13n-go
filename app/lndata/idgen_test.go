package lndata

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSequencePool(t *testing.T) {
	fastForward := func(sp *sequencePool, afterId uint64) {
		sp.lastId = afterId
	}

	generateSeq := func(start uint64, count int) []uint64 {
		res := make([]uint64, count)
		for i := 0; i < count; i++ {
			res[i] = start + uint64(i)
		}
		return res
	}

	checkIdCount := 10

	cases := map[string]func(*testing.T){
		"start at 1": func(t *testing.T) {
			pool := NewSequencePool(false)

			for _, expected := range generateSeq(1, checkIdCount) {
				id := pool.Get()
				assert.EqualValues(t, expected, id)
			}
		},
		"stop at max uint64": func(t *testing.T) {
			pool := NewSequencePool(false)

			var lastExpectedId uint64 = math.MaxUint64
			fastForward(pool, lastExpectedId-1)

			lastId := pool.Get()
			assert.EqualValues(t, lastExpectedId, lastId)

			// Assert id generation without wraparound
			// returns invalid ids (0) after max value.
			for _, expected := range make([]uint64, checkIdCount) {
				id := pool.Get()
				assert.EqualValues(t, expected, id)
			}
		},
		"reset after max with wrap around": func(t *testing.T) {
			pool := NewSequencePool(true)

			var startId uint64 = math.MaxUint64
			fastForward(pool, startId)

			// Assert id generation with wraparound
			// starts from the first valid id again after max.
			for _, expected := range generateSeq(1, checkIdCount) {
				id := pool.Get()
				assert.EqualValues(t, expected, id)
			}
		},
	}

	for name, tcase := range cases {
		t.Run(name, tcase)
	}
}
