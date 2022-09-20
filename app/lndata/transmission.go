package lndata

type Address = [33]byte

// Transmission is the type for a data transmission.
type Transmission struct {
	Data                []byte
	Source, Destination Address
	FragsetId           uint64

	fragments []fragment
}

// Verified returns whether the transmission
// included a signature and its sender has been verified.
func (t Transmission) Verified() bool {
	for _, frag := range t.fragments {
		if !frag.verified {
			return false
		}
	}

	return true
}

func (t Transmission) clone() Transmission {
	data := make([]byte, len(t.Data))
	copy(data, t.Data)

	return Transmission{
		Data:        data,
		Source:      t.Source,
		Destination: t.Destination,
		FragsetId:   t.FragsetId,
	}
}

func (t Transmission) populateFragment(start, length uint32) fragment {
	startOffset, endOffset := int(start), int(start+length)
	return fragment{
		start:     start,
		payload:   t.Data[startOffset:endOffset],
		totalSize: uint32(len(t.Data)),
		fragsetId: t.FragsetId,
	}
}
