// Package bintree is a binary search tree (BST) implementation.
package bintree

// Interface must be implemented by elements (values) inserted in the
// tree (i.e. tree elements must be comparable to each other).
type Interface interface {
	// The Cmp method returns zero if the receiver is equal to the
	// argument, a negative if the receiver is less than
	// (precedes) the argument, and a positive otherwise.
	Cmp(Interface) int
}

// Node is a tree node. A *Node (pointer to the root node) represents
// a tree. A nil *Node is an empty tree.
type Node struct {
	// The node's element (value)
	V    Interface
	l, r *Node
}

// Init initializes a pre-allocated tree node with the given element
// (value)
func (n *Node) Init(v Interface) *Node {
	n.V, n.l, n.r = v, nil, nil
	return n
}

// New allocates a tree node, initializes it with the given element
// (value) and returns a pointer to it
func New(v Interface) *Node {
	return &Node{v, nil, nil}
}

// Insert adds "node" to the tree. An empty tree is a nil *Node
// pointer. If "unique" is true, the insetion will fail if there is
// already a node in the tree with the same element (value). Returns a
// pointer to the new tree root (i.e. to the new tree) and true, if
// the insertion was succesful, or false if the insertion failed. An
// Insert call with "unique" == false cannot fail.
func (tree *Node) Insert(node *Node, unique bool) (*Node, bool) {
	var ok bool
	var cn *Node

	// Special case, empty tree
	if tree == nil {
		return node, true
	}
	// Non-empty tree, search and place new node
	cn = tree
	for {
		if cmp := node.V.Cmp(cn.V); cmp < 0 {
			if cn.l == nil {
				cn.l = node
				ok = true
				break
			} else {
				cn = cn.l
			}
		} else if cmp > 0 || (cmp == 0 && !unique) {
			if cn.r == nil {
				cn.r = node
				ok = true
				break
			} else {
				cn = cn.r
			}
		} else {
			// cmp == 0 && unique 
			ok = false
			break
		}
	}
	return tree, ok
}

// minNode returns a pointer to the node with the the minimum element
// (value) in the tree (min node), and a pointer to the parent of that
// node. Returns (nil, nil) for an empty tree. Returns (tree, nil) if
// the min node is the tree root.
func (tree *Node) minNode() (*Node, *Node) {
	var parent *Node = nil
	if tree == nil {
		return tree, parent
	}
	for tree.l != nil {
		parent = tree
		tree = tree.l
	}
	return tree, parent
}

// maxNode returns a pointer to the node with the the maximum element
// (value) in the tree (max node), and a pointer to the parent of that
// node. Returns (nil, nil) for an empty tree. Returns (tree, nil) if
// the max node is the tree root.
func (tree *Node) maxNode() (*Node, *Node) {
	var parent *Node = nil
	if tree == nil {
		return tree, parent
	}
	for tree.r != nil {
		parent = tree
		tree = tree.r
	}
	return tree, parent
}

// findNode locates the first node in the tree with element (value)
// that satisfies key.Cmp(v) == 0. Returns a pointer to this node and
// a pointer to its father. Returns nil as a node-pointer if no node
// was located for the given key. Returns (tree, nil) if the located
// node is the tree root.
func (tree *Node) findNode(key Interface) (node, parent *Node) {
	node, parent = tree, nil
	for node != nil {
		if k := key.Cmp(node.V); k == 0 {
			break
		} else if k < 0 {
			parent = node
			node = node.l
		} else {
			parent = node
			node = node.r
		}
	}
	return node, parent
}

// rmNode removes the tree node pointed to by "node". "parent" is a
// pointer to the node's parent (or nil if the tree root is to be
// deleted). Returns a pointer to the new tree root; the root changes
// only if the root node is deleted.
func (tree *Node) rmNode(node, parent *Node) *Node {
	if node.l == nil {
		// Node to be removed has no left subtree. Replace
		// node with right subtree
		if parent == nil {
			// Removing root node
			tree = node.r
		} else if parent.l == node {
			parent.l = node.r
		} else {
			parent.r = node.r
		}
	} else if node.r == nil {
		// Node to be removed has no right subtree. Replace
		// node with left subtree
		if parent == nil {
			// Removing root node
			tree = node.l
		} else if parent.l == node {
			parent.l = node.l
		} else {
			parent.r = node.l
		}
	} else {
		// Node to be removed has both subtrees. Find the
		// adjacent (next) node in the sort-order sense
		// (nnode). It will be a node with at most one
		// subtree. Swap values between node and nnode and
		// remove nnode instead
		nnode, pnnode := node.r.minNode()
		if pnnode == nil {
			pnnode = node
		}
		node.V = nnode.V
		tree = tree.rmNode(nnode, pnnode)
	}
	return tree
}

