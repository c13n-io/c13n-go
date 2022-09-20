package lndata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	cases := map[string]func(*testing.T){
		"DataStructVersion": func(t *testing.T) {
			assert.EqualValues(t, 1, DataStructVersion)
		},
		"DataSigVersion": func(t *testing.T) {
			assert.EqualValues(t, 1, DataSigVersion)
		},
		"default DataStructKey": func(t *testing.T) {
			var expected uint64 = 0x117C17A7
			assert.EqualValues(t, expected, defaultDataStructKey)

			assert.EqualValues(t, defaultDataStructKey,
				DataStructKey)
		},
		"default DataSigKey": func(t *testing.T) {
			var expected uint64 = 0x117C17A9
			assert.EqualValues(t, expected, defaultDataSigKey)

			assert.EqualValues(t, defaultDataSigKey, DataSigKey)
		},
	}

	for name, tcase := range cases {
		t.Run(name, tcase)
	}
}
