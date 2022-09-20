package lndata

import (
	"container/heap"
	"sync"
	"time"
)

type Fields = map[uint64][]byte

// IncompleteClearTimeout denotes the duration a transmission
// will remain incomplete with no new fragments before deletion.
var IncompleteClearTimeout = 2 * time.Minute

// Aggregator allows reconstruction of transmissions from fragments.
type Aggregator struct {
	verifier Verifier
	self     Address

	lock       sync.Mutex
	incomplete map[transmissionID]*transmissionState
}

// NewAggregator creates an fragment aggregator.
// If a verifier is not provided, all returned transmissions are unverified.
func NewAggregator(verifier Verifier) (*Aggregator, error) {
	var selfAddress Address
	if verifier != nil {
		selfAddress = verifier.Address()
	}

	return &Aggregator{
		verifier:   verifier,
		self:       selfAddress,
		incomplete: make(map[transmissionID]*transmissionState),
	}, nil
}

// Update receives a set of fields and returns a channel over which
// the transmission will be sent once reconstructed.
// Fragments of the same transmission return the same channel.
//
// If the transmission is not reconstructed before timeout,
// it will be discarded and the channel closed.
//
// An error is returned if the fields cannot be unmarshalled
// or if verification fails with error.
func (ag *Aggregator) Update(fields Fields) (<-chan Transmission, error) {
	frag, sender, err := unmarshalAndVerify(fields, ag.verifier)
	if err != nil {
		return nil, err
	}

	tid := transmissionID{
		source:    sender,
		fragsetId: frag.fragsetId,
		length:    frag.totalSize,
	}

	return ag.updateStateWithFragment(tid, &frag), nil
}

func (ag *Aggregator) updateStateWithFragment(tid transmissionID,
	frag *fragment) <-chan Transmission {

	ag.lock.Lock()
	defer ag.lock.Unlock()

	state, exists := ag.incomplete[tid]
	if !exists {
		state = newState(tid, ag.self, func() {
			ag.removeState(tid)
		})
		ag.incomplete[tid] = state
	}

	state.fragmentCh <- frag

	return state.notificationCh
}

func (ag *Aggregator) removeState(tid transmissionID) {
	ag.lock.Lock()
	defer ag.lock.Unlock()

	delete(ag.incomplete, tid)
}

type transmissionID struct {
	source    Address
	fragsetId uint64
	length    uint32
}

// transmissionState keeps track of an incomplete transmission
// by incrementally reconstructing it starting from the beginning.
//
// If the transmission remains in incomplete state for more than
// IncompleteClearTimeout with no new fragments, it is discarded
// and the notification channel is closed.
//
// If the transmission is completed, the result is sent
// on the notification channel, which is then closed.
type transmissionState struct {
	transmission *Transmission

	// firstPendingOffset indicates the first byte offset in
	// the transmission data that has not been reconstructed.
	firstPendingOffset uint32
	// pending holds the received fragments that have not
	// been utilized in transmission reconstruction (yet).
	pending *fragmentHeap

	fragmentCh     chan *fragment
	notificationCh chan Transmission
}

func newState(tid transmissionID, dest Address,
	afterFunc func()) *transmissionState {

	state := &transmissionState{
		transmission: &Transmission{
			Data:        make([]byte, tid.length),
			Source:      tid.source,
			Destination: dest,
			FragsetId:   tid.fragsetId,
		},
		pending:        new(fragmentHeap),
		fragmentCh:     make(chan *fragment, 1),
		notificationCh: make(chan Transmission, 1),
	}

	go state.wait(afterFunc)

	return state
}

func (s *transmissionState) wait(cleanupFunc func()) {
	timer := time.NewTimer(IncompleteClearTimeout)

gatherLoop:
	for !s.isComplete() {
		select {
		case <-timer.C:
			break gatherLoop
		case frag := <-s.fragmentCh:
			if !timer.Stop() {
				<-timer.C
			}

			s.add(frag)

			if !s.isComplete() {
				timer.Reset(IncompleteClearTimeout)
			}
		}
	}

	if s.isComplete() {
		s.notificationCh <- *s.transmission
	}

	close(s.notificationCh)
	cleanupFunc()
}

// add adds a fragment to the fragment heap and attempts
// to pop any pending fragments adjacent (or overlapping)
// to the current reconstruction index of the parent transmission.
func (s *transmissionState) add(frag *fragment) {
	heap.Push(s.pending, frag)

	pending, t := s.pending, s.transmission
	for pending.Len() > 0 && pending.Peek().start <= s.firstPendingOffset {
		nextFrag := heap.Pop(pending).(*fragment)

		copy(t.Data[nextFrag.start:], nextFrag.payload)
		t.fragments = append(t.fragments, *nextFrag)

		nextOffset := nextFrag.start + uint32(len(nextFrag.payload))
		if nextOffset > s.firstPendingOffset {
			s.firstPendingOffset = nextOffset
		}
	}
}

func (s *transmissionState) isComplete() bool {
	return s.firstPendingOffset == uint32(len(s.transmission.Data))
}

// fragmentHeap represents a min-heap for fragments by start offset.
type fragmentHeap []*fragment

func (fh fragmentHeap) Len() int {
	return len(fh)
}

func (fh fragmentHeap) Less(i, j int) bool {
	return fh[i].start < fh[j].start
}

func (fh fragmentHeap) Swap(i, j int) {
	fh[i], fh[j] = fh[j], fh[i]
}

func (fh *fragmentHeap) Push(x interface{}) {
	*fh = append(*fh, x.(*fragment))
}

func (fh *fragmentHeap) Pop() interface{} {
	old := *fh
	oldLen := len(old)
	frag := old[oldLen-1]
	old[oldLen-1] = nil
	*fh = old[0 : oldLen-1]
	return frag
}

func (fh fragmentHeap) Peek() *fragment {
	return fh[0]
}