// Remove locates and removes a tree node. The first node with element
// (value) that satisfies key.Cmp(v) == 0 is located and removed from
// the tree. Returns a pointer to the the new tree root, a pointer to
// the removed node, and a boolean. The boolean return value is true
// if a node was removed, and false if no node was found / removed for
// the given key.
func (tree *Node) Remove(key Interface) (*Node, *Node, bool) {
	n, p := tree.findNode(key)
	if n == nil {
		return tree, nil, false
	}
	tree = tree.rmNode(n, p)
	return tree, n, true
}

// BUG(npat): The (*Node) Height method uses a naive (and very
// expensive) recursive implementation.  Should be replaced with
// something better, or even eliminated altogether.

// Height returns the height of the tree
func (tree *Node) Height() int {
	if tree == nil {
		return 0
	}
	lh := tree.l.Height()
	rh := tree.r.Height()
	if lh < rh {
		return rh + 1
	}
	return lh + 1
}

// Find searches the tree for a value (element) v that satisfies
// key.Cmp(v) == 0. If found, returns (v, true). If not returns
// (nil, false)
func (tree *Node) Find(key Interface) (Interface, bool) {
	for tree != nil {
		if k := key.Cmp(tree.V); k == 0 {
			break
		} else if k < 0 {
			tree = tree.l
		} else {
			tree = tree.r
		}
	}
	if tree == nil {
		return nil, false
	}
	return tree.V, true
}

// The Scanner type is used to recursively scan the tree. Scanning is
// implemented by launching a go-routine that walks the tree and emits
// node-values on a channel.
type Scanner struct {
	ch   <-chan Interface
	quit chan<- int
}

func scan(root *Node, reverse bool, low, hi Interface,
	ch chan<- Interface, quit <-chan int, top bool) {
	var clow, chi int
	var pre, post *Node
	var left, emit, right, dopre, dopost bool

	if low != nil {
		clow = low.Cmp(root.V)
	} else {
		clow = -1
	}
	if hi != nil {
		chi = hi.Cmp(root.V)
	} else {
		chi = 1
	}
	if clow > 0 {
		left, emit, right = false, false, true
	} else if clow == 0 {
		left, emit, right = false, true, true
	} else if chi >= 0 {
		left, emit, right = true, true, true
	} else {
		left, emit, right = true, false, false
	}
	if reverse {
		pre, post, dopre, dopost = root.r, root.l, right, left
	} else {
		pre, post, dopre, dopost = root.l, root.r, left, right
	}
	if dopre && pre != nil {
		scan(pre, reverse, low, hi, ch, quit, false)
	}
	if emit {
		select {
		case <-quit:
			if top {
				close(ch)
			}
			return
		case ch <- root.V:
		}
	}
	if dopost && post != nil {
		scan(post, reverse, low, hi, ch, quit, false)
	}
	if top {
		close(ch)
	}
}

// NewScanner creates a new tree-scanner, initializes it, and spawns
// the respective scanning go-routine. The scanner walks the tree in
// ascending element (value) order if "reverse" is false (or in
// descending value order if "reverse" is true), emiting the values
// (v) of nodes for which: low.Cmp(v) <= 0 && hi.Cmp(v) >= 0.
func (tree *Node) NewScanner(reverse bool, low, hi Interface) Scanner {
	ch := make(chan Interface)
	quit := make(chan int)
	if tree != nil {
		go scan(tree, reverse, low, hi, ch, quit, true)
	} else {
		close(ch)
	}
	return Scanner{ch, quit}
}

// Next returns the next tree element (value). If "ok" (the second
// return value) is true, then "e" (the first return value) is the
// element. If "ok" is false, then there are no more elements.
func (sc Scanner) Next() (e Interface, ok bool) {
	v, ok := <-sc.ch
	return v, ok
}

// Stop must be called in order to stop the scanner (and free the
// resources used by it) without completing the scan. There is no need
// (but it doesn't hurt) to call Stop after the scanner returns "ok"
// == false
func (sc Scanner) Stop() {
	close(sc.quit)
}
