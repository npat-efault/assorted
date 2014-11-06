package bintree

import (
	"math/rand"
	"testing"
)

type Element int

func (e Element) Cmp(o Interface) int { return int(e - o.(Element)) }

// Helpers

func mkdata(n int) []Element {
	e := make([]Element, n)
	rand.Seed(42)
	for i := 0; i < n; i++ {
		e[i] = Element(rand.Intn(n))
	}

	return e
}

func mktree(elems []Element) *Node {
	var tree *Node
	var e Element

	for _, e = range elems {
		tree, _ = tree.Insert(New(e), false)
	}

	return tree
}

func lg2(i int) int {
	if i <= 0 {
		panic("lg2 of zero or negative!")
	}
	r := 0
	for i >>= 1; i != 0; i >>= 1 {
		r++
	}
	return r
}

func assert_sorted(tree *Node, nelems int, t *testing.T) {
	var sc Scanner
	var e0, e1 Element
	var i int

	i = 0
	sc = tree.NewScanner(false, nil, nil)
	defer sc.Stop()
	for v, ok := sc.Next(); ok; v, ok = sc.Next() {
		e0 = v.(Element)
		if i > 0 && e0 < e1 {
			t.Fatalf("el[%d] = %d < el[%d] = %d",
				i, e0, i-1, e1)
		}
		i++
		e1 = e0
	}
	if i != nelems {
		t.Fatalf("%d els inserted, %d els scanned",
			nelems, i)
	}
}

func log_tree(tree *Node, t *testing.T) {
	if tree.l != nil {
		log_tree(tree.l, t)
	}
	t.Logf("%p: %p, %v, %p\n", tree, tree.l, tree.V, tree.r)
	if tree.r != nil {
		log_tree(tree.r, t)
	}
}

// Tests

func TestInsert(t *testing.T) {
	const nelems = 100
	var tree *Node
	var ok bool

	elems := mkdata(nelems)
	tree = mktree(elems)
	tree, ok = tree.Insert(New(elems[0]), false)
	if !ok {
		t.Fatalf("cannot ins dup %d with unique == false", elems[0])
	}
	tree, ok = tree.Insert(New(elems[0]), true)
	if ok {
		t.Fatalf("ins dup %d with unique == true", elems[0])
	}
}

func TestSorting(t *testing.T) {
	const nelems = 100000
	var tree *Node

	tree = mktree(mkdata(nelems))
	assert_sorted(tree, nelems, t)
}

func TestFind(t *testing.T) {
	const nelems = 100000
	var tree *Node

	elems := mkdata(nelems)
	tree = mktree(elems)
	for _, e := range elems {
		v, ok := tree.Find(e)
		if !ok {
			t.Fatalf("elem %d, not found", e)
		}
		if v.(Element) != e {
			t.Fatalf("elem found %d != %d", v.(Element), e)
		}
	}
	_, ok := tree.Find(Element(nelems + 1))
	if ok {
		t.Fatalf("elem found %d\n", Element(nelems+1))
	}
}

func TestRemove(t *testing.T) {
	const nelems = 100000
	var tree *Node
	var ok bool

	elems := mkdata(nelems)
	tree = mktree(elems)
	tree, _, ok = tree.Remove(Element(nelems + 1))
	if ok {
		t.Fatalf("del non exist. elem %d", Element(nelems+1))
	}
	for _, e := range elems {
		tree, _, ok = tree.Remove(e)
		if !ok {
			t.Fatalf("elem %d, not deleted", e)
		}
	}
	if tree != nil {
		t.Fatalf("tree not empty: %p", tree)
	}
}

func TestBalance(t *testing.T) {
	const nelems = 10000
	var tree *Node
	var uh, bh, ebh int

	// Test with sequential elements (left-heavy)
	for i := nelems - 1; i >= 0; i-- {
		var ok bool
		tree, ok = tree.Insert(New(Element(i)), true)
		if !ok {
			t.Fatalf("Failed to insert: %d", i)
		}
	}
	if uh = tree.Height(); uh != nelems {
		t.Fatalf("SEQ: Unbal. tree height %d != %d", uh, nelems)
	} else {
		t.Logf("SEQ: Unbal. tree height %d", uh)
	}
	tree = tree.Balance()
	bh = tree.Height()
	if ebh = lg2(nelems) + 1; bh != ebh {
		t.Fatalf("SEQ: Balanced tree height %d != %d", bh, ebh)
	} else {
		t.Logf("SEQ: Balanced tree height %d", bh)
	}
	assert_sorted(tree, nelems, t)

	// Test with random elements
	tree = mktree(mkdata(nelems))
	uh = tree.Height()
	t.Logf("RAN: Unbal. tree height %d", uh)
	tree = tree.Balance()
	bh = tree.Height()
	if ebh = lg2(nelems) + 1; bh != ebh {
		t.Fatalf("RAN: Balanced tree height %d != %d", bh, ebh)
	} else {
		t.Logf("RAN: Balanced tree height %d", bh)
	}
	assert_sorted(tree, nelems, t)
}

// Benchmarks

func BenchmarkInsert(b *testing.B) {
	var tree *Node
	var elems []Element

	b.StopTimer()
	elems = mkdata(b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tree, _ = tree.Insert(New(elems[i]), false)
	}
}

func BenchmarkScan(b *testing.B) {
	var tree *Node
	var sc Scanner

	b.StopTimer()
	tree = mktree(mkdata(b.N))
	b.StartTimer()
	sc = tree.NewScanner(false, nil, nil)
	for v, ok := sc.Next(); ok; v, ok = sc.Next() {
		_ = v.(Element)
	}
	sc.Stop()
}
