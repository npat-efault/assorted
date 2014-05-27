// Auto-generated. !! DO NOT EDIT !!
// Generated: Tue May 27 22:55:54 EEST 2014

package queue

// Fifo Q made with channels. Thread safe.
type CQ chan interface{}

// New allocates and returns a new Q with space for sz elements.
func NewCQ(sz int) CQ {
	return make(CQ, sz)
}

// Empty tests if Q is empty.
func (cq CQ) Empty() bool {
	return len(cq) == 0
}

// Full tests if Q is full.
func (cq CQ) Full() bool {
	return len(cq) == cap(cq)
}

// Len returns the number of elements waiting in the Q.
func (cq CQ) Len() int {
	return len(cq)
}

// Cap returns the capacity of the Q (# of element slots).
func (cq CQ) Cap() int {
	return cap(cq)
}

// Pop removes the first element from the Q and returns it. Panics if
// Q is empty.
func (cq CQ) Pop() interface{} {
	if cq.Empty() {
		panic("CQ: pop from empty Q")
	}
	return <-cq
}

// Push adds element "e" to the tail of the Q. Panics if Q is full.
func (cq CQ) Push(e interface{}) {
	if cq.Full() {
		panic("CQ: push to full Q")
	}
	cq <- e
}
