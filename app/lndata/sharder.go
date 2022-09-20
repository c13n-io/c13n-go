package lndata

import (
	"fmt"
	"sync"

	"github.com/c13n-io/c13n-go/app/lndata/skiplist"
)

// ErrDestinationRequired is returned upon attempt to construct
// a sharder with valid signer but empty destination.
var ErrDestinationRequired = fmt.Errorf("data signing " +
	"requires valid destination address")

// Sharder allows dynamic fragmentation of the provided data.
type Sharder struct {
	signer Signer

	lock         sync.RWMutex
	transmission Transmission
	fragments    *fragmentStates
}

// NewSharder creates a sharder for the requested data.
// If a signer is provided, a valid destination address is required.
func NewSharder(data []byte, dest []byte, signer Signer) (*Sharder, error) {
	var source, destination Address
	if len(dest) != 0 && len(dest) != AddressSize {
		return nil, ErrInvalidAddress{dest}
	}
	copy(destination[:], dest)

	if signer != nil {
		if len(dest) != AddressSize {
			return nil, ErrDestinationRequired
		}
		source = signer.Address()
	}

	t := Transmission{
		Data:        data,
		Source:      source,
		Destination: destination,
		FragsetId:   IDGenerator.Get(),
	}

	return &Sharder{
		signer:       signer,
		transmission: t,
		fragments:    newStates(uint32(len(t.Data))),
	}, nil
}

// Get returns a transmission fragment of the requested size
// as a set of fields as well as a cancel function
// that releases the fragment back to the sharder in case of failure.
//
// An error is returned if the fragment cannot be marshalled
// or if signing fails with error.
func (s *Sharder) Get(size uint32) (fs Fields, cancelFunc func(), err error) {
	frag, ok := s.selectFragment(size)
	if !ok {
		return
	}
	cancelFunc = func() {
		s.cancelFragment(frag.start)
	}

	fs, err = marshalAndSign(frag, s.transmission.Destination, s.signer)
	if err != nil {
		cancelFunc()
	}

	return
}

// Pending returns the total number of pending payload bytes.
func (s *Sharder) Pending() (bytes uint32) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, bytes = s.fragments.countPending()
	return
}

// Result returns a copy of the transmission and a boolean indicating
// whether the transmission is totally covered by the generated fragments.
func (s *Sharder) Result() (t Transmission, complete bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	t = s.transmission.clone()

	frags, allFragmentsLeased := s.fragments.getComplete()
	t.fragments = make([]fragment, len(frags))
	for i, frag := range frags {
		t.fragments[i] = t.populateFragment(frag.start, frag.length)
		t.fragments[i].verified = (s.signer != nil)
	}

	return t, allFragmentsLeased
}

func (s *Sharder) selectFragment(length uint32) (fragment, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if sel := s.fragments.get(length); sel != nil {
		frag := s.transmission.populateFragment(sel.start, sel.length)
		return frag, true
	}

	return fragment{}, false
}

func (s *Sharder) cancelFragment(start uint32) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.fragments.put(start)
}

// fragmentStates keeps track of the fragment set states.
type fragmentStates struct {
	// pending contains all pending fragments sorted by length.
	// The key is constructed by key = (length << 32 | start_offset).
	pending *skiplist.Skiplist
	// all contains info on all (pending and released)
	// fragments, with key the fragment's start offset.
	all map[uint32]*fragmentInfo
}

type fragmentInfo struct {
	start, length uint32
	prev, next    *fragmentInfo
	leased        bool
}

func (fi *fragmentInfo) split(length uint32) (remainder *fragmentInfo) {
	if fi.length <= length {
		return
	}
	remainder = &fragmentInfo{
		start:  fi.start + length,
		length: fi.length - length,
		leased: fi.leased,
	}
	remainder.prev, fi.next = fi, remainder
	fi.length = length

	return remainder
}

func newStates(totalLength uint32) *fragmentStates {
	fs := &fragmentStates{
		pending: skiplist.New(),
		all:     make(map[uint32]*fragmentInfo),
	}

	fs.pending.Insert(fs.encodeKey(totalLength, 0), nil)
	fs.all[0] = &fragmentInfo{
		length: totalLength,
	}

	return fs
}

func (fs *fragmentStates) get(length uint32) *fragmentInfo {
	// If no pending fragments remain, return nothing.
	if fs.pending.Length() == 0 {
		return nil
	}

	var target *skiplist.Node
	maxPending := fs.pending.Tail()
	switch maxPendingLen, _ := fs.decodeKey(maxPending.Key()); true {
	case maxPendingLen < length:
		// If the requested length is larger than the largest
		// pending fragment, select the largest fragment.
		target = maxPending
	default:
		// Otherwise, the iterator will be called at least once,
		// since pending nodes with at least equal length exist.
		searchKey := fs.encodeKey(length, 0)
		fs.pending.Iterate(searchKey, func(node *skiplist.Node) bool {
			target = node
			return false
		})
	}
	target = fs.pending.Delete(target.Key())

	_, targetStart := fs.decodeKey(target.Key())
	result := fs.all[targetStart]
	if remainder := result.split(length); remainder != nil {
		fs.all[remainder.start] = remainder

		remKey := fs.encodeKey(remainder.length, remainder.start)
		fs.pending.Insert(remKey, nil)
	}

	result.leased = true
	return result
}

func (fs *fragmentStates) put(start uint32) {
	target, ok := fs.all[start]
	if !ok || !target.leased {
		return
	}

	delete(fs.all, start)
	target.leased = false

	if prev := target.prev; prev != nil && !prev.leased {
		prevKey := fs.encodeKey(prev.length, prev.start)
		fs.pending.Delete(prevKey)

		delete(fs.all, prev.start)

		target.start -= prev.length
		target.length += prev.length
	}
	if next := target.next; next != nil && !next.leased {
		nextKey := fs.encodeKey(next.length, next.start)
		fs.pending.Delete(nextKey)

		delete(fs.all, next.start)

		target.length += next.length
	}

	fs.all[target.start] = target
}

func (fs *fragmentStates) countPending() (fragCount, totalBytes uint32) {
	fragCount = uint32(fs.pending.Length())

	fs.pending.Iterate(0, func(node *skiplist.Node) bool {
		fragLen, _ := fs.decodeKey(node.Key())
		totalBytes += fragLen

		return true
	})

	return
}

func (fs *fragmentStates) getComplete() (frags []fragmentInfo, complete bool) {
	frags = make([]fragmentInfo, 0, len(fs.all))
	for _, frag := range fs.all {
		if !frag.leased {
			continue
		}

		frags = append(frags, fragmentInfo{
			start:  frag.start,
			length: frag.length,
		})
	}

	return frags, fs.pending.Length() == 0
}

func (fs *fragmentStates) decodeKey(key uint64) (length, start uint32) {
	var mask uint64 = ^uint64(0) >> 32
	return uint32((key >> 32) & mask), uint32(key & mask)
}

func (fs *fragmentStates) encodeKey(length, start uint32) (slKey uint64) {
	return uint64(length)<<32 | uint64(start)
}
