// Auto-generated. !! DO NOT EDIT !!

package queue

import "sync"

// Fifo Q made with a slice. Thread safe.
type SQ struct {
	mu sync.Mutex
	sz uint32        /* queue size */
	m  uint32        /* queue mask (sz - 1) */
	s  uint32        /* start index */
	e  uint32        /* end index */
	b  []interface{} /* buffer */
}

// New allocates and returns a new Q with space for sz elements.
func NewSQ(sz int) *SQ {
	if uint32(sz)&(uint32(sz)-1) != 0 {
		panic("SQ: invalid Q size (not a power of 2)")
	}
	sq := &SQ{sz: uint32(sz), m: uint32(sz) - 1, s: 0, e: 0}
	sq.b = make([]interface{}, sz, sz)
	return sq
}

// Empty tests if Q is empty.
func (sq *SQ) Empty() bool {
	sq.mu.Lock()
	e := sq.s == sq.e
	sq.mu.Unlock()
	return e
}

// Full tests if Q is full.
func (sq *SQ) Full() bool {
	sq.mu.Lock()
	f := sq.e-sq.s == sq.sz
	sq.mu.Unlock()
	return f
}

// Len returns the number of elements waiting in the Q.
func (sq *SQ) Len() int {
	sq.mu.Lock()
	l := int(sq.e - sq.s)
	sq.mu.Unlock()
	return l
}

// Cap returns the capacity of the Q (# of element slots).
func (sq *SQ) Cap() int {
	sq.mu.Lock()
	c := int(sq.sz)
	sq.mu.Unlock()
	return c
}

// Peek returns the first element in the Q, without removing
// it. Panics if Q is empty.
func (sq *SQ) Peek() interface{} {
	sq.mu.Lock()
	if sq.s == sq.e {
		sq.mu.Unlock()
		panic("SQ: peek at empty Q")
	}
	e := sq.b[sq.s&sq.m]
	sq.mu.Unlock()
	return e
}

// Pop removes the first element from the Q and returns it. Panics if
// Q is empty.
func (sq *SQ) Pop() interface{} {
	sq.mu.Lock()
	if sq.s == sq.e {
		sq.mu.Unlock()
		panic("SQ: pop from empty Q")
	}
	e := sq.b[sq.s&sq.m]
	sq.s++
	sq.mu.Unlock()
	return e
}

// Push adds element "e" to the tail of the Q. Panics if Q is full.
func (sq *SQ) Push(e interface{}) {
	sq.mu.Lock()
	if sq.e-sq.s == sq.sz {
		panic("SQ: push to full Q")
	}
	sq.b[sq.e&sq.m] = e
	sq.e++
	sq.mu.Unlock()
}
