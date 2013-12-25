/* 

Balance an arbitrary binary search tree (BST) using the
Day-Stout-Warren (DWS) algorithm. See:

  http://dl.acm.org/citation.cfm?id=820173
  http://www.eecs.umich.edu/~qstout/pap/CACM86.pdf
  http://penguin.ewu.edu/~trolfe/DSWpaper/

*/

package bintree

// tree_to_vine transforms the tree pointed to by "root" to a vine (a
// tree where every node's left subtree is empty) by preforming
// successive right rotations. It also counts the number of nodes in
// the tree.
//
//      (d)        (b)      a       a
//      / \        / \       \       \
//     b   e  =>  a   d  =>   b   =>  b
//    / \            / \       \       \
//   a   c          c   e      (d)      c
//                             / \       \
//                            c   e       d
//                                         \
//                                          e
//
// Returns a pointer to the new tree root, and the number of nodes in
// the tree.
func tree_to_vine(root *Node) (*Node, int) {
	var vt, rm *Node
	var sz int

	vt = nil  // vine tail
	rm = root // remainder
	sz = 0
	for rm != nil {
		if rm.l == nil {
			// Advance vt and rm
			vt = rm
			rm = rm.r
			sz++
		} else {
			// Rotate right
			//
			//       d <-rm     b <-rm
			//      / \        / \
			// t-> b   e  =>  a   d
			//    / \            / \
			//   a   c          c   e
			//
			t := rm.l
			rm.l = t.r
			t.r = rm
			if vt == nil {
				root = t
			} else {
				vt.r = t
			}
			rm = t
		}
	}
	return root, sz
}

// compress transforms ("compresses") the tree pointed to by "root" by
// performing "count" left rotations:
//
//    (a)             b              b
//      \            / \            / \
//       b     1    a  (c)    2    a   d
//        \    =>        \    =>      / \
//         c              d          c   e
//          \              \
//           d              e
//            \
//             e
//
// Rotations are performed starting from the root and accross the
// right spine of the tree. Rotations are pivoted on the 1st (root),
// 3rd, 5th, etc nodes of the spine. In the example tree shown above,
// the first rotation is pivoted on node "a", the second on node "c",
// the third would be pivoted on node "e", etc. The function does not
// check that the requested number of rotations is possible on the
// given tree; the caller must make sure of this. Returns a pointer to
// the new tree root.
func compress(root *Node, count int) *Node {
	var sc *Node // scanner
	var ch *Node // child

	sc = nil
	for i := 0; i < count; i++ {
		// Rotate left
		//
		//    * <-sc         *
		//     \              \
		//      b <-ch         d <- sc
		//     / \            / \
		//    a   d     =>   b   e
		//       / \        / \
		//      c   e      a   c
		//
		if sc == nil {
			ch = root
			root = ch.r
		} else {
			ch = sc.r
			sc.r = ch.r
		}
		sc = ch.r
		ch.r = sc.l
		sc.l = ch
	}
	return root
}

// np2 calculates the nearest power of 2 that is less than or equal to
// "i". In effect, it calculates: 2 ^ floor(log2(i))
func np2(i int) int {
	r := 1
	for r <= i {
		r <<= 1
	}
	return r >> 1
}

//  vine_to_tree transforms the vine pointed to by "root" to a
//  route-balanced tree by performing multiple compress operations
func vine_to_tree(root *Node, sz int) *Node {
	lc := sz + 1 - np2(sz+1)
	root = compress(root, lc)
	sz -= lc
	for sz > 1 {
		root = compress(root, sz>>1)
		sz >>= 1
	}
	return root
}

// Balance balances an arbitrary binary tree using the DSW
// (Day-Stout-Warren) algorithm. If n is the number of nodes in the
// tree, the algorithm performs in O(n) time, and uses constant
// additional space (irrespective of n). The result is a
// route-balanced tree exhibiting the best possible worst-case time
// and the best possible expected-case time for the strandard tree
// operations. The height of the resulting balanced tree is
// floor(log2(n)) + 1. Returns a pointer to the new tree root.
func (tree *Node) Balance() *Node {
	var sz int
	tree, sz = tree_to_vine(tree)
	tree = vine_to_tree(tree, sz)
	return tree
}
