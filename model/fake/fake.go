package fake

import (
	"fmt"
	"math/rand"
)

// GenerateAddress generates a fake Lightning address for testing.
func GenerateAddress() string {
	prefixes := []byte{0x02, 0x03}
	res := make([]byte, 33)

	res[0] = prefixes[rand.Intn(len(prefixes))]
	rand.Read(res[1:])

	return fmt.Sprintf("%x", res)
}

func randomUint64Range(low, high uint64) uint64 {
	if high <= low {
		return low
	}
	return uint64(rand.Int63n(int64(high-low)) + int64(low))
}
