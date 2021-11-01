package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/model"
)

func TestDefaultOptions(t *testing.T) {
	assert.EqualValues(t, model.MessageOptions{
		FeeLimitMsat: 3000,
		Anonymous:    false,
	}, DefaultOptions)
}

func TestOverrideOptions(t *testing.T) {
	type testCase struct {
		opts         model.MessageOptions
		overrideOpts []model.MessageOptions
		allowRelax   bool
		expected     model.MessageOptions
	}

	cases := []testCase{
		{
			opts: model.MessageOptions{
				FeeLimitMsat: 3000,
				Anonymous:    false,
			},
			overrideOpts: []model.MessageOptions{
				model.MessageOptions{
					FeeLimitMsat: 0,
					Anonymous:    true,
				},
			},
			allowRelax: true,
			expected: model.MessageOptions{
				FeeLimitMsat: 3000,
				Anonymous:    true,
			},
		},
		{
			opts: model.MessageOptions{
				FeeLimitMsat: 5000,
				Anonymous:    false,
			},
			overrideOpts: []model.MessageOptions{
				model.MessageOptions{
					FeeLimitMsat: 3000,
					Anonymous:    false,
				},
			},
			allowRelax: true,
			expected: model.MessageOptions{
				FeeLimitMsat: 3000,
				Anonymous:    false,
			},
		},
		{
			opts: model.MessageOptions{
				FeeLimitMsat: 3000,
				Anonymous:    false,
			},
			overrideOpts: []model.MessageOptions{
				model.MessageOptions{
					FeeLimitMsat: 50000,
					Anonymous:    true,
				},
			},
			allowRelax: false,
			expected: model.MessageOptions{
				FeeLimitMsat: 3000,
				Anonymous:    true,
			},
		},
		{
			opts: model.MessageOptions{
				FeeLimitMsat: 3000,
				Anonymous:    false,
			},
			overrideOpts: []model.MessageOptions{
				model.MessageOptions{
					FeeLimitMsat: 50000,
					Anonymous:    true,
				},
			},
			allowRelax: true,
			expected: model.MessageOptions{
				FeeLimitMsat: 50000,
				Anonymous:    true,
			},
		},
		{
			opts: model.MessageOptions{
				FeeLimitMsat: 3000,
				Anonymous:    false,
			},
			overrideOpts: []model.MessageOptions{
				model.MessageOptions{
					FeeLimitMsat: 500,
					Anonymous:    true,
				},
				model.MessageOptions{
					FeeLimitMsat: 2000,
					Anonymous:    true,
				},
			},
			allowRelax: true,
			expected: model.MessageOptions{
				FeeLimitMsat: 2000,
				Anonymous:    true,
			},
		},
		{
			opts: model.MessageOptions{
				FeeLimitMsat: 3000,
				Anonymous:    false,
			},
			overrideOpts: []model.MessageOptions{
				model.MessageOptions{
					FeeLimitMsat: 500,
					Anonymous:    true,
				},
				model.MessageOptions{
					FeeLimitMsat: 2000,
					Anonymous:    true,
				},
			},
			allowRelax: false,
			expected: model.MessageOptions{
				FeeLimitMsat: 2000,
				Anonymous:    true,
			},
		},
		{
			opts: model.MessageOptions{
				FeeLimitMsat: 3000,
				Anonymous:    false,
			},
			overrideOpts: nil,
			allowRelax:   false,
			expected: model.MessageOptions{
				FeeLimitMsat: 3000,
				Anonymous:    false,
			},
		},
	}

	for _, c := range cases {
		res := overrideOptions(c.opts, c.allowRelax, c.overrideOpts...)

		assert.EqualValues(t, c.expected, res)
	}
}
