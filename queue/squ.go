// Auto-generated. !! DO NOT EDIT !!
// Generated: Tue May 27 22:55:54 EEST 2014

package queue

// Fifo Q made with a slice. *NOT* THREAD SAFE.
type SQU struct {
	sz uint32        /* queue size */
	m  uint32        /* queue mask (sz - 1) */
	s  uint32        /* start index */
	e  uint32        /* end index */
	b  []interface{} /* buffer */
}

// New allocates and returns a new Q with space for sz elements.
func NewSQU(sz int) *SQU {
	if uint32(sz)&(uint32(sz)-1) != 0 {
		panic("SQU: invalid Q size (not a power of 2)")
	}
	sq := &SQU{sz: uint32(sz), m: uint32(sz) - 1, s: 0, e: 0}
	sq.b = make([]interface{}, sz, sz)
	return sq
}

// Empty tests if Q is empty.
func (sq *SQU) Empty() bool {
	return sq.s == sq.e
}

// Full tests if Q is full.
func (sq *SQU) Full() bool {
	return sq.e-sq.s == sq.sz
}

// Len returns the number of elements waiting in the Q.
func (sq *SQU) Len() int {
	return int(sq.e - sq.s)
}

// Cap returns the capacity of the Q (# of element slots).
func (sq *SQU) Cap() int {
	return int(sq.sz)
}

// Peek returns the first element in the Q, without removing
// it. Panics if Q is empty.
func (sq *SQU) Peek() interface{} {
	if sq.Empty() {
		panic("SQU: peek at empty Q")
	}
	return sq.b[sq.s&sq.m]
}

// Pop removes the first element from the Q and returns it. Panics if
// Q is empty.
func (sq *SQU) Pop() interface{} {
	if sq.Empty() {
		panic("SQU: pop from empty Q")
	}
	i := sq.s & sq.m
	sq.s++
	return sq.b[i]
}

// Push adds element "e" to the tail of the Q. Panics if Q is full.
func (sq *SQU) Push(e interface{}) {
	if sq.Full() {
		panic("SQU: push to full Q")
	}
	sq.b[sq.e&sq.m] = e
	sq.e++
}
