package queue

import (
	"testing"
)

type qif interface {
	Empty() bool
	Full() bool
	Len() int
	Cap() int
	Pop() interface{}
	Push(e interface{})
}

/* Tests */

func testQ(t *testing.T, sq qif, qsz int) {
	// check emptyness
	if !sq.Empty() {
		t.Fatal("Q not empty!")
	}
	if len := sq.Len(); len != 0 {
		t.Fatal("Bad Q len:", len)
	}
	if cp := sq.Cap(); cp != qsz {
		t.Fatal("Bad Q cap:", cp)
	}

	// fill
	for i := 0; i < qsz; i++ {
		if sq.Full() {
			t.Fatal("Q full!")
		}
		sq.Push(i)
	}
	if !sq.Full() {
		t.Fatal("Q not full!")
	}

	// roll
	for i := 0; i < qsz; i++ {
		if !sq.Full() {
			t.Fatal("Q full!")
		}
		e := sq.Pop().(int)
		if e != i {
			t.Fatal("Bad elem", e, "!=", i)
		}
		if sq.Full() {
			t.Fatal("Q full!")
		}
		sq.Push(i)
		if !sq.Full() {
			t.Fatal("Q full!")
		}
	}
	if !sq.Full() {
		t.Fatal("Q not full!")
	}

	// empty
	for i := 0; i < qsz; i++ {
		if sq.Empty() {
			t.Fatal("Q empty")
		}
		e := sq.Pop().(int)
		if e != i {
			t.Fatal("Bad elem", e, "!=", i)
		}
	}
	if !sq.Empty() {
		t.Fatal("Q not empty")
	}
}

func TestSQU(t *testing.T) {
	sq := NewSQU(128)
	testQ(t, sq, 128)
}

func TestSQ(t *testing.T) {
	sq := NewSQ(128)
	testQ(t, sq, 128)
}

func TestCQ(t *testing.T) {
	cq := NewCQ(128)
	testQ(t, cq, 128)
}

/* Benchmarks */

type eT struct {
	f1, f2, f3, f4 int
}

/* bench with structs */

func benchQ(b *testing.B, sq qif, qsz int) {
	var e eT
	for i := 0; i < qsz; i++ {
		e.f1 = i
		sq.Push(e)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !sq.Empty() {
			e = sq.Pop().(eT)
		}
		if !sq.Full() {
			sq.Push(e)
		}
	}
}

func BenchmarkSQU_S(b *testing.B) {
	sq := NewSQU(128)
	benchQ(b, sq, 128)
}

func BenchmarkSQ_S(b *testing.B) {
	sq := NewSQ(128)
	benchQ(b, sq, 128)
}

func BenchmarkCQ_S(b *testing.B) {
	cq := NewCQ(128)
	benchQ(b, cq, 128)
}

/* bench with pointers */

func benchPQ(b *testing.B, sq qif, qsz int) {
	var e *eT
	for i := 0; i < qsz; i++ {
		e = &eT{f1: i}
		sq.Push(e)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !sq.Empty() {
			e = sq.Pop().(*eT)
		}
		if !sq.Full() {
			sq.Push(e)
		}
	}
}

func BenchmarkSQU_P(b *testing.B) {
	sq := NewSQU(128)
	benchPQ(b, sq, 128)
}

func BenchmarkSQ_P(b *testing.B) {
	sq := NewSQ(128)
	benchPQ(b, sq, 128)
}

func BenchmarkCQ_P(b *testing.B) {
	cq := NewCQ(128)
	benchPQ(b, cq, 128)
}
