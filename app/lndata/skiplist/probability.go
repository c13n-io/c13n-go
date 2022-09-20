package skiplist

import (
	"math"
	"math/rand"
	"time"
)

// DefaultSkipProbability defines the default probability for
// a node of a level to also be included in the next higher level.
const DefaultSkipProbability float64 = 1. / 3.

type distribution []float64

func generatePMF(baseP float64, count uint) distribution {
	// Assuming (the same) skip probability p = 1-q for each event,
	// the probability of an element selecting event i follows a
	// geometric distribution with parameter q = 1-p, P[i] = (1-p)*p^(i-1),
	// which due to 0-indexing becomes P[i] = (1-p)*p^i
	// (with event 0 probability being 1-p = q).
	// This is assuming the utilized rng is uniformly distributed in [0, 1).
	probabilities := make([]float64, count)
	for i := range probabilities {
		probabilities[i] = math.Pow(baseP, float64(i+1))
	}

	return probabilities
}

// randomSource is the interface a random number generator must satisfy.
type randomSource interface {
	Float64() float64
}

// Creates a new (safe for concurrent use) random generator.
// If the provided seed is 0, the current time is used.
func newRandomSource(seed int64) randomSource {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return rand.New(rand.NewSource(seed))
}

type selector struct {
	pmf  distribution
	rand randomSource
}

// choose returns one of the available distribution events.
// The result is 0-indexed, and compatible with use as slice subscript.
func (s *selector) choose() uint {
	if s == nil || len(s.pmf) == 0 {
		return 0
	}

	rv := s.rand.Float64()
	for lvl, probability := range s.pmf {
		if rv > probability {
			return uint(lvl)
		}
	}

	// Bias towards 0 in case the generated
	// random value falls outside the pmf values.
	return 0
}

func newSelector(skipProb float64, evtCount uint, seed int64) *selector {
	return &selector{
		pmf:  generatePMF(skipProb, evtCount),
		rand: newRandomSource(seed),
	}
}
