package bintree_test

import (
	"fmt"
	"github.com/npat-efault/gohacks/bintree"
)

// Type for the element keys
type Key int

// Type for the tree elements (key + value)
type Element struct {
	k Key
	v string
}

// Both, keys and elements must implement "bintree.Interface".

// Keys must compare to elements.
func (k Key) Cmp(e bintree.Interface) int {
	return int(k - e.(*Element).k)
}

// Elements must compare to each other
func (e *Element) Cmp(o bintree.Interface) int {
	return e.k.Cmp(o)
}

// Alternativelly, you can define only an Element type, and use dummy
// elements as keys

func Example() {
	// nil pointer is empty tree
	var tree *bintree.Node
	var sc bintree.Scanner
	var e *Element
	els := []*Element{
		{73, "foo"}, {32, "bar"}, {33, "baz"}, {10, "qux"},
		{42, "quux"}, {5, "corge"}, {8, "grault"}, {4, "garply"},
	}

	for _, e = range els {
		// create a new node with element e and insert it
		// in the tree
		var ok bool
		tree, ok = tree.Insert(bintree.New(e), true)
		if !ok {
			fmt.Println("Elem %d already in tree!", e.k)
		}
	}

	// Remove (and delete) node with key == 10
	tree, _, _ = tree.Remove(Key(10))

	// Scan the tree for elements (e): e <= 70
	sc = tree.NewScanner(false, nil, Key(70))
	for v, ok := sc.Next(); ok; v, ok = sc.Next() {
		e = v.(*Element)
		fmt.Println(*e)
	}
	sc.Stop()
	// Output:
	// {4 garply}
	// {5 corge}
	// {8 grault}
	// {32 bar}
	// {33 baz}
	// {42 quux}
}
