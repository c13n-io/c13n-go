package skiplist

const (
	// minHeadLevelCnt denotes the minimum skiplist head length.
	minHeadLevelCnt uint = 1
	// maxLevelCnt denotes the maximum skiplist level count.
	maxLevelCnt = 64
)

// Node represents a skiplist node.
type Node struct {
	key uint64
	val interface{}

	next []*Node
	prev *Node
}

// Key returns a node's key.
// Panics if called on nil receiver.
func (n *Node) Key() uint64 {
	return n.key
}

// Value returns a node's value.
// Panics if called on nil receiver.
func (n *Node) Value() interface{} {
	return n.val
}

// Set sets a node's value.
// Panics if called on nil receiver.
func (n *Node) Set(v interface{}) {
	n.val = v
}

// Previous returns the immediately preceding node.
// Panics if called on nil receiver.
func (n *Node) Previous() *Node {
	return n.prev
}

// Next returns the immediately following node.
// Panics if called on nil receiver.
func (n *Node) Next() *Node {
	return n.next[0]
}

func (n *Node) is(k uint64) bool {
	return n != nil && n.key == k
}

func (n *Node) lessThan(k uint64) bool {
	return n != nil && n.key < k
}

// Skiplist represents a skiplist.
type Skiplist struct {
	*selector

	// Pointers to the head nodes for all skiplist levels,
	// as well as the last node on level 0.
	head  []*Node
	tail  *Node
	count uint64
}

// Length returns the number of nodes in the skiplist.
func (sl *Skiplist) Length() uint64 {
	return sl.count
}

// Head retrieves the first skiplist node, smallest in key order.
// If the skiplist is empty, returns nil.
func (sl *Skiplist) Head() *Node {
	return sl.head[0]
}

// Tail retrieves the last skiplist node, largest in key order.
// If the skiplist is empty, returns nil.
func (sl *Skiplist) Tail() *Node {
	return sl.tail
}

// Search finds the node with the specified key and returns it.
// Returns nil if the key is not present.
func (sl *Skiplist) Search(k uint64) *Node {
	path := sl.find(k)
	if node := sl.next(path[0]); node.is(k) {
		return node
	}

	return nil
}

// Insert adds a node containing value in the skiplist
// under the specified key and returns the new node,
// or the existing node if the key already exists.
// A boolean indicating whether the insertion failed is also returned.
func (sl *Skiplist) Insert(k uint64, v interface{}) (*Node, bool) {
	path := sl.find(k)
	if next := sl.next(path[0]); next.is(k) {
		return next, false
	}

	node := sl.newNode(k, v)
	for i := 0; i < len(node.next) && i < len(path); i++ {
		switch path[i] {
		case nil:
			// The new node is first for level i on the list.
			node.next[i], sl.head[i] = sl.head[i], node
		default:
			// The new node has a previous neighbour on level i.
			node.next[i], path[i].next[i] = path[i].next[i], node
		}
	}
	// Update list tail and next and new nodes' previous links.
	node.prev = path[0]
	switch node.next[0] {
	case nil:
		sl.tail = node
	default:
		node.next[0].prev = node
	}

	// Update the head list with any new levels.
	if extraLevels := len(node.next) - len(sl.head); extraLevels > 0 {
		sl.extend(extraLevels, node)
	}
	sl.count++

	return node, true
}

// Delete removes a node from the skiplist and returns it.
// If the key is not found, it returns nil.
func (sl *Skiplist) Delete(k uint64) *Node {
	path := sl.find(k)
	node := sl.next(path[0])
	if !node.is(k) {
		return nil
	}

	// Update list tail and next node's previous link.
	switch node.next[0] {
	case nil:
		sl.tail = node.prev
	default:
		node.next[0].prev = node.prev
	}
	for lvl := len(node.next) - 1; lvl >= 0; lvl-- {
		switch path[lvl] {
		case nil:
			sl.head[lvl] = node.next[lvl]
		default:
			path[lvl].next[lvl] = node.next[lvl]
		}
	}

	redundantLevels := 0
	for lvl := len(path) - 1; lvl >= 0 && sl.head[lvl] == nil; lvl-- {
		redundantLevels++
	}
	if redundantLevels > 0 {
		sl.shrink(redundantLevels)
	}
	sl.count--

	node.next, node.prev = nil, nil
	return node
}

// Iterate visits all nodes in ascending key order, starting with
// the first node with key equal or greater to the provided one.
// If the visitor function is nil, the iterator returns immediately.
// Iteration stops either after the last node is visited or
// once the visitor function returns false.
// Returns the number of visited nodes.
func (sl *Skiplist) Iterate(fromKey uint64, visit func(*Node) (cont bool)) uint64 {
	if visit == nil {
		return 0
	}

	start := sl.next(sl.find(fromKey)[0])
	var count uint64
	for cur := start; cur != nil; cur = cur.next[0] {
		count++
		if !visit(cur) {
			break
		}
	}

	return count
}

// Returns a list of previous nodes to the appropriate path
// where the key would be (or indeed is).
func (sl *Skiplist) find(k uint64) []*Node {
	path := make([]*Node, len(sl.head)+1)

	for lvl := len(sl.head) - 1; lvl >= 0; lvl-- {
		// Avoid overshooting the previous node in case the target key
		// is smaller than the highest-level node in the head list.
		start := path[lvl+1]
		if start == nil {
			start = sl.head[lvl]
		}
		for cur := start; cur.lessThan(k); cur = cur.next[lvl] {
			path[lvl] = cur
		}
	}

	return path[:len(sl.head)]
}

// Returns the node after the provided one on level 0.
// A nil argument signifies the start of the list,
// making it suitable for use with the result of find.
func (sl *Skiplist) next(n *Node) *Node {
	if n != nil {
		return n.next[0]
	}
	return sl.head[0]
}

func (sl *Skiplist) newNode(k uint64, v interface{}) *Node {
	return &Node{
		key:  k,
		val:  v,
		next: make([]*Node, sl.choose()+1),
	}
}

func (sl *Skiplist) extend(extraLevels int, n *Node) {
	for i := 0; i < extraLevels; i++ {
		sl.head = append(sl.head, n)
	}
}

func (sl *Skiplist) shrink(redundantLevels int) {
	remaining := len(sl.head) - redundantLevels
	if minLevelCount := int(minHeadLevelCnt); remaining < minLevelCount {
		for lvl := minLevelCount - 1; lvl >= remaining; lvl-- {
			sl.head[lvl] = nil
		}
		remaining = minLevelCount
	}
	sl.head = sl.head[:remaining]
}

// New creates a new skiplist.
func New(options ...SkiplistOption) *Skiplist {
	s := &Skiplist{
		selector: newSelector(DefaultSkipProbability, maxLevelCnt, 0),
		head:     make([]*Node, minHeadLevelCnt),
	}

	for _, opt := range options {
		opt(s)
	}

	return s
}

// SkiplistOption represents the option type of a skiplist.
type SkiplistOption func(*Skiplist)

// WithSkipProbability specifies the skip probability of the skiplist.
// A probability of 1 or higher will have no effect.
func WithSkipProbability(p float64) SkiplistOption {
	if p < 1.0 {
		return func(sl *Skiplist) {}
	}

	return func(sl *Skiplist) {
		sl.selector = newSelector(p, maxLevelCnt, 0)
	}
}
