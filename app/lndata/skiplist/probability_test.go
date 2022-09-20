package skiplist

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	floatEpsilon      float64 = .01
	samplingVariation float64 = 2 * floatEpsilon
)

func TestDefaultProbability(t *testing.T) {
	var expected float64 = 0.33333333
	assert.InEpsilon(t, expected, DefaultSkipProbability, floatEpsilon)
}

func TestGeneratePMF(t *testing.T) {
	type testcase struct {
		baseProb   float64
		eventCount uint
	}
	cases := map[string]testcase{
		"no events": testcase{
			baseProb:   .5,
			eventCount: 0,
		},
		"small event count": testcase{
			baseProb:   .5,
			eventCount: 3,
		},
		"large event count": testcase{
			baseProb:   .3,
			eventCount: 64,
		},
		"normal event count": testcase{
			baseProb:   .3,
			eventCount: 15,
		},
		"probability 0.9": testcase{
			baseProb:   .9,
			eventCount: 10,
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			result := generatePMF(tcase.baseProb, tcase.eventCount)

			// Should contain the requested number of elements.
			assert.Len(t, result, int(tcase.eventCount))
			// Result consists of probabilities (range [0, 1]).
			for _, prob := range result {
				assert.GreaterOrEqual(t, prob, .0)
				assert.LessOrEqual(t, prob, 1.)
			}
		})
	}
}

func TestDistributionSampling(t *testing.T) {
	type testcase struct {
		prob        float64
		buckets     uint
		repetitions uint64
	}

	cases := map[string]testcase{
		"prob: 1/3, buckets: 18, reps: 200k": testcase{
			prob:        1. / 3.,
			buckets:     18,
			repetitions: 200_000,
		},
		"prob: 1/2, buckets: 30, reps: 500k": testcase{
			prob:        1. / 2.,
			buckets:     30,
			repetitions: 500_000,
		},
		"prob: .9, buckets: 64, reps: 2M": testcase{
			prob:        .9,
			buckets:     64,
			repetitions: 2_000_000,
		},
	}

	if testing.Short() {
		t.Skip("skipping distribution sampling in short mode")
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			prob, reps := tcase.prob, tcase.repetitions
			sel := newSelector(prob, tcase.buckets, 0)

			results := make([]uint64, tcase.buckets)
			for i := uint64(0); i < reps; i++ {
				selected := sel.choose()
				results[selected]++
			}

			// Mean should be near the expected mean
			expectedMean := 1. / (1 - prob)
			mean := func(bs []uint64) float64 {
				var res, count float64
				for i := range bs {
					el, elCount := float64(i+1), float64(bs[i])
					count += elCount
					res += el * elCount
				}
				return res / count
			}(results)

			assert.InEpsilon(t, expectedMean, mean, floatEpsilon)

			// Cumulative counts should be near the expeceted values.
			expectedCumulCounts := make([]float64, len(results))
			actualCumulCounts := make([]float64, len(results))
			for i := range results {
				expectedCumulCounts[i] = (1. - math.Pow(prob,
					float64(i+1))) * float64(reps)

				actualCumulCounts[i] = float64(results[i])
				if i > 0 {
					actualCumulCounts[i] += actualCumulCounts[i-1]
				}
			}

			assert.InEpsilonSlice(t,
				expectedCumulCounts, actualCumulCounts, samplingVariation)
		})
	}
}
