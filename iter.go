package iradix

import (
	"bytes"
)

// Iterator is used to iterate over a set of nodes in pre-order
type Iterator struct {
	node  *Node
	stack [][]*Node
}

// SeekPrefixWatch seeks the iterator to a given prefix and returns the watch channel.
func (i *Iterator) SeekPrefixWatch(prefix []byte) (watch <-chan struct{}) {
	// Wipe the stack
	i.stack = nil
	n := i.node
	watch = n.mutateCh
	search := prefix
	for {
		// Check for key exhaustion
		if len(search) == 0 {
			i.node = n
			return
		}

		// Look for a child
		_, child := n.getEdge(search[0])
		if child == nil {
			i.node = nil
			return
		}
		n = child

		// Update watch
		watch = n.mutateCh

		// Consume the search prefix
		if bytes.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]

		} else if bytes.HasPrefix(n.prefix, search) {
			// search is a prefix of n.prefix
			i.node = n
			return
		} else {
			// prefix doesn't match
			i.node = nil
			return
		}
	}
}

// SeekPrefix is used to seek the iterator to a given prefix
func (i *Iterator) SeekPrefix(prefix []byte) {
	i.SeekPrefixWatch(prefix)
}

// recurseMin traverses to the minimum (lexicographically smallest) child node.
func (i *Iterator) recurseMin(n *Node) *Node {
	// If there's a leaf, return it.
	if n.leaf != nil {
		return n
	}
	nChildren := len(n.edges)
	if nChildren > 1 {
		// Add all but the first child to the stack.
		// The first child is the minimum; we recurse into it.
		i.stack = append(i.stack, n.edges[1:])
	}
	if nChildren > 0 {
		return i.recurseMin(n.edges[0])
	}
	// No edges means no minimum node
	return nil
}

// SeekLowerBound sets the iterator to the smallest key >= 'key'.
func (i *Iterator) SeekLowerBound(key []byte) {
	// Wipe the stack.
	i.stack = nil
	n := i.node
	i.node = nil
	search := key

	found := func(n *Node) {
		i.stack = append(i.stack, []*Node{n})
	}

	findMin := func(n *Node) {
		n = i.recurseMin(n)
		if n != nil {
			found(n)
		}
	}

	for {
		var prefixCmp int
		if len(n.prefix) < len(search) {
			prefixCmp = bytes.Compare(n.prefix, search[:len(n.prefix)])
		} else {
			prefixCmp = bytes.Compare(n.prefix, search)
		}

		if prefixCmp > 0 {
			// Current prefix > search: all keys in this subtree are >= search
			findMin(n)
			return
		}

		if prefixCmp < 0 {
			// Current prefix < search: no lower bound in this subtree
			return
		}

		// prefixCmp == 0
		if n.leaf != nil && bytes.Equal(n.leaf.key, key) {
			// Exact match
			found(n)
			return
		}

		search = search[len(n.prefix):]

		if len(search) == 0 {
			// Matched the prefix fully, all edges are >= key
			findMin(n)
			return
		}

		// Find the lower bound child
		idx, lbNode := n.getLowerBoundEdge(search[0])
		if lbNode == nil {
			// no child >= search[0]
			return
		}

		// Children after lbNode are strictly greater
		if idx+1 < len(n.edges) {
			i.stack = append(i.stack, n.edges[idx+1:])
		}

		n = lbNode
	}
}

// Next returns the next node in order (pre-order).
func (i *Iterator) Next() ([]byte, interface{}, bool) {
	// Initialize stack if needed
	if i.stack == nil && i.node != nil {
		i.stack = append(i.stack, []*Node{i.node})
	}

	for len(i.stack) > 0 {
		// Inspect the last element of the stack
		n := len(i.stack)
		last := i.stack[n-1]
		elem := last[0] // Take the first node from the top slice

		// Update the stack
		if len(last) > 1 {
			i.stack[n-1] = last[1:]
		} else {
			i.stack = i.stack[:n-1]
		}

		// Pre-order: node first, then edges.
		// If the node has edges, push them as a new slice to the stack.
		if len(elem.edges) > 0 {
			i.stack = append(i.stack, elem.edges)
		}

		// If this node has a leaf, return it.
		if elem.leaf != nil {
			return elem.leaf.key, elem.leaf.val, true
		}
	}

	return nil, nil, false
}
