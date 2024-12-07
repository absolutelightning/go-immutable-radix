package iradix

import (
	"bytes"
)

// ReverseIterator is used to iterate over a set of nodes in reverse in-order.
type ReverseIterator struct {
	i *Iterator

	// expandedParents keeps track of nodes whose children have been pushed.
	expandedParents map[*Node]struct{}
}

// NewReverseIterator returns a new ReverseIterator at a node
func NewReverseIterator(n *Node) *ReverseIterator {
	return &ReverseIterator{
		i: &Iterator{node: n},
	}
}

// SeekPrefixWatch seeks the iterator to a given prefix and returns the watch channel.
func (ri *ReverseIterator) SeekPrefixWatch(prefix []byte) (watch <-chan struct{}) {
	return ri.i.SeekPrefixWatch(prefix)
}

// SeekPrefix seeks the iterator to a given prefix.
func (ri *ReverseIterator) SeekPrefix(prefix []byte) {
	ri.i.SeekPrefixWatch(prefix)
}

// SeekReverseLowerBound sets the iterator to the largest key <= 'key'.
func (ri *ReverseIterator) SeekReverseLowerBound(key []byte) {
	ri.i.stack = nil
	n := ri.i.node
	ri.i.node = nil
	search := key

	if ri.expandedParents == nil {
		ri.expandedParents = make(map[*Node]struct{})
	}

	// found adds a single node as a slice and marks it as expanded
	found := func(n *Node) {
		ri.i.stack = append(ri.i.stack, []*Node{n})
		ri.expandedParents[n] = struct{}{}
	}

	for {
		var prefixCmp int
		if len(n.prefix) < len(search) {
			prefixCmp = bytes.Compare(n.prefix, search[:len(n.prefix)])
		} else {
			prefixCmp = bytes.Compare(n.prefix, search)
		}

		if prefixCmp < 0 {
			// n.prefix < search => reverse lower bound is under this subtree.
			// Push this node; the reverse iteration (Previous) will descend into it.
			ri.i.stack = append(ri.i.stack, []*Node{n})
			return
		}

		if prefixCmp > 0 {
			// n.prefix > search => no reverse lower bound here.
			return
		}

		// prefixCmp == 0
		if n.isLeaf() {
			if bytes.Equal(n.leaf.key, key) {
				// Exact match
				found(n)
				return
			}

			// Leaf < key (since not equal). If no children, this leaf is the lower bound.
			if len(n.children) == 0 {
				found(n)
				return
			}

			// Leaf with children. Push node first, mark expanded.
			ri.i.stack = append(ri.i.stack, []*Node{n})
			ri.expandedParents[n] = struct{}{}
		}

		// Consume matched prefix
		search = search[len(n.prefix):]

		if len(search) == 0 {
			// Exhausted search key, not at a leaf, all children > search => no lower bound here.
			return
		}

		idx, lbNode := n.getLowerBoundEdge(search[0])
		if idx == -1 {
			idx = len(n.children)
		}

		// Children before idx are strictly lower than search
		if idx > 0 {
			ri.i.stack = append(ri.i.stack, n.children[:idx])
		}

		if lbNode == nil {
			// No lower bound child
			return
		}

		n = lbNode
	}
}

// Previous returns the previous node in reverse order.
func (ri *ReverseIterator) Previous() ([]byte, interface{}, bool) {
	if ri.i.stack == nil && ri.i.node != nil {
		// Initialize stack with the root node if needed
		ri.i.stack = append(ri.i.stack, []*Node{ri.i.node})
	}

	if ri.expandedParents == nil {
		ri.expandedParents = make(map[*Node]struct{})
	}

	for len(ri.i.stack) > 0 {
		// Get the top slice of nodes
		n := len(ri.i.stack)
		top := ri.i.stack[n-1]
		m := len(top)
		elem := top[m-1] // The top node on the stack

		// Pop this node from the top slice
		if m > 1 {
			ri.i.stack[n-1] = top[:m-1]
		} else {
			ri.i.stack = ri.i.stack[:n-1]
		}

		_, alreadyExpanded := ri.expandedParents[elem]

		// If this node has children and isn't expanded, expand now.
		if len(elem.children) > 0 && !alreadyExpanded {
			ri.expandedParents[elem] = struct{}{}

			// After processing children, we want to revisit this node (elem).
			// Push it back as a single-node slice, so its leaf is considered after its children.
			ri.i.stack = append(ri.i.stack, []*Node{elem})

			// For reverse order, we want to visit the largest child first.
			// By default, children are in ascending order. We rely on popping last element first,
			// so we can append children as is. The last child in children is largest.
			ri.i.stack = append(ri.i.stack, elem.children)

			continue
		}

		// If already expanded or no children, we've fully popped elem now.
		if alreadyExpanded {
			delete(ri.expandedParents, elem)
		}

		// If elem has a leaf, return it
		if elem.leaf != nil {
			return elem.leaf.key, elem.leaf.val, true
		}
		// If no leaf, continue
	}

	return nil, nil, false
}
